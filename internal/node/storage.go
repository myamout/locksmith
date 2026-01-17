package node

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

type NodeStorage struct {
	log      raft.LogStore
	stable   raft.StableStore
	snapshot raft.SnapshotStore
	boltDB   *raftboltdb.BoltStore
}

func NewNodeStorage(dataDir string) (*NodeStorage, error) {
	if err := os.Mkdir(dataDir, 0755); err != nil {
		return nil, err
	}
	boltStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft.db"))
	if err != nil {
		return nil, err
	}
	snapshot, err := raft.NewFileSnapshotStore(filepath.Join(dataDir, "snapshots"), 3, os.Stderr)
	if err != nil {
		boltStore.Close()
		return nil, err
	}

	return &NodeStorage{
		log:      boltStore,
		stable:   boltStore,
		snapshot: snapshot,
		boltDB:   boltStore,
	}, nil
}

func (ns *NodeStorage) GetLogStore() raft.LogStore {
	return ns.log
}

func (ns *NodeStorage) GetStableStore() raft.StableStore {
	return ns.stable
}

func (ns *NodeStorage) GetSnapshotStore() raft.SnapshotStore {
	return ns.snapshot
}

func (ns *NodeStorage) Close() error {
	if ns.boltDB != nil {
		return ns.boltDB.Close()
	}

	return nil
}
