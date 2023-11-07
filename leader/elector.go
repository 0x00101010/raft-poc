package leader

import (
	"context"
	"fmt"
	"net"
	"path/filepath"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/Jille/raftadmin"
	"github.com/base-org/leader-election/leader/health"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Elector struct {
	config        *Config
	raft          *raft.Raft
	logStore      raft.LogStore
	stableStore   raft.StableStore
	snapshotStore raft.SnapshotStore
	transport     raft.Transport
	leader        *atomic.Bool
	leaderCh      chan bool

	// TODO: clean up later when we switch off from raft-grpc-transport lib
	tm *transport.Manager

	monitor health.HealthMonitor
}

func NewElector(ctx context.Context, cfg *Config) (*Elector, error) {
	e := &Elector{
		config:   cfg,
		leader:   atomic.NewBool(false),
		leaderCh: make(chan bool, 1),
		monitor:  health.NewSimpleHealthMonitor(),
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

	e.tm = transport.New(raft.ServerAddress(e.config.ServerAddr), nil)
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
		case state := <-e.leaderCh:
			fmt.Println("leader election occured", state)
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
