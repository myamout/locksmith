package lock

type ReleaseLockCommand struct {
	key        string
	clientID   string
	fenceToken int
}

func NewReleaseLockCommand(key, clientID string, fenceToken int) *ReleaseLockCommand {
	return &ReleaseLockCommand{
		key:        key,
		clientID:   clientID,
		fenceToken: fenceToken,
	}
}

func (r *ReleaseLockCommand) Type() LockCommandType {
	return LockCommandTypeRelease
}

func (r *ReleaseLockCommand) Key() string {
	return r.key
}

func (r *ReleaseLockCommand) ClientID() string {
	return r.clientID
}

func (r *ReleaseLockCommand) FenceToken() int {
	return r.fenceToken
}
