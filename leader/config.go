package leader

import "github.com/hashicorp/raft"

type Config struct {
	RaftConfig    *raft.Config
	ServerAddr    string
	Port          string
	StorageDir    string
	SnapshotLimit int
	Bootstrap     bool
}
