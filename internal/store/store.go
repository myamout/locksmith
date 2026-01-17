package store

import (
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/myamout/locksmith/internal/lock"
)

type LockSessionStore struct {
	locks      map[string]*lock.LockEntry
	fenceToken int
}

type lockStoreSnapshot struct {
	Locks      map[string]*lock.LockEntry `json:"locks"`
	FenceToken int                        `json:"fenceToken"`
}

func NewLockSessionStore() *LockSessionStore {
	return &LockSessionStore{
		locks: make(map[string]*lock.LockEntry),
	}
}

func (ls *LockSessionStore) Encode() ([]byte, error) {
	snapshot := lockStoreSnapshot{
		Locks:      make(map[string]*lock.LockEntry),
		FenceToken: ls.fenceToken,
	}

	for k, v := range ls.locks {
		snapshot.Locks[k] = v
	}

	return sonic.Marshal(snapshot)
}

func (ls *LockSessionStore) Decode(data []byte) error {
	var snapshot lockStoreSnapshot
	if err := sonic.Unmarshal(data, &snapshot); err != nil {
		return err
	}

	ls.fenceToken = snapshot.FenceToken
	ls.locks = make(map[string]*lock.LockEntry)

	for k, v := range snapshot.Locks {
		ls.locks[k] = v
	}

	return nil
}

func (ls *LockSessionStore) acquire(cmd *lock.AcquireLockCommand) *lock.AcquireLockCommandResponse {
	l, ok := ls.locks[cmd.Key()]
	if ok && time.Now().Before(l.ExpiresAt) {
		if l.ClientID == cmd.ClientID() {
			return &lock.AcquireLockCommandResponse{
				FenceToken: l.FenceToken,
				ExpiresAt:  l.ExpiresAt,
			}
		}
		err := fmt.Errorf("lock already held by client id %s", l.ClientID)
		return &lock.AcquireLockCommandResponse{
			LockCommandResponse: lock.LockCommandResponse{
				Error: &err,
			},
			HeldBy:    l.ClientID,
			ExpiresAt: l.ExpiresAt,
		}
	}

	ls.fenceToken++
	sessionLock := &lock.LockEntry{
		Key:        cmd.Key(),
		ClientID:   cmd.ClientID(),
		FenceToken: ls.fenceToken,
		ExpiresAt:  cmd.ExpiresAt(),
		CreatedAt:  cmd.AppliedAt(),
	}
	ls.locks[cmd.Key()] = sessionLock

	return &lock.AcquireLockCommandResponse{
		FenceToken: sessionLock.FenceToken,
		ExpiresAt:  sessionLock.ExpiresAt,
	}
}

func (ls *LockSessionStore) release(cmd *lock.ReleaseLockCommand) *lock.LockCommandResponse {
	l, ok := ls.locks[cmd.Key()]

	if !ok {
		err := errors.New("lock already released")
		return &lock.LockCommandResponse{
			Error: &err,
		}
	}

	if l.ClientID != cmd.ClientID() {
		err := fmt.Errorf("lock client id %s does not match requesting client id %s", l.ClientID, cmd.ClientID())
		return &lock.LockCommandResponse{
			Error: &err,
		}
	}

	if l.FenceToken != cmd.FenceToken() {
		err := fmt.Errorf("lock fence token %d does not matching requesting fence token %d", l.FenceToken, cmd.FenceToken())
		return &lock.LockCommandResponse{
			Error: &err,
		}
	}

	delete(ls.locks, cmd.Key())

	return &lock.LockCommandResponse{}
}

func (ls *LockSessionStore) renew(cmd *lock.RenewLockCommand) *lock.RenewLockCommandResponse {
	l, ok := ls.locks[cmd.Key()]
	if !ok {
		err := fmt.Errorf("cannot renew lock %s as it does not exist", cmd.Key())
		return &lock.RenewLockCommandResponse{
			LockCommandResponse: lock.LockCommandResponse{
				Error: &err,
			},
		}
	}

	if time.Now().After(l.ExpiresAt) {
		err := fmt.Errorf("lock %s is expired", cmd.Key())
		return &lock.RenewLockCommandResponse{
			LockCommandResponse: lock.LockCommandResponse{
				Error: &err,
			},
		}
	}

	if l.ClientID != cmd.ClientID() {
		err := fmt.Errorf("cannot renew lock %s, invalid owner", cmd.Key())
		return &lock.RenewLockCommandResponse{
			LockCommandResponse: lock.LockCommandResponse{
				Error: &err,
			},
		}
	}

	if l.FenceToken != cmd.FenceToken() {
		err := fmt.Errorf("cannot renew lock %s due to invalid fence token", cmd.Key())
		return &lock.RenewLockCommandResponse{
			LockCommandResponse: lock.LockCommandResponse{
				Error: &err,
			},
		}
	}

	l.ExpiresAt = cmd.ExpiresAt()
	return &lock.RenewLockCommandResponse{
		ExpiresAt: l.ExpiresAt,
	}
}

func (ls *LockSessionStore) Action(cmd lock.LockCommand) any {
	switch action := cmd.(type) {
	case *lock.AcquireLockCommand:
		return ls.acquire(action)
	case *lock.ReleaseLockCommand:
		return ls.release(action)
	case *lock.RenewLockCommand:
		return ls.renew(action)
	default:
		err := errors.New("invalid lock command provided")
		return &lock.LockCommandResponse{
			Error: &err,
		}
	}
}
