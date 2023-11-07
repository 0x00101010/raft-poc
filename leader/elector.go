package leader

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"go.uber.org/atomic"
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
}

func NewElector(ctx context.Context, cfg *Config) (*Elector, error) {
	e := &Elector{
		config:   cfg,
		leader:   atomic.NewBool(false),
		leaderCh: make(chan bool, 1),
	}

	if err := e.makeRaft(ctx); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Elector) Run() {
	// TODO: implement
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
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
