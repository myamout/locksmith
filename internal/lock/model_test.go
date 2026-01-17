package lock

import (
	"testing"
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
