package flags

import (
	"fmt"

	"github.com/urfave/cli"
)

var (
	ServerAddr = &cli.StringFlag{
		Name:   "server-addr",
		Usage:  "The address to bind to",
		EnvVar: "SERVER_ADDR",
		Value:  "127.0.0.1:50051",
	}

	ServerID = &cli.StringFlag{
		Name:   "server-id",
		Usage:  "The Raft ID of this server",
		EnvVar: "SERVER_ID",
	}

	StorageDir = &cli.StringFlag{
		Name:   "storage-dir",
		Usage:  "The directory to store raft data",
		EnvVar: "STORAGE_DIR",
	}

	SnapshotLimit = &cli.IntFlag{
		Name:   "snapshot-limit",
		Usage:  "The number of snapshots to retain on disk",
		EnvVar: "SNAPSHOT_LIMIT",
		Value:  3,
	}

	Bootstrap = &cli.BoolFlag{
		Name:   "bootstrap",
		Usage:  "Bootstrap the Raft cluster",
		EnvVar: "BOOTSTRAP",
	}
)

var requiredFlags = []cli.Flag{
	ServerAddr,
	ServerID,
	StorageDir,
}

var optionalFlags = []cli.Flag{
	SnapshotLimit,
	Bootstrap,
}

// Flags is the collection of flags used by the binary.
var Flags []cli.Flag

func init() {
	Flags = append(requiredFlags, optionalFlags...)
}

func CheckRequired(ctx *cli.Context) error {
	for _, f := range requiredFlags {
		if !ctx.IsSet(f.GetName()) {
			return cli.NewExitError(fmt.Sprintf("required flag %s not set", f.GetName()), 1)
		}
	}
	return nil
}
