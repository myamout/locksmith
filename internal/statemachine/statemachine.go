package statemachine

import (
	"io"
	"sync"

	"github.com/hashicorp/raft"
	"github.com/myamout/locksmith/internal/lock"
	"github.com/myamout/locksmith/internal/store"
)

type LockStateMachine struct {
	mu    sync.RWMutex
	store *store.LockSessionStore
}

func NewLockStateMachine() *LockStateMachine {
	return &LockStateMachine{
		store: store.NewLockSessionStore(),
	}
}

func (sm *LockStateMachine) Apply(log *raft.Log) any {
	if log.Type != raft.LogCommand {
		return nil
	}

	cmd, err := lock.Decode(log.Data)
	if err != nil {
		return err
	}

	if cmd.Type() == lock.LockCommandTypeAcquire {
		c := cmd.(*lock.AcquireLockCommand)
		c.SetAppliedAt(log.AppendedAt)
	}

	if cmd.Type() == lock.LockCommandTypeRenew {
		c := cmd.(*lock.RenewLockCommand)
		c.SetAppliedAt(log.AppendedAt)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.store.Action(cmd)
}

func (sm *LockStateMachine) Snapshot() (raft.FSMSnapshot, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	data, err := sm.store.Encode()
	if err != nil {
		return nil, err
	}

	return &LockStateMachineSnapshot{
		data: data,
	}, nil
}

func (sm *LockStateMachine) Restore(reader io.ReadCloser) error {
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	store := store.NewLockSessionStore()
	if err := store.Decode(data); err != nil {
		return err
	}

	sm.store = store
	return nil
}
