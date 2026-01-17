package lock

import (
	"testing"
	"time"
)

func TestEncode_AcquireLockCommand(t *testing.T) {
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

			data, err := Encode(cmd)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if len(data) == 0 {
				t.Error("Encode() returned empty data")
			}
		})
	}
}

func TestEncode_ReleaseLockCommand(t *testing.T) {
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
			name:       "with zero fence token",
			key:        "zero-token-key",
			clientID:   "client-456",
			fenceToken: 0,
		},
		{
			name:       "with large fence token",
			key:        "large-token-key",
			clientID:   "client-789",
			fenceToken: 999999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewReleaseLockCommand(tt.key, tt.clientID, tt.fenceToken)

			data, err := Encode(cmd)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if len(data) == 0 {
				t.Error("Encode() returned empty data")
			}
		})
	}
}

func TestEncode_RenewLockCommand(t *testing.T) {
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
			name:       "with zero fence token",
			key:        "zero-token-key",
			clientID:   "client-456",
			fenceToken: 0,
			expiresIn:  time.Minute,
		},
		{
			name:       "with zero expiration",
			key:        "zero-expire-key",
			clientID:   "client-789",
			fenceToken: 5,
			expiresIn:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRenewLockCommand(tt.key, tt.clientID, tt.fenceToken, tt.expiresIn)

			data, err := Encode(cmd)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if len(data) == 0 {
				t.Error("Encode() returned empty data")
			}
		})
	}
}

func TestDecode_AcquireLockCommand(t *testing.T) {
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
			name:      "with millisecond precision",
			key:       "precision-key",
			clientID:  "client-abc",
			expiresIn: 1500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := NewAcquireLockCommand(tt.key, tt.clientID, tt.expiresIn)

			data, err := Encode(original)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			decoded, err := Decode(data)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			acquireCmd, ok := decoded.(*AcquireLockCommand)
			if !ok {
				t.Fatalf("Decode() returned %T, want *AcquireLockCommand", decoded)
			}

			if acquireCmd.Key() != tt.key {
				t.Errorf("Key() = %q, want %q", acquireCmd.Key(), tt.key)
			}
			if acquireCmd.ClientID() != tt.clientID {
				t.Errorf("ClientID() = %q, want %q", acquireCmd.ClientID(), tt.clientID)
			}
			if acquireCmd.ExpiresIn() != tt.expiresIn {
				t.Errorf("ExpiresIn() = %v, want %v", acquireCmd.ExpiresIn(), tt.expiresIn)
			}
			if acquireCmd.Type() != LockCommandTypeAcquire {
				t.Errorf("Type() = %v, want %v", acquireCmd.Type(), LockCommandTypeAcquire)
			}
		})
	}
}

func TestDecode_ReleaseLockCommand(t *testing.T) {
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
			name:       "with zero fence token",
			key:        "zero-token-key",
			clientID:   "client-456",
			fenceToken: 0,
		},
		{
			name:       "with large fence token",
			key:        "large-token-key",
			clientID:   "client-789",
			fenceToken: 999999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := NewReleaseLockCommand(tt.key, tt.clientID, tt.fenceToken)

			data, err := Encode(original)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			decoded, err := Decode(data)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			releaseCmd, ok := decoded.(*ReleaseLockCommand)
			if !ok {
				t.Fatalf("Decode() returned %T, want *ReleaseLockCommand", decoded)
			}

			if releaseCmd.Key() != tt.key {
				t.Errorf("Key() = %q, want %q", releaseCmd.Key(), tt.key)
			}
			if releaseCmd.ClientID() != tt.clientID {
				t.Errorf("ClientID() = %q, want %q", releaseCmd.ClientID(), tt.clientID)
			}
			if releaseCmd.FenceToken() != tt.fenceToken {
				t.Errorf("FenceToken() = %d, want %d", releaseCmd.FenceToken(), tt.fenceToken)
			}
			if releaseCmd.Type() != LockCommandTypeRelease {
				t.Errorf("Type() = %v, want %v", releaseCmd.Type(), LockCommandTypeRelease)
			}
		})
	}
}

func TestDecode_RenewLockCommand(t *testing.T) {
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
			name:       "with zero fence token",
			key:        "zero-token-key",
			clientID:   "client-456",
			fenceToken: 0,
			expiresIn:  time.Minute,
		},
		{
			name:       "with zero expiration",
			key:        "zero-expire-key",
			clientID:   "client-789",
			fenceToken: 5,
			expiresIn:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := NewRenewLockCommand(tt.key, tt.clientID, tt.fenceToken, tt.expiresIn)

			data, err := Encode(original)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			decoded, err := Decode(data)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			renewCmd, ok := decoded.(*RenewLockCommand)
			if !ok {
				t.Fatalf("Decode() returned %T, want *RenewLockCommand", decoded)
			}

			if renewCmd.Key() != tt.key {
				t.Errorf("Key() = %q, want %q", renewCmd.Key(), tt.key)
			}
			if renewCmd.ClientID() != tt.clientID {
				t.Errorf("ClientID() = %q, want %q", renewCmd.ClientID(), tt.clientID)
			}
			if renewCmd.FenceToken() != tt.fenceToken {
				t.Errorf("FenceToken() = %d, want %d", renewCmd.FenceToken(), tt.fenceToken)
			}
			if renewCmd.ExpiresIn() != tt.expiresIn {
				t.Errorf("ExpiresIn() = %v, want %v", renewCmd.ExpiresIn(), tt.expiresIn)
			}
			if renewCmd.Type() != LockCommandTypeRenew {
				t.Errorf("Type() = %v, want %v", renewCmd.Type(), LockCommandTypeRenew)
			}
		})
	}
}

