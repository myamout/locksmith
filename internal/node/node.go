package node

import (
	"errors"
	"time"

	"github.com/hashicorp/raft"
	"github.com/myamout/locksmith/internal/statemachine"
)

type LocksmithNode struct {
	cfg       *NodeConfig
	raft      *raft.Raft
	sm        *statemachine.LockStateMachine
	storage   *NodeStorage
	transport *raft.NetworkTransport
}

func NewLocksmithNode(cfg *NodeConfig) (*LocksmithNode, error) {
	sm := statemachine.NewLockStateMachine()

	storage, err := NewNodeStorage(cfg.DataDir)
	if err != nil {
		return nil, err
	}

	transport, err := NewNodeTransport(cfg.Addr)
	if err != nil {
		storage.Close()
		return nil, err
	}

	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(cfg.ID)
	raftCfg.HeartbeatTimeout = cfg.HeartbeatTimeout
	raftCfg.ElectionTimeout = cfg.ElectionTimeout
	raftCfg.CommitTimeout = cfg.CommitTimeout
	raftCfg.LeaderLeaseTimeout = cfg.LeaderLeaseTimeout
	raftCfg.SnapshotInterval = cfg.SnapshotInterval
	raftCfg.SnapshotThreshold = uint64(cfg.SnapshotThreshold)

	r, err := raft.NewRaft(
		raftCfg,
		sm,
		storage.GetLogStore(),
		storage.GetStableStore(),
		storage.GetSnapshotStore(),
		transport)
	if err != nil {
		storage.Close()
		transport.Close()
		return nil, err
	}

	return &LocksmithNode{
		cfg:       cfg,
		raft:      r,
		sm:        sm,
		storage:   storage,
		transport: transport,
	}, nil
}

func (l *LocksmithNode) Bootstrap() error {
	cfg := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(l.cfg.ID),
				Address: raft.ServerAddress(l.cfg.Addr),
			},
		},
	}

	fut := l.raft.BootstrapCluster(cfg)
	if err := fut.Error(); err != nil {
		if errors.Is(err, raft.ErrCantBootstrap) {
			return nil
		}
		return err
	}

	return nil
}

func (l *LocksmithNode) AddVoter(id, addr string) error {
	if l.raft.State() != raft.Leader {
		return errors.New("non-leader node cannot add voters")
	}

	fut := l.raft.AddVoter(
		raft.ServerID(id),
		raft.ServerAddress(addr),
		0,
		time.Second*30)
	if err := fut.Error(); err != nil {
		return err
	}

	return nil
}

func (l *LocksmithNode) RemoveNode(id string) error {
	if l.raft.State() != raft.Leader {
		return errors.New("non-leader node cannot remove other nodes")
	}

	fut := l.raft.RemoveServer(
		raft.ServerID(id),
		0,
		time.Second*30)

	return fut.Error()
}

func (l *LocksmithNode) Apply(data []byte, timeout time.Duration) (any, error) {
	if l.raft.State() != raft.Leader {
		return nil, errors.New("non-leader node cannot apply changes across the cluster")
	}

	fut := l.raft.Apply(data, timeout)
	if err := fut.Error(); err != nil {
		return nil, err
	}

	return fut.Response(), nil
}

func (l *LocksmithNode) Shutdown() error {
	fut := l.raft.Shutdown()
	if err := fut.Error(); err != nil {
		return err
	}

	if err := l.storage.Close(); err != nil {
		return err
	}

	if err := l.transport.Close(); err != nil {
		return err
	}

	return nil
}
