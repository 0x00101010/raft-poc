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

	OpNodeAddr = &cli.StringFlag{
		Name:   "op-node-addr",
		Usage:  "The addr to bind to for the op-node service",
		EnvVar: "OP_NODE_ADDR",
	}

	OpBatcherAddr = &cli.StringFlag{
		Name:   "op-batcher-addr",
		Usage:  "The port to bind to for the op-batcher service",
		EnvVar: "OP_BATCHER_ADDR",
	}

	OpGethAddr = &cli.StringFlag{
		Name:   "op-geth-addr",
		Usage:  "The port to bind to for the op-geth service",
		EnvVar: "OP_GETH_ADDR",
	}

	// ============================
	// Test related flags
	// ============================
	Test = &cli.BoolFlag{
		Name:   "test",
		Usage:  "Run in test mode",
		EnvVar: "TEST",
	}

	HealthCheckPath = &cli.StringFlag{
		Name:   "health-check-path",
		Usage:  "The file path to use for health checks",
		EnvVar: "HEALTH_CHECK_PATH",
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
	OpNodeAddr,
	OpBatcherAddr,
}

var testFlags = []cli.Flag{
	Test,
	HealthCheckPath,
}

// Flags is the collection of flags used by the binary.
var Flags []cli.Flag

func init() {
	Flags = append(requiredFlags, optionalFlags...)
	Flags = append(Flags, testFlags...)
}

func CheckRequired(ctx *cli.Context) error {
	for _, f := range requiredFlags {
		if !ctx.IsSet(f.GetName()) {
			return cli.NewExitError(fmt.Sprintf("required flag %s not set", f.GetName()), 1)
		}
	}
	return nil
}