func TestDecode_InvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "invalid JSON",
			data: []byte("not json"),
		},
		{
			name: "incomplete JSON",
			data: []byte(`{"type": 1`),
		},
		{
			name: "null",
			data: []byte("null"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode(tt.data)
			if err == nil {
				t.Error("Decode() expected error for invalid JSON")
			}
		})
	}
}

func TestDecode_UnknownCommandType(t *testing.T) {
	// Valid JSON but unknown command type
	data := []byte(`{"type": 999, "command": {}}`)
	_, err := Decode(data)
	if err == nil {
		t.Error("Decode() expected error for unknown command type")
	}
}

func TestDecode_InvalidCommandPayload(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "acquire with invalid command payload",
			data: []byte(`{"type": 1, "command": "not an object"}`),
		},
		{
			name: "release with invalid command payload",
			data: []byte(`{"type": 2, "command": "not an object"}`),
		},
		{
			name: "renew with invalid command payload",
			data: []byte(`{"type": 3, "command": "not an object"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode(tt.data)
			if err == nil {
				t.Error("Decode() expected error for invalid command payload")
			}
		})
	}
}

func TestEncode_NilCommand(t *testing.T) {
	// Test that encoding nil returns an error
	_, err := Encode(nil)
	if err == nil {
		t.Error("Encode(nil) expected error")
	}
}

// mockUnknownCommand is a test helper to verify encoding handles unknown command types
type mockUnknownCommand struct{}

func (m *mockUnknownCommand) Type() LockCommandType {
	return LockCommandType(999)
}

func (m *mockUnknownCommand) Key() string {
	return "unknown"
}

func TestEncode_UnknownCommandType(t *testing.T) {
	cmd := &mockUnknownCommand{}
	_, err := Encode(cmd)
	if err == nil {
		t.Error("Encode() expected error for unknown command type")
	}
}

func TestRoundTrip_AllCommands(t *testing.T) {
	t.Run("acquire command round trip", func(t *testing.T) {
		original := NewAcquireLockCommand("my-lock", "client-xyz", 45*time.Second)

		encoded, err := Encode(original)
		if err != nil {
			t.Fatalf("Encode() error = %v", err)
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Decode() error = %v", err)
		}

		result, ok := decoded.(*AcquireLockCommand)
		if !ok {
			t.Fatalf("decoded type = %T, want *AcquireLockCommand", decoded)
		}

		if result.Key() != original.Key() {
			t.Errorf("Key() = %q, want %q", result.Key(), original.Key())
		}
		if result.ClientID() != original.ClientID() {
			t.Errorf("ClientID() = %q, want %q", result.ClientID(), original.ClientID())
		}
		if result.ExpiresIn() != original.ExpiresIn() {
			t.Errorf("ExpiresIn() = %v, want %v", result.ExpiresIn(), original.ExpiresIn())
		}
	})

	t.Run("release command round trip", func(t *testing.T) {
		original := NewReleaseLockCommand("my-lock", "client-xyz", 42)

		encoded, err := Encode(original)
		if err != nil {
			t.Fatalf("Encode() error = %v", err)
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Decode() error = %v", err)
		}

		result, ok := decoded.(*ReleaseLockCommand)
		if !ok {
			t.Fatalf("decoded type = %T, want *ReleaseLockCommand", decoded)
		}

		if result.Key() != original.Key() {
			t.Errorf("Key() = %q, want %q", result.Key(), original.Key())
		}
		if result.ClientID() != original.ClientID() {
			t.Errorf("ClientID() = %q, want %q", result.ClientID(), original.ClientID())
		}
		if result.FenceToken() != original.FenceToken() {
			t.Errorf("FenceToken() = %d, want %d", result.FenceToken(), original.FenceToken())
		}
	})

	t.Run("renew command round trip", func(t *testing.T) {
		original := NewRenewLockCommand("my-lock", "client-xyz", 42, 60*time.Second)

		encoded, err := Encode(original)
		if err != nil {
			t.Fatalf("Encode() error = %v", err)
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Decode() error = %v", err)
		}

		result, ok := decoded.(*RenewLockCommand)
		if !ok {
			t.Fatalf("decoded type = %T, want *RenewLockCommand", decoded)
		}

		if result.Key() != original.Key() {
			t.Errorf("Key() = %q, want %q", result.Key(), original.Key())
		}
		if result.ClientID() != original.ClientID() {
			t.Errorf("ClientID() = %q, want %q", result.ClientID(), original.ClientID())
		}
		if result.FenceToken() != original.FenceToken() {
			t.Errorf("FenceToken() = %d, want %d", result.FenceToken(), original.FenceToken())
		}
		if result.ExpiresIn() != original.ExpiresIn() {
			t.Errorf("ExpiresIn() = %v, want %v", result.ExpiresIn(), original.ExpiresIn())
		}
	})
}

func TestEncode_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		clientID string
	}{
		{
			name:     "unicode characters",
			key:      "lock-key",
			clientID: "client-id",
		},
		{
			name:     "special JSON characters",
			key:      `key"with"quotes`,
			clientID: `client\with\backslash`,
		},
		{
			name:     "newlines and tabs",
			key:      "key\nwith\nnewlines",
			clientID: "client\twith\ttabs",
		},
		{
			name:     "null bytes",
			key:      "key\x00with\x00nulls",
			clientID: "client\x00id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := NewAcquireLockCommand(tt.key, tt.clientID, time.Second)

			encoded, err := Encode(original)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			decoded, err := Decode(encoded)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			result, ok := decoded.(*AcquireLockCommand)
			if !ok {
				t.Fatalf("decoded type = %T, want *AcquireLockCommand", decoded)
			}

			if result.Key() != tt.key {
				t.Errorf("Key() = %q, want %q", result.Key(), tt.key)
			}
			if result.ClientID() != tt.clientID {
				t.Errorf("ClientID() = %q, want %q", result.ClientID(), tt.clientID)
			}
		})
	}
}

func TestEncode_DurationPrecision(t *testing.T) {
	// The serialization uses milliseconds, so test precision
	tests := []struct {
		name      string
		expiresIn time.Duration
		expected  time.Duration
	}{
		{
			name:      "exact milliseconds",
			expiresIn: 1000 * time.Millisecond,
			expected:  1000 * time.Millisecond,
		},
		{
			name:      "sub-millisecond truncation",
			expiresIn: 1500*time.Millisecond + 500*time.Microsecond,
			expected:  1500 * time.Millisecond, // Microseconds are truncated
		},
		{
			name:      "nanoseconds truncated",
			expiresIn: time.Millisecond + 999*time.Nanosecond,
			expected:  time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := NewAcquireLockCommand("key", "client", tt.expiresIn)

			encoded, err := Encode(original)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			decoded, err := Decode(encoded)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			result := decoded.(*AcquireLockCommand)
			if result.ExpiresIn() != tt.expected {
				t.Errorf("ExpiresIn() = %v, want %v", result.ExpiresIn(), tt.expected)
			}
		})
	}
}

func TestDecode_ZeroCommandType(t *testing.T) {
	// Type 0 is not a valid command type
	data := []byte(`{"type": 0, "command": {}}`)
	_, err := Decode(data)
	if err == nil {
		t.Error("Decode() expected error for zero command type")
	}
}

func TestDecode_NegativeCommandType(t *testing.T) {
	data := []byte(`{"type": -1, "command": {}}`)
	_, err := Decode(data)
	if err == nil {
		t.Error("Decode() expected error for negative command type")
	}
}

// Benchmark tests for serialization performance
func BenchmarkEncode_AcquireLockCommand(b *testing.B) {
	cmd := NewAcquireLockCommand("benchmark-key", "benchmark-client", 30*time.Second)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Encode(cmd)
	}
}

func BenchmarkEncode_ReleaseLockCommand(b *testing.B) {
	cmd := NewReleaseLockCommand("benchmark-key", "benchmark-client", 42)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Encode(cmd)
	}
}

func BenchmarkEncode_RenewLockCommand(b *testing.B) {
	cmd := NewRenewLockCommand("benchmark-key", "benchmark-client", 42, 30*time.Second)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Encode(cmd)
	}
}

func BenchmarkDecode_AcquireLockCommand(b *testing.B) {
	cmd := NewAcquireLockCommand("benchmark-key", "benchmark-client", 30*time.Second)
	data, _ := Encode(cmd)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Decode(data)
	}
}

func BenchmarkDecode_ReleaseLockCommand(b *testing.B) {
	cmd := NewReleaseLockCommand("benchmark-key", "benchmark-client", 42)
	data, _ := Encode(cmd)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Decode(data)
	}
}

func BenchmarkDecode_RenewLockCommand(b *testing.B) {
	cmd := NewRenewLockCommand("benchmark-key", "benchmark-client", 42, 30*time.Second)
	data, _ := Encode(cmd)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Decode(data)
	}
}

func BenchmarkRoundTrip_AcquireLockCommand(b *testing.B) {
	cmd := NewAcquireLockCommand("benchmark-key", "benchmark-client", 30*time.Second)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, _ := Encode(cmd)
		_, _ = Decode(data)
	}
}
