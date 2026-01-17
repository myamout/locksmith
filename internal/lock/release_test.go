package lock

import (
	"testing"
)

func TestNewReleaseLockCommand(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		clientID   string
		fenceToken int
	}{
		{
			name:       "basic release command",
			key:        "test-key",
			clientID:   "client-123",
			fenceToken: 1,
		},
		{
			name:       "with empty key",
			key:        "",
			clientID:   "client-456",
			fenceToken: 100,
		},
		{
			name:       "with empty client ID",
			key:        "another-key",
			clientID:   "",
			fenceToken: 50,
		},
		{
			name:       "with zero fence token",
			key:        "zero-token-key",
			clientID:   "client-789",
			fenceToken: 0,
		},
		{
			name:       "with large fence token",
			key:        "large-token-key",
			clientID:   "client-abc",
			fenceToken: 999999999,
		},
		{
			name:       "with negative fence token",
			key:        "negative-token-key",
			clientID:   "client-def",
			fenceToken: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewReleaseLockCommand(tt.key, tt.clientID, tt.fenceToken)

			if cmd == nil {
				t.Fatal("NewReleaseLockCommand returned nil")
			}
			if cmd.Key() != tt.key {
				t.Errorf("Key() = %q, want %q", cmd.Key(), tt.key)
			}
			if cmd.ClientID() != tt.clientID {
				t.Errorf("ClientID() = %q, want %q", cmd.ClientID(), tt.clientID)
			}
			if cmd.FenceToken() != tt.fenceToken {
				t.Errorf("FenceToken() = %d, want %d", cmd.FenceToken(), tt.fenceToken)
			}
		})
	}
}

func TestReleaseLockCommand_Type(t *testing.T) {
	cmd := NewReleaseLockCommand("key", "client", 1)
	if cmd.Type() != LockCommandTypeRelease {
		t.Errorf("Type() = %v, want %v", cmd.Type(), LockCommandTypeRelease)
	}
}

func TestReleaseLockCommand_ImplementsLockCommand(t *testing.T) {
	var _ LockCommand = (*ReleaseLockCommand)(nil)
}

func TestReleaseLockCommand_Key(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "simple key", key: "simple"},
		{name: "empty key", key: ""},
		{name: "key with special chars", key: "lock/resource/123"},
		{name: "key with spaces", key: "key with spaces"},
		{name: "long key", key: "this-is-a-very-long-key-that-contains-many-characters-for-testing-purposes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewReleaseLockCommand(tt.key, "client", 1)
			if cmd.Key() != tt.key {
				t.Errorf("Key() = %q, want %q", cmd.Key(), tt.key)
			}
		})
	}
}

func TestReleaseLockCommand_ClientID(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
	}{
		{name: "simple client ID", clientID: "client-1"},
		{name: "empty client ID", clientID: ""},
		{name: "UUID client ID", clientID: "550e8400-e29b-41d4-a716-446655440000"},
		{name: "complex client ID", clientID: "service:instance:12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewReleaseLockCommand("key", tt.clientID, 1)
			if cmd.ClientID() != tt.clientID {
				t.Errorf("ClientID() = %q, want %q", cmd.ClientID(), tt.clientID)
			}
		})
	}
}

func TestReleaseLockCommand_FenceToken(t *testing.T) {
	tests := []struct {
		name       string
		fenceToken int
	}{
		{name: "zero fence token", fenceToken: 0},
		{name: "positive fence token", fenceToken: 1},
		{name: "large fence token", fenceToken: 1000000},
		{name: "max int fence token", fenceToken: int(^uint(0) >> 1)},
		{name: "negative fence token", fenceToken: -1},
		{name: "min int fence token", fenceToken: -int(^uint(0)>>1) - 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewReleaseLockCommand("key", "client", tt.fenceToken)
			if cmd.FenceToken() != tt.fenceToken {
				t.Errorf("FenceToken() = %d, want %d", cmd.FenceToken(), tt.fenceToken)
			}
		})
	}
}

func TestReleaseLockCommand_Immutability(t *testing.T) {
	// ReleaseLockCommand has no setters, so once created the values should be immutable
	key := "test-key"
	clientID := "test-client"
	fenceToken := 42

	cmd := NewReleaseLockCommand(key, clientID, fenceToken)

	// Verify values remain consistent across multiple calls
	for i := 0; i < 100; i++ {
		if cmd.Key() != key {
			t.Errorf("Key() changed unexpectedly on iteration %d", i)
		}
		if cmd.ClientID() != clientID {
			t.Errorf("ClientID() changed unexpectedly on iteration %d", i)
		}
		if cmd.FenceToken() != fenceToken {
			t.Errorf("FenceToken() changed unexpectedly on iteration %d", i)
		}
		if cmd.Type() != LockCommandTypeRelease {
			t.Errorf("Type() changed unexpectedly on iteration %d", i)
		}
	}
}
