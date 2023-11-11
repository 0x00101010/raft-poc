package leader

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/Jille/raftadmin"
	"github.com/base-org/leader-election/leader/config"
	"github.com/base-org/leader-election/leader/control"
	lh "github.com/base-org/leader-election/leader/health"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Elector struct {
	log           hclog.Logger
	config        *config.Config
	raft          *raft.Raft
	logStore      raft.LogStore
	stableStore   raft.StableStore
	snapshotStore raft.SnapshotStore
	transport     raft.Transport
	leader        *atomic.Bool
	leaderCh      <-chan bool

	// TODO: clean up later when we switch off from raft-grpc-transport lib
	tm *transport.Manager

	monitor    lh.HealthMonitor
	batcherRPC control.BatcherRPC
	nodeRPC    control.NodeRPC
}

func NewElector(ctx context.Context, cfg *config.Config) (*Elector, error) {
	var batcherRPC control.BatcherRPC
	var nodeRPC control.NodeRPC
	var monitor lh.HealthMonitor
	// Run mock clients if in test mode.
	if cfg.Test {
		batcherRPC = control.NewMockBatcherRPC()
		nodeRPC = control.NewMockNodeRPC()
		monitor = lh.NewMockHealthMonitor(cfg.HealthCheckPath)
	} else {
		batcherRPC = control.NewBatcherRPC(cfg.BatcherAddr)
		nodeRPC = control.NewNodeRPC(cfg.NodeAddr)
		monitor = lh.NewSimpleHealthMonitor(cfg)
	}

	e := &Elector{
		log:        cfg.RaftConfig.Logger,
		config:     cfg,
		leader:     atomic.NewBool(false),
		monitor:    monitor,
		batcherRPC: batcherRPC,
		nodeRPC:    nodeRPC,
	}

	if err := e.makeRaft(ctx); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Elector) Run(ctx context.Context) {
	// go e.monitorLeadership(ctx)
	// go e.monitorSequencerHealth(ctx)
	go e.run(ctx)

	s := grpc.NewServer()
	e.tm.Register(s)
	raftadmin.Register(s, e.raft)
	reflection.Register(s)
	hs := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, hs)

	_, port, err := net.SplitHostPort(e.config.ServerAddr)
	if err != nil {
		log.Fatalf("failed to split host port: %v", err)
	}
	sock, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err = s.Serve(sock); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (e *Elector) makeRaft(ctx context.Context) error {
	log := e.config.RaftConfig.Logger

	if _, err := os.Stat(e.config.StorageDir); os.IsNotExist(err) {
		if err := os.MkdirAll(e.config.StorageDir, 0755); err != nil {
			return fmt.Errorf("error creating storage dir: %v", err)
		}
	}

	var err error
	e.logStore, err = boltdb.NewBoltStore(filepath.Join(e.config.StorageDir, "logs.dat"))
	if err != nil {
		return fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(e.config.StorageDir, "logs.dat"), err)
	}

	e.stableStore, err = boltdb.NewBoltStore(filepath.Join(e.config.StorageDir, "stable.dat"))
	if err != nil {
		return fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(e.config.StorageDir, "stable.dat"), err)
	}

	e.snapshotStore, err = raft.NewFileSnapshotStoreWithLogger(e.config.StorageDir, e.config.SnapshotLimit, log)
	if err != nil {
		return fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, e.config.StorageDir, err)
	}

	e.tm = transport.New(raft.ServerAddress(e.config.ServerAddr), []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	e.transport = e.tm.Transport()

	e.raft, err = raft.NewRaft(e.config.RaftConfig, nil, e.logStore, e.stableStore, e.snapshotStore, e.transport)
	if err != nil {
		return fmt.Errorf("raft.NewRaft: %v", err)
	}
	e.leaderCh = e.raft.LeaderCh()

	if e.config.Bootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       e.config.RaftConfig.LocalID,
					Address:  raft.ServerAddress(e.config.ServerAddr),
				},
			},
		}

		f := e.raft.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			return err
		}
	}

	return nil
}

