package lock

import "time"

type AcquireLockCommand struct {
	key       string
	clientID  string
	expiresIn time.Duration
	appliedAt time.Time
}

func NewAcquireLockCommand(key, clientID string, expiresIn time.Duration) *AcquireLockCommand {
	return &AcquireLockCommand{
		key:       key,
		clientID:  clientID,
		expiresIn: expiresIn,
	}
}

func (a *AcquireLockCommand) Type() LockCommandType {
	return LockCommandTypeAcquire
}

func (a *AcquireLockCommand) Key() string {
	return a.key
}

func (a *AcquireLockCommand) AppliedAt() time.Time {
	return a.appliedAt
}

func (a *AcquireLockCommand) SetAppliedAt(t time.Time) {
	a.appliedAt = t
}

func (a *AcquireLockCommand) ClientID() string {
	return a.clientID
}

func (a *AcquireLockCommand) ExpiresIn() time.Duration {
	return a.expiresIn
}

func (a *AcquireLockCommand) ExpiresAt() time.Time {
	return a.appliedAt.Add(a.ExpiresIn())
}

func (a *AcquireLockCommand) Expired() bool {
	return time.Now().After(a.ExpiresAt())
}
