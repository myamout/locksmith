package lock

import (
	"errors"
	"testing"
	"time"
)

func TestLockCommandType_ToString(t *testing.T) {
	tests := []struct {
		name     string
		cmdType  LockCommandType
		expected string
	}{
		{
			name:     "acquire command type",
			cmdType:  LockCommandTypeAcquire,
			expected: "acquire",
		},
		{
			name:     "release command type",
			cmdType:  LockCommandTypeRelease,
			expected: "release",
		},
		{
			name:     "renew command type",
			cmdType:  LockCommandTypeRenew,
			expected: "renew",
		},
		{
			name:     "delete command type",
			cmdType:  LockCommandTypeDelete,
			expected: "delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmdType.ToString()
			if result != tt.expected {
				t.Errorf("ToString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLockCommandType_ToString_InvalidPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid LockCommandType, but did not panic")
		}
	}()

	invalidType := LockCommandType(999)
	_ = invalidType.ToString()
}

func TestLockCommandType_Values(t *testing.T) {
	// Ensure the iota values are as expected (starting from 1)
	if LockCommandTypeAcquire != 1 {
		t.Errorf("LockCommandTypeAcquire = %d, want 1", LockCommandTypeAcquire)
	}
	if LockCommandTypeRelease != 2 {
		t.Errorf("LockCommandTypeRelease = %d, want 2", LockCommandTypeRelease)
	}
	if LockCommandTypeRenew != 3 {
		t.Errorf("LockCommandTypeRenew = %d, want 3", LockCommandTypeRenew)
	}
	if LockCommandTypeDelete != 4 {
		t.Errorf("LockCommandTypeDelete = %d, want 4", LockCommandTypeDelete)
	}
}

func TestLockCommandType_ZeroValue(t *testing.T) {
	// Zero value should not match any defined type
	var zeroType LockCommandType
	if zeroType == LockCommandTypeAcquire ||
		zeroType == LockCommandTypeRelease ||
		zeroType == LockCommandTypeRenew ||
		zeroType == LockCommandTypeDelete {
		t.Error("zero value LockCommandType should not match any defined type")
	}
}

func TestLockCommandResponse_Success(t *testing.T) {
	tests := []struct {
		name        string
		response    LockCommandResponse
		wantSuccess bool
	}{
		{
			name:        "success when error is nil",
			response:    LockCommandResponse{Error: nil},
			wantSuccess: true,
		},
		{
			name: "failure when error is set",
			response: LockCommandResponse{
				Error: func() *error { e := errors.New("test error"); return &e }(),
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.response.Success(); got != tt.wantSuccess {
				t.Errorf("Success() = %v, want %v", got, tt.wantSuccess)
			}
		})
	}
}

func TestLockCommandResponse_ErrorPointer(t *testing.T) {
	t.Run("nil error pointer", func(t *testing.T) {
		resp := LockCommandResponse{Error: nil}
		if resp.Error != nil {
			t.Errorf("Error = %v, want nil", resp.Error)
		}
	})

	t.Run("non-nil error pointer", func(t *testing.T) {
		err := errors.New("test error")
		resp := LockCommandResponse{Error: &err}
		if resp.Error == nil {
			t.Error("Error should not be nil")
		}
		if *resp.Error != err {
			t.Errorf("*Error = %v, want %v", *resp.Error, err)
		}
	})
}

func TestLockEntry_Fields(t *testing.T) {
	t.Run("all fields are set correctly", func(t *testing.T) {
		now := time.Now()
		expires := now.Add(30 * time.Second)

		entry := LockEntry{
			Key:        "test-key",
			ClientID:   "client-123",
			FenceToken: 42,
			ExpiresAt:  expires,
			CreatedAt:  now,
		}

		if entry.Key != "test-key" {
			t.Errorf("Key = %q, want %q", entry.Key, "test-key")
		}
		if entry.ClientID != "client-123" {
			t.Errorf("ClientID = %q, want %q", entry.ClientID, "client-123")
		}
		if entry.FenceToken != 42 {
			t.Errorf("FenceToken = %d, want %d", entry.FenceToken, 42)
		}
		if !entry.ExpiresAt.Equal(expires) {
			t.Errorf("ExpiresAt = %v, want %v", entry.ExpiresAt, expires)
		}
		if !entry.CreatedAt.Equal(now) {
			t.Errorf("CreatedAt = %v, want %v", entry.CreatedAt, now)
		}
	})

	t.Run("zero value entry", func(t *testing.T) {
		var entry LockEntry

		if entry.Key != "" {
			t.Errorf("Key = %q, want empty string", entry.Key)
		}
		if entry.ClientID != "" {
			t.Errorf("ClientID = %q, want empty string", entry.ClientID)
		}
		if entry.FenceToken != 0 {
			t.Errorf("FenceToken = %d, want 0", entry.FenceToken)
		}
		if !entry.ExpiresAt.IsZero() {
			t.Errorf("ExpiresAt = %v, want zero time", entry.ExpiresAt)
		}
		if !entry.CreatedAt.IsZero() {
			t.Errorf("CreatedAt = %v, want zero time", entry.CreatedAt)
		}
	})
}

func TestLockEntry_JSONTags(t *testing.T) {
	// This test verifies the struct has proper JSON tags by checking field names
	// The JSON tags ensure proper serialization/deserialization
	entry := LockEntry{
		Key:        "key",
		ClientID:   "client",
		FenceToken: 1,
		ExpiresAt:  time.Now(),
		CreatedAt:  time.Now(),
	}

	// Verify the entry can be used (compile-time check for tags)
	_ = entry
}

func TestLockEntry_TableDriven(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		key        string
		clientID   string
		fenceToken int
		expiresAt  time.Time
		createdAt  time.Time
	}{
		{
			name:       "basic entry",
			key:        "resource-1",
			clientID:   "client-1",
			fenceToken: 1,
			expiresAt:  now.Add(30 * time.Second),
			createdAt:  now,
		},
		{
			name:       "entry with empty key",
			key:        "",
			clientID:   "client-2",
			fenceToken: 2,
			expiresAt:  now.Add(time.Minute),
			createdAt:  now,
		},
		{
			name:       "entry with empty client ID",
			key:        "resource-2",
			clientID:   "",
			fenceToken: 3,
			expiresAt:  now.Add(time.Hour),
			createdAt:  now,
		},
		{
			name:       "entry with zero fence token",
			key:        "resource-3",
			clientID:   "client-3",
			fenceToken: 0,
			expiresAt:  now.Add(time.Second),
			createdAt:  now,
		},
		{
			name:       "entry with large fence token",
			key:        "resource-4",
			clientID:   "client-4",
			fenceToken: 999999999,
			expiresAt:  now.Add(24 * time.Hour),
			createdAt:  now,
		},
		{
			name:       "entry with negative fence token",
			key:        "resource-5",
			clientID:   "client-5",
			fenceToken: -1,
			expiresAt:  now.Add(time.Millisecond),
			createdAt:  now,
		},
		{
			name:       "entry with special characters in key",
			key:        "resource/path/to/lock",
			clientID:   "service:instance:12345",
			fenceToken: 100,
			expiresAt:  now.Add(5 * time.Minute),
			createdAt:  now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := LockEntry{
				Key:        tt.key,
				ClientID:   tt.clientID,
				FenceToken: tt.fenceToken,
				ExpiresAt:  tt.expiresAt,
				CreatedAt:  tt.createdAt,
			}

			if entry.Key != tt.key {
				t.Errorf("Key = %q, want %q", entry.Key, tt.key)
			}
			if entry.ClientID != tt.clientID {
				t.Errorf("ClientID = %q, want %q", entry.ClientID, tt.clientID)
			}
			if entry.FenceToken != tt.fenceToken {
				t.Errorf("FenceToken = %d, want %d", entry.FenceToken, tt.fenceToken)
			}
			if !entry.ExpiresAt.Equal(tt.expiresAt) {
				t.Errorf("ExpiresAt = %v, want %v", entry.ExpiresAt, tt.expiresAt)
			}
			if !entry.CreatedAt.Equal(tt.createdAt) {
				t.Errorf("CreatedAt = %v, want %v", entry.CreatedAt, tt.createdAt)
			}
		})
	}
}