func (e *Elector) run(ctx context.Context) {
	healthCh := e.monitor.Subscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case leader := <-e.leaderCh:
			// Handle leadership change
			fmt.Printf("leader election occured, leader status is now: %t\n", leader)
			e.leader.Store(leader)

			if leader {
				fmt.Printf("Starting sequencer at %s\n", e.config.ServerAddr)
				// Start sequencer when changing to leader
				current, _ := e.nodeRPC.LatestBlock()
				e.nodeRPC.StartSequencer(current)
				e.batcherRPC.StartBatcher()
			} else {
				fmt.Printf("Stopping sequencer at %s\n", e.config.ServerAddr)
				// Stop sequencer when stepping down from leader
				e.batcherRPC.StopBatcher()
				e.nodeRPC.StopSequencer()
			}
		case healthy := <-healthCh:
			fmt.Println("received health update", healthy)
			if healthy {
				continue
			}

			// TODO: make it more robust, handle error better
			fmt.Println("sequencer is unhealthy, trying to transfer leadership to another node")
			if err := e.raft.LeadershipTransfer().Error(); err != nil {
				fmt.Println("failed to transfer leadership", err)
			}
		default:
			leader := e.leader.Load()
			seqActive, err := e.nodeRPC.SequencerActive()
			if err != nil {
				fmt.Println("failed to get sequencer status", err)
			}

			if leader && !seqActive {
				fmt.Printf("Starting sequencer at %s\n", e.config.ServerAddr)
				// Start sequencer when changing to leader
				current, _ := e.nodeRPC.LatestBlock()
				e.nodeRPC.StartSequencer(current)
				e.batcherRPC.StartBatcher()
			} else if !leader && seqActive {
				fmt.Printf("Stopping sequencer at %s\n", e.config.ServerAddr)
				// Stop sequencer when stepping down from leader
				e.batcherRPC.StopBatcher()
				e.nodeRPC.StopSequencer()
			} else {
				// do nothing...
				fmt.Println("sequencer in correct state")
			}

			time.Sleep(1 * time.Second)
			// // need to add locks to avoid race condition here.
			// if e.leader.Load() {

			// } else {
			// 	fmt.Printf("Stopping sequencer at %s\n", e.config.ServerAddr)
			// 	// Stop sequencer when stepping down from leader
			// 	e.batcherRPC.StopBatcher()
			// 	e.nodeRPC.StopSequencer()
			// }
		}

	}
}

func (e *Elector) monitorLeadership(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case leader := <-e.leaderCh:
			fmt.Println("leader election occured", leader)
			e.leader.Store(leader)

			// TODO: handle error situation
			if leader {
				fmt.Printf("Starting sequencer at %s\n", e.config.ServerAddr)
				// Start sequencer when changing to leader
				current, _ := e.nodeRPC.LatestBlock()
				e.nodeRPC.StartSequencer(current)
				e.batcherRPC.StartBatcher()
			} else {
				fmt.Printf("Stopping sequencer at %s\n", e.config.ServerAddr)
				// Stop sequencer when stepping down from leader
				e.batcherRPC.StopBatcher()
				e.nodeRPC.StopSequencer()
			}
		}
	}
}

func (e *Elector) monitorSequencerHealth(ctx context.Context) {
	healthUpdate := e.monitor.Subscribe()
	fmt.Println("Started to monitor sequencer health")

	for {
		select {
		case <-ctx.Done():
			return
		case healthy := <-healthUpdate:
			fmt.Println("received health update", healthy)
			if healthy {
				continue
			}

			// TODO: make it more robust, handle error better
			fmt.Println("sequencer is unhealthy, trying to transfer leadership to another node")
			if err := e.raft.LeadershipTransfer().Error(); err != nil {
				fmt.Println("failed to transfer leadership", err)
			}
		}
	}
}
