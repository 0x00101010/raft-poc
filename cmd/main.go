package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/base-org/leader-election/leader"
	"github.com/base-org/leader-election/leader/flags"
	"github.com/hashicorp/raft"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1"
	app.Flags = flags.Flags
	app.Name = "leader-elector"
	app.Usage = "Sequencer Leader Election Service"
	app.Description = "A service that uses Raft to elect a leader for a sequencer"
	app.Action = LeaderElectorMain

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func LeaderElectorMain(ctx *cli.Context) error {
	cfg, err := ReadConfig(ctx)
	if err != nil {
		return err
	}

	le, err := leader.NewElector(context.Background(), cfg)
	if err != nil {
		return err
	}

	le.Run()
	return nil
}

func ReadConfig(ctx *cli.Context) (*leader.Config, error) {
	if err := flags.CheckRequired(ctx); err != nil {
		return nil, err
	}

	rc := raft.DefaultConfig()
	rc.LocalID = raft.ServerID(ctx.String(flags.ServerID.Name))

	cfg := &leader.Config{
		RaftConfig:    rc,
		ServerAddr:    ctx.String(flags.ServerAddr.Name),
		StorageDir:    filepath.Join(ctx.String(flags.StorageDir.Name), ctx.String(flags.ServerID.Name)),
		SnapshotLimit: ctx.Int(flags.SnapshotLimit.Name),
		Bootstrap:     ctx.Bool(flags.Bootstrap.Name),
	}

	return cfg, nil
}
