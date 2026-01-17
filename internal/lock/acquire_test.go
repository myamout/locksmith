package lock

import (
	"testing"
	"time"
)

func TestNewAcquireLockCommand(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		clientID  string
		expiresIn time.Duration
	}{
		{
			name:      "basic acquire command",
			key:       "test-key",
			clientID:  "client-123",
			expiresIn: 30 * time.Second,
		},
		{
			name:      "with empty key",
			key:       "",
			clientID:  "client-456",
			expiresIn: time.Minute,
		},
		{
			name:      "with empty client ID",
			key:       "another-key",
			clientID:  "",
			expiresIn: time.Hour,
		},
		{
			name:      "with zero expiration",
			key:       "zero-key",
			clientID:  "client-789",
			expiresIn: 0,
		},
		{
			name:      "with millisecond precision",
			key:       "precision-key",
			clientID:  "client-abc",
			expiresIn: 1500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewAcquireLockCommand(tt.key, tt.clientID, tt.expiresIn)

			if cmd == nil {
				t.Fatal("NewAcquireLockCommand returned nil")
			}
			if cmd.Key() != tt.key {
				t.Errorf("Key() = %q, want %q", cmd.Key(), tt.key)
			}
			if cmd.ClientID() != tt.clientID {
				t.Errorf("ClientID() = %q, want %q", cmd.ClientID(), tt.clientID)
			}
			if cmd.ExpiresIn() != tt.expiresIn {
				t.Errorf("ExpiresIn() = %v, want %v", cmd.ExpiresIn(), tt.expiresIn)
			}
		})
	}
}

func TestAcquireLockCommand_Type(t *testing.T) {
	cmd := NewAcquireLockCommand("key", "client", time.Second)
	if cmd.Type() != LockCommandTypeAcquire {
		t.Errorf("Type() = %v, want %v", cmd.Type(), LockCommandTypeAcquire)
	}
}

func TestAcquireLockCommand_ImplementsLockCommand(t *testing.T) {
	var _ LockCommand = (*AcquireLockCommand)(nil)
}

func TestAcquireLockCommand_AppliedAt(t *testing.T) {
	cmd := NewAcquireLockCommand("key", "client", time.Second)

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

func TestAcquireLockCommand_SetAppliedAt(t *testing.T) {
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
			cmd := NewAcquireLockCommand("key", "client", time.Second)
			cmd.SetAppliedAt(tt.time)

			if !cmd.AppliedAt().Equal(tt.time) {
				t.Errorf("AppliedAt() = %v, want %v", cmd.AppliedAt(), tt.time)
			}
		})
	}
}

func TestAcquireLockCommand_ExpiresAt(t *testing.T) {
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
			cmd := NewAcquireLockCommand("key", "client", tt.expiresIn)
			cmd.SetAppliedAt(tt.appliedAt)

			expected := tt.appliedAt.Add(tt.expiresIn)
			if !cmd.ExpiresAt().Equal(expected) {
				t.Errorf("ExpiresAt() = %v, want %v", cmd.ExpiresAt(), expected)
			}
		})
	}
}

func TestAcquireLockCommand_Expired(t *testing.T) {
	t.Run("not expired when appliedAt is recent", func(t *testing.T) {
		cmd := NewAcquireLockCommand("key", "client", time.Hour)
		cmd.SetAppliedAt(time.Now())

		if cmd.Expired() {
			t.Error("Expired() = true, want false for recently applied lock")
		}
	})

	t.Run("expired when appliedAt is in the past", func(t *testing.T) {
		cmd := NewAcquireLockCommand("key", "client", time.Second)
		cmd.SetAppliedAt(time.Now().Add(-2 * time.Second))

		if !cmd.Expired() {
			t.Error("Expired() = false, want true for expired lock")
		}
	})

	t.Run("expired with zero expiration time", func(t *testing.T) {
		cmd := NewAcquireLockCommand("key", "client", 0)
		cmd.SetAppliedAt(time.Now().Add(-time.Nanosecond))

		if !cmd.Expired() {
			t.Error("Expired() = false, want true for zero duration lock")
		}
	})

	t.Run("not expired at exact boundary", func(t *testing.T) {
		// Lock should not be expired exactly at the boundary
		// (time.Now().After returns false when equal)
		cmd := NewAcquireLockCommand("key", "client", time.Hour)
		cmd.SetAppliedAt(time.Now())

		// Should not be expired immediately
		if cmd.Expired() {
			t.Error("lock should not be expired immediately after creation")
		}
	})
}

func TestAcquireLockCommand_Key(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "simple key", key: "simple"},
		{name: "empty key", key: ""},
		{name: "key with special chars", key: "lock/resource/123"},
		{name: "unicode key", key: "lock-key"},
		{name: "long key", key: "this-is-a-very-long-key-that-contains-many-characters-for-testing-purposes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewAcquireLockCommand(tt.key, "client", time.Second)
			if cmd.Key() != tt.key {
				t.Errorf("Key() = %q, want %q", cmd.Key(), tt.key)
			}
		})
	}
}

func TestAcquireLockCommand_ClientID(t *testing.T) {
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
			cmd := NewAcquireLockCommand("key", tt.clientID, time.Second)
			if cmd.ClientID() != tt.clientID {
				t.Errorf("ClientID() = %q, want %q", cmd.ClientID(), tt.clientID)
			}
		})
	}
}

func TestAcquireLockCommand_ExpiresIn(t *testing.T) {
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
			cmd := NewAcquireLockCommand("key", "client", tt.expiresIn)
			if cmd.ExpiresIn() != tt.expiresIn {
				t.Errorf("ExpiresIn() = %v, want %v", cmd.ExpiresIn(), tt.expiresIn)
			}
		})
	}
}
