package leader

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/Jille/raftadmin"
	"github.com/base-org/leader-election/leader/config"
	lh "github.com/base-org/leader-election/leader/health"
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
	config        *config.Config
	raft          *raft.Raft
	logStore      raft.LogStore
	stableStore   raft.StableStore
	snapshotStore raft.SnapshotStore
	transport     raft.Transport
	leader        *atomic.Bool
	leaderCh      chan bool

	// TODO: clean up later when we switch off from raft-grpc-transport lib
	tm *transport.Manager

	monitor lh.HealthMonitor
}

func NewElector(ctx context.Context, cfg *config.Config) (*Elector, error) {
	e := &Elector{
		config:   cfg,
		leader:   atomic.NewBool(false),
		leaderCh: make(chan bool, 1),
		monitor:  lh.NewSimpleHealthMonitor(cfg),
	}

	if err := e.makeRaft(ctx); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Elector) Run(ctx context.Context) {
	go e.monitorLeadership(ctx)
	go e.monitorSequencerHealth(ctx)

	s := grpc.NewServer()
	e.tm.Register(s)
	raftadmin.Register(s, e.raft)
	reflection.Register(s)
	hs := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, hs)

	sock, err := net.Listen("tcp", e.config.ServerAddr)
	if err != nil {
		panic(err)
	}
	if err = s.Serve(sock); err != nil {
		panic(err)
	}
}

func (e *Elector) makeRaft(ctx context.Context) error {
	log := e.config.RaftConfig.Logger

	if _, err := os.Stat(e.config.StorageDir); os.IsNotExist(err) {
		if err := os.Mkdir(e.config.StorageDir, 0755); err != nil {
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

func (e *Elector) monitorLeadership(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case leader := <-e.leaderCh:
			fmt.Println("leader election occured", leader)
			e.leader.Store(leader)

			if leader {
				// Start sequencer when changing to leader
			} else {
				// Stop sequencer when stepping down from leader

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

			fmt.Println("sequencer is unhealthy")
		}
	}
}
