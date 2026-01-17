package lock

import (
	"testing"
	"time"
)

func TestNewRenewLockCommand(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		clientID   string
		fenceToken int
		expiresIn  time.Duration
	}{
		{
			name:       "basic renew command",
			key:        "test-key",
			clientID:   "client-123",
			fenceToken: 1,
			expiresIn:  30 * time.Second,
		},
		{
			name:       "with empty key",
			key:        "",
			clientID:   "client-456",
			fenceToken: 100,
			expiresIn:  time.Minute,
		},
		{
			name:       "with empty client ID",
			key:        "another-key",
			clientID:   "",
			fenceToken: 50,
			expiresIn:  time.Hour,
		},
		{
			name:       "with zero fence token",
			key:        "zero-token-key",
			clientID:   "client-789",
			fenceToken: 0,
			expiresIn:  time.Second,
		},
		{
			name:       "with zero expiration",
			key:        "zero-expire-key",
			clientID:   "client-abc",
			fenceToken: 5,
			expiresIn:  0,
		},
		{
			name:       "with large fence token",
			key:        "large-token-key",
			clientID:   "client-def",
			fenceToken: 999999999,
			expiresIn:  time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRenewLockCommand(tt.key, tt.clientID, tt.fenceToken, tt.expiresIn)

			if cmd == nil {
				t.Fatal("NewRenewLockCommand returned nil")
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
			if cmd.ExpiresIn() != tt.expiresIn {
				t.Errorf("ExpiresIn() = %v, want %v", cmd.ExpiresIn(), tt.expiresIn)
			}
		})
	}
}

func TestRenewLockCommand_Type(t *testing.T) {
	cmd := NewRenewLockCommand("key", "client", 1, time.Second)
	if cmd.Type() != LockCommandTypeRenew {
		t.Errorf("Type() = %v, want %v", cmd.Type(), LockCommandTypeRenew)
	}
}

func TestRenewLockCommand_ImplementsLockCommand(t *testing.T) {
	var _ LockCommand = (*RenewLockCommand)(nil)
}

func TestRenewLockCommand_AppliedAt(t *testing.T) {
	cmd := NewRenewLockCommand("key", "client", 1, time.Second)

	// Initially appliedAt should be zero
	if !cmd.AppliedAt().IsZero() {
		t.Error("AppliedAt() should be zero initially")
	}

	// Set applied at
	now := time.Now()
	cmd.SetAppliedAt(now)

	if !cmd.AppliedAt().Equal(now) {
		t.Errorf("AppliedAt() = %v, want %v", cmd.AppliedAt(), now)
	}
}

func TestRenewLockCommand_SetAppliedAt(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
	}{
		{
			name: "current time",
			time: time.Now(),
		},
		{
			name: "past time",
			time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "future time",
			time: time.Now().Add(24 * time.Hour),
		},
		{
			name: "zero time",
			time: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRenewLockCommand("key", "client", 1, time.Second)
			cmd.SetAppliedAt(tt.time)

			if !cmd.AppliedAt().Equal(tt.time) {
				t.Errorf("AppliedAt() = %v, want %v", cmd.AppliedAt(), tt.time)
			}
		})
	}
}

func TestRenewLockCommand_ExpiresAt(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn time.Duration
		appliedAt time.Time
	}{
		{
			name:      "one second expiration",
			expiresIn: time.Second,
			appliedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:      "one minute expiration",
			expiresIn: time.Minute,
			appliedAt: time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:      "zero expiration",
			expiresIn: 0,
			appliedAt: time.Now(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRenewLockCommand("key", "client", 1, tt.expiresIn)
			cmd.SetAppliedAt(tt.appliedAt)

			expected := tt.appliedAt.Add(tt.expiresIn)
			if !cmd.ExpiresAt().Equal(expected) {
				t.Errorf("ExpiresAt() = %v, want %v", cmd.ExpiresAt(), expected)
			}
		})
	}
}

