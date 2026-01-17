package lock

import "time"

type LockCommandType int

const (
	LockCommandTypeAcquire LockCommandType = iota + 1
	LockCommandTypeRelease
	LockCommandTypeRenew
	LockCommandTypeDelete // System forces a release
)

func (l LockCommandType) ToString() string {
	switch l {
	case LockCommandTypeAcquire:
		return "acquire"
	case LockCommandTypeRelease:
		return "release"
	case LockCommandTypeRenew:
		return "renew"
	case LockCommandTypeDelete:
		return "delete"
	default:
		panic("invalid LockCommandType")
	}
}

type LockCommand interface {
	Type() LockCommandType
	Key() string
}

type LockCommandResponse struct {
	Error *error
}

func (l *LockCommandResponse) Success() bool {
	return l.Error == nil
}

type LockEntry struct {
	Key        string    `json:"key"`
	ClientID   string    `json:"clientID"`
	FenceToken int       `json:"fenceToken"`
	ExpiresAt  time.Time `json:"expiresAt"`
	CreatedAt  time.Time `json:"createdAt"`
}