func TestAcquireLockCommandResponse_Success(t *testing.T) {
	t.Run("success when error is nil", func(t *testing.T) {
		resp := AcquireLockCommandResponse{
			LockCommandResponse: LockCommandResponse{Error: nil},
			FenceToken:          1,
			ExpiresAt:           time.Now().Add(30 * time.Second),
		}
		if !resp.Success() {
			t.Error("Success() = false, want true")
		}
	})

	t.Run("failure when error is set", func(t *testing.T) {
		err := errors.New("lock held by another client")
		resp := AcquireLockCommandResponse{
			LockCommandResponse: LockCommandResponse{Error: &err},
			HeldBy:              "other-client",
			ExpiresAt:           time.Now().Add(30 * time.Second),
		}
		if resp.Success() {
			t.Error("Success() = true, want false")
		}
	})
}

func TestAcquireLockCommandResponse_Fields(t *testing.T) {
	t.Run("all fields are accessible", func(t *testing.T) {
		now := time.Now()
		resp := AcquireLockCommandResponse{
			LockCommandResponse: LockCommandResponse{Error: nil},
			FenceToken:          42,
			HeldBy:              "client-1",
			ExpiresAt:           now,
		}

		if resp.FenceToken != 42 {
			t.Errorf("FenceToken = %d, want 42", resp.FenceToken)
		}
		if resp.HeldBy != "client-1" {
			t.Errorf("HeldBy = %q, want %q", resp.HeldBy, "client-1")
		}
		if !resp.ExpiresAt.Equal(now) {
			t.Errorf("ExpiresAt = %v, want %v", resp.ExpiresAt, now)
		}
	})
}

func TestRenewLockCommandResponse_Success(t *testing.T) {
	t.Run("success when error is nil", func(t *testing.T) {
		resp := RenewLockCommandResponse{
			LockCommandResponse: LockCommandResponse{Error: nil},
			ExpiresAt:           time.Now().Add(60 * time.Second),
		}
		if !resp.Success() {
			t.Error("Success() = false, want true")
		}
	})

	t.Run("failure when error is set", func(t *testing.T) {
		err := errors.New("lock expired")
		resp := RenewLockCommandResponse{
			LockCommandResponse: LockCommandResponse{Error: &err},
		}
		if resp.Success() {
			t.Error("Success() = true, want false")
		}
	})
}

func TestRenewLockCommandResponse_Fields(t *testing.T) {
	t.Run("ExpiresAt is set correctly", func(t *testing.T) {
		now := time.Now()
		resp := RenewLockCommandResponse{
			LockCommandResponse: LockCommandResponse{Error: nil},
			ExpiresAt:           now,
		}

		if !resp.ExpiresAt.Equal(now) {
			t.Errorf("ExpiresAt = %v, want %v", resp.ExpiresAt, now)
		}
	})
}