func TestRenewLockCommand_Expired(t *testing.T) {
	t.Run("not expired when appliedAt is recent", func(t *testing.T) {
		cmd := NewRenewLockCommand("key", "client", 1, time.Hour)
		cmd.SetAppliedAt(time.Now())

		if cmd.Expired() {
			t.Error("Expired() = true, want false for recently applied renewal")
		}
	})

	t.Run("expired when appliedAt is in the past", func(t *testing.T) {
		cmd := NewRenewLockCommand("key", "client", 1, time.Second)
		cmd.SetAppliedAt(time.Now().Add(-2 * time.Second))

		if !cmd.Expired() {
			t.Error("Expired() = false, want true for expired renewal")
		}
	})

	t.Run("expired with zero expiration time", func(t *testing.T) {
		cmd := NewRenewLockCommand("key", "client", 1, 0)
		cmd.SetAppliedAt(time.Now().Add(-time.Nanosecond))

		if !cmd.Expired() {
			t.Error("Expired() = false, want true for zero duration renewal")
		}
	})

	t.Run("not expired at exact boundary", func(t *testing.T) {
		cmd := NewRenewLockCommand("key", "client", 1, time.Hour)
		cmd.SetAppliedAt(time.Now())

		// Should not be expired immediately
		if cmd.Expired() {
			t.Error("renewal should not be expired immediately after creation")
		}
	})
}

func TestRenewLockCommand_Key(t *testing.T) {
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
			cmd := NewRenewLockCommand(tt.key, "client", 1, time.Second)
			if cmd.Key() != tt.key {
				t.Errorf("Key() = %q, want %q", cmd.Key(), tt.key)
			}
		})
	}
}

func TestRenewLockCommand_ClientID(t *testing.T) {
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
			cmd := NewRenewLockCommand("key", tt.clientID, 1, time.Second)
			if cmd.ClientID() != tt.clientID {
				t.Errorf("ClientID() = %q, want %q", cmd.ClientID(), tt.clientID)
			}
		})
	}
}

func TestRenewLockCommand_FenceToken(t *testing.T) {
	tests := []struct {
		name       string
		fenceToken int
	}{
		{name: "zero fence token", fenceToken: 0},
		{name: "positive fence token", fenceToken: 1},
		{name: "large fence token", fenceToken: 1000000},
		{name: "max int fence token", fenceToken: int(^uint(0) >> 1)},
		{name: "negative fence token", fenceToken: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRenewLockCommand("key", "client", tt.fenceToken, time.Second)
			if cmd.FenceToken() != tt.fenceToken {
				t.Errorf("FenceToken() = %d, want %d", cmd.FenceToken(), tt.fenceToken)
			}
		})
	}
}

func TestRenewLockCommand_ExpiresIn(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn time.Duration
	}{
		{name: "zero duration", expiresIn: 0},
		{name: "nanoseconds", expiresIn: time.Nanosecond},
		{name: "milliseconds", expiresIn: 100 * time.Millisecond},
		{name: "seconds", expiresIn: 30 * time.Second},
		{name: "minutes", expiresIn: 5 * time.Minute},
		{name: "hours", expiresIn: 24 * time.Hour},
		{name: "negative duration", expiresIn: -time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRenewLockCommand("key", "client", 1, tt.expiresIn)
			if cmd.ExpiresIn() != tt.expiresIn {
				t.Errorf("ExpiresIn() = %v, want %v", cmd.ExpiresIn(), tt.expiresIn)
			}
		})
	}
}

func TestRenewLockCommand_SetAppliedAtOverwrites(t *testing.T) {
	cmd := NewRenewLockCommand("key", "client", 1, time.Hour)

	firstTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cmd.SetAppliedAt(firstTime)

	secondTime := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	cmd.SetAppliedAt(secondTime)

	if !cmd.AppliedAt().Equal(secondTime) {
		t.Errorf("AppliedAt() = %v, want %v after second SetAppliedAt", cmd.AppliedAt(), secondTime)
	}

	// Also verify ExpiresAt reflects the new appliedAt
	expectedExpiry := secondTime.Add(time.Hour)
	if !cmd.ExpiresAt().Equal(expectedExpiry) {
		t.Errorf("ExpiresAt() = %v, want %v after updating appliedAt", cmd.ExpiresAt(), expectedExpiry)
	}
}
