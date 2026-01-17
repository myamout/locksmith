package lock

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/bytedance/sonic"
)

type lockCommandJson struct {
	Type    LockCommandType `json:"type"`
	Command json.RawMessage `json:"command"`
}

type acquireLockCommandJson struct {
	Key       string `json:"key"`
	ClientID  string `json:"clientID"`
	ExpiresIn uint64 `json:"expiresIn"`
}

type releaseLockCommandJson struct {
	Key        string `json:"key"`
	ClientID   string `json:"clientID"`
	FenceToken int    `json:"fenceToken"`
}

type renewLockCommandJson struct {
	Key        string `json:"key"`
	ClientID   string `json:"clientID"`
	FenceToken int    `json:"fenceToken"`
	ExpiresIn  uint64 `json:"expiresIn"`
}

func Encode(cmd LockCommand) ([]byte, error) {
	var data []byte
	var err error

	switch c := cmd.(type) {
	case *AcquireLockCommand:
		data, err = sonic.Marshal(acquireLockCommandJson{
			Key:       c.Key(),
			ClientID:  c.ClientID(),
			ExpiresIn: uint64(c.ExpiresIn().Milliseconds()),
		})
	case *ReleaseLockCommand:
		data, err = sonic.Marshal(releaseLockCommandJson{
			Key:        c.Key(),
			ClientID:   c.ClientID(),
			FenceToken: c.FenceToken(),
		})
	case *RenewLockCommand:
		data, err = sonic.Marshal(renewLockCommandJson{
			Key:        c.Key(),
			ClientID:   c.ClientID(),
			FenceToken: c.FenceToken(),
			ExpiresIn:  uint64(c.ExpiresIn().Milliseconds()),
		})
	default:
		return nil, errors.New("failed to encode command")
	}

	if err != nil {
		return nil, errors.New("failed to encode command")
	}

	return sonic.Marshal(lockCommandJson{
		Type:    cmd.Type(),
		Command: data,
	})
}

func Decode(data []byte) (LockCommand, error) {
	var cmd lockCommandJson
	if err := sonic.Unmarshal(data, &cmd); err != nil {
		return nil, errors.New("failed to decode command")
	}

	switch cmd.Type {
	case LockCommandTypeAcquire:
		var a acquireLockCommandJson
		if err := sonic.Unmarshal(cmd.Command, &a); err != nil {
			return nil, errors.New("failed to decode command")
		}
		return NewAcquireLockCommand(a.Key, a.ClientID, time.Duration(a.ExpiresIn)*time.Millisecond), nil
	case LockCommandTypeRelease:
		var r releaseLockCommandJson
		if err := sonic.Unmarshal(cmd.Command, &r); err != nil {
			return nil, errors.New("failed to decode command")
		}
		return NewReleaseLockCommand(r.Key, r.ClientID, r.FenceToken), nil
	case LockCommandTypeRenew:
		var r renewLockCommandJson
		if err := sonic.Unmarshal(cmd.Command, &r); err != nil {
			return nil, errors.New("failed to decode command")
		}
		return NewRenewLockCommand(r.Key, r.ClientID, r.FenceToken, time.Duration(r.ExpiresIn)*time.Millisecond), nil
	default:
		return nil, errors.New("unable to decode unknown command")
	}
}
