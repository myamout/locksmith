package lock

import "time"

type RenewLockCommand struct {
	key        string
	clientID   string
	fenceToken int
	expiresIn  time.Duration
	appliedAt  time.Time
}

func NewRenewLockCommand(key, clientID string, fenceToken int, expiresIn time.Duration) *RenewLockCommand {
	return &RenewLockCommand{
		key:        key,
		clientID:   clientID,
		fenceToken: fenceToken,
		expiresIn:  expiresIn,
	}
}

func (r *RenewLockCommand) Type() LockCommandType {
	return LockCommandTypeRenew
}

func (r *RenewLockCommand) Key() string {
	return r.key
}

func (r *RenewLockCommand) ClientID() string {
	return r.clientID
}

func (r *RenewLockCommand) ExpiresIn() time.Duration {
	return r.expiresIn
}

func (r *RenewLockCommand) ExpiresAt() time.Time {
	return r.appliedAt.Add(r.ExpiresIn())
}

func (r *RenewLockCommand) Expired() bool {
	return time.Now().After(r.ExpiresAt())
}

func (r *RenewLockCommand) FenceToken() int {
	return r.fenceToken
}

func (r *RenewLockCommand) AppliedAt() time.Time {
	return r.appliedAt
}

func (r *RenewLockCommand) SetAppliedAt(t time.Time) {
	r.appliedAt = t
}

type RenewLockCommandResponse struct {
	LockCommandResponse

	ExpiresAt time.Time
}
