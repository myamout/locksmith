package node

import "time"

type NodeConfig struct {
	ID                 string        `json:"id"`
	Addr               string        `json:"addr"`
	DataDir            string        `json:"dataDir"`
	Leader             bool          `json:"leader"`
	HeartbeatTimeout   time.Duration `json:"heartbeatTimeout"`
	ElectionTimeout    time.Duration `json:"electionTimeout"`
	LeaderLeaseTimeout time.Duration `json:"leaderLeaseTimeout"`
	CommitTimeout      time.Duration `json:"commitTimeout"`
	SnapshotInterval   time.Duration `json:"snapshotInterval"`
	SnapshotThreshold  int           `json:"snapshotThreshold"`
}

func DefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		HeartbeatTimeout:   time.Second * 1,
		ElectionTimeout:    time.Second * 1,
		LeaderLeaseTimeout: time.Second * 1,
		CommitTimeout:      time.Second * 1,
		SnapshotInterval:   time.Second * 180,
		SnapshotThreshold:  8192,
	}
}
