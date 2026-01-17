package statemachine

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/myamout/locksmith/internal/lock"
)

func TestNewLockStateMachine(t *testing.T) {
	t.Run("creates new state machine", func(t *testing.T) {
		sm := NewLockStateMachine()
		if sm == nil {
			t.Fatal("NewLockStateMachine returned nil")
		}
	})

	t.Run("state machine has initialized store", func(t *testing.T) {
		sm := NewLockStateMachine()
		if sm.store == nil {
			t.Error("store should be initialized")
		}
	})
}

// createRaftLog creates a raft.Log for testing with the given command data.
func createRaftLog(t *testing.T, cmd lock.LockCommand, appendedAt time.Time) *raft.Log {
	t.Helper()
	data, err := lock.Encode(cmd)
	if err != nil {
		t.Fatalf("failed to encode command: %v", err)
	}
	return &raft.Log{
		Type:       raft.LogCommand,
		Data:       data,
		AppendedAt: appendedAt,
	}
}

func TestLockStateMachine_Apply(t *testing.T) {
	t.Run("ignores non-command logs", func(t *testing.T) {
		sm := NewLockStateMachine()

		log := &raft.Log{
			Type: raft.LogNoop,
			Data: []byte("some data"),
		}

		result := sm.Apply(log)
		if result != nil {
			t.Errorf("expected nil result for non-command log, got %v", result)
		}
	})

	t.Run("ignores configuration logs", func(t *testing.T) {
		sm := NewLockStateMachine()

		log := &raft.Log{
			Type: raft.LogConfiguration,
			Data: []byte("config data"),
		}

		result := sm.Apply(log)
		if result != nil {
			t.Errorf("expected nil result for configuration log, got %v", result)
		}
	})

	t.Run("returns error for invalid command data", func(t *testing.T) {
		sm := NewLockStateMachine()

		log := &raft.Log{
			Type: raft.LogCommand,
			Data: []byte("invalid json"),
		}

		result := sm.Apply(log)
		if result == nil {
			t.Fatal("expected error result for invalid command data")
		}
		if _, ok := result.(error); !ok {
			t.Errorf("expected error type, got %T", result)
		}
	})

	t.Run("apply acquire command successfully", func(t *testing.T) {
		sm := NewLockStateMachine()

		cmd := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		appendedAt := time.Now()
		log := createRaftLog(t, cmd, appendedAt)

		result := sm.Apply(log)
		resp, ok := result.(*lock.AcquireLockCommandResponse)
		if !ok {
			t.Fatalf("expected *lock.AcquireLockCommandResponse, got %T", result)
		}

		if resp.Error != nil {
			t.Errorf("unexpected error: %v", *resp.Error)
		}
		if resp.FenceToken != 1 {
			t.Errorf("FenceToken = %d, want 1", resp.FenceToken)
		}
	})

	t.Run("apply sets appliedAt for acquire command", func(t *testing.T) {
		sm := NewLockStateMachine()

		cmd := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		appendedAt := time.Now()
		log := createRaftLog(t, cmd, appendedAt)

		result := sm.Apply(log)
		resp := result.(*lock.AcquireLockCommandResponse)

		if resp.Error != nil {
			t.Fatalf("unexpected error: %v", *resp.Error)
		}

		// The ExpiresAt should be based on the appendedAt time
		expectedExpiry := appendedAt.Add(30 * time.Second)
		if !resp.ExpiresAt.Equal(expectedExpiry) {
			t.Errorf("ExpiresAt = %v, want %v", resp.ExpiresAt, expectedExpiry)
		}
	})

	t.Run("apply release command successfully", func(t *testing.T) {
		sm := NewLockStateMachine()

		// First acquire a lock
		acquireCmd := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		acquireLog := createRaftLog(t, acquireCmd, time.Now())
		acquireResult := sm.Apply(acquireLog).(*lock.AcquireLockCommandResponse)

		if acquireResult.Error != nil {
			t.Fatalf("failed to acquire lock: %v", *acquireResult.Error)
		}

		// Then release it
		releaseCmd := lock.NewReleaseLockCommand("resource-1", "client-1", acquireResult.FenceToken)
		releaseLog := createRaftLog(t, releaseCmd, time.Now())
		releaseResult := sm.Apply(releaseLog)

		resp, ok := releaseResult.(*lock.LockCommandResponse)
		if !ok {
			t.Fatalf("expected *lock.LockCommandResponse, got %T", releaseResult)
		}

		if resp.Error != nil {
			t.Errorf("unexpected error: %v", *resp.Error)
		}
	})

	t.Run("apply renew command successfully", func(t *testing.T) {
		sm := NewLockStateMachine()

		// First acquire a lock
		acquireCmd := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		acquireLog := createRaftLog(t, acquireCmd, time.Now())
		acquireResult := sm.Apply(acquireLog).(*lock.AcquireLockCommandResponse)

		if acquireResult.Error != nil {
			t.Fatalf("failed to acquire lock: %v", *acquireResult.Error)
		}

		// Then renew it
		renewCmd := lock.NewRenewLockCommand("resource-1", "client-1", acquireResult.FenceToken, 60*time.Second)
		renewLog := createRaftLog(t, renewCmd, time.Now())
		renewResult := sm.Apply(renewLog)

		resp, ok := renewResult.(*lock.RenewLockCommandResponse)
		if !ok {
			t.Fatalf("expected *lock.RenewLockCommandResponse, got %T", renewResult)
		}

		if resp.Error != nil {
			t.Errorf("unexpected error: %v", *resp.Error)
		}
	})

	t.Run("apply sets appliedAt for renew command", func(t *testing.T) {
		sm := NewLockStateMachine()

		// First acquire a lock
		acquireCmd := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		acquireLog := createRaftLog(t, acquireCmd, time.Now())
		acquireResult := sm.Apply(acquireLog).(*lock.AcquireLockCommandResponse)

		if acquireResult.Error != nil {
			t.Fatalf("failed to acquire lock: %v", *acquireResult.Error)
		}

		// Then renew it
		renewAppendedAt := time.Now().Add(10 * time.Second) // Simulate later renewal
		renewCmd := lock.NewRenewLockCommand("resource-1", "client-1", acquireResult.FenceToken, 60*time.Second)
		renewLog := createRaftLog(t, renewCmd, renewAppendedAt)
		renewResult := sm.Apply(renewLog)

		resp := renewResult.(*lock.RenewLockCommandResponse)

		if resp.Error != nil {
			t.Fatalf("unexpected error: %v", *resp.Error)
		}

		// The ExpiresAt should be based on the renew appendedAt time
		expectedExpiry := renewAppendedAt.Add(60 * time.Second)
		if !resp.ExpiresAt.Equal(expectedExpiry) {
			t.Errorf("ExpiresAt = %v, want %v", resp.ExpiresAt, expectedExpiry)
		}
	})

	t.Run("apply acquire fails when lock held by another client", func(t *testing.T) {
		sm := NewLockStateMachine()

		// Client 1 acquires lock
		cmd1 := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		log1 := createRaftLog(t, cmd1, time.Now())
		sm.Apply(log1)

		// Client 2 tries to acquire same lock
		cmd2 := lock.NewAcquireLockCommand("resource-1", "client-2", 30*time.Second)
		log2 := createRaftLog(t, cmd2, time.Now())
		result := sm.Apply(log2)

		resp := result.(*lock.AcquireLockCommandResponse)
		if resp.Error == nil {
			t.Error("expected error when acquiring lock held by another client")
		}
		if resp.HeldBy != "client-1" {
			t.Errorf("HeldBy = %q, want %q", resp.HeldBy, "client-1")
		}
	})

	t.Run("apply release fails for non-existent lock", func(t *testing.T) {
		sm := NewLockStateMachine()

		cmd := lock.NewReleaseLockCommand("non-existent", "client-1", 1)
		log := createRaftLog(t, cmd, time.Now())
		result := sm.Apply(log)

		resp := result.(*lock.LockCommandResponse)
		if resp.Error == nil {
			t.Error("expected error when releasing non-existent lock")
		}
	})

	t.Run("apply renew fails for non-existent lock", func(t *testing.T) {
		sm := NewLockStateMachine()

		cmd := lock.NewRenewLockCommand("non-existent", "client-1", 1, 30*time.Second)
		log := createRaftLog(t, cmd, time.Now())
		result := sm.Apply(log)

		resp := result.(*lock.RenewLockCommandResponse)
		if resp.Error == nil {
			t.Error("expected error when renewing non-existent lock")
		}
	})
}

func TestLockStateMachine_Apply_Concurrency(t *testing.T) {
	t.Run("concurrent applies are thread-safe", func(t *testing.T) {
		sm := NewLockStateMachine()

		done := make(chan bool)
		numGoroutines := 10
		numOperations := 100

		for i := range numGoroutines {
			go func(clientNum int) {
				for j := range numOperations {
					key := "resource-" + string(rune('a'+j%26))
					clientID := "client-" + string(rune('0'+clientNum))

					// Acquire
					acquireCmd := lock.NewAcquireLockCommand(key, clientID, 100*time.Millisecond)
					acquireLog := createRaftLog(t, acquireCmd, time.Now())
					result := sm.Apply(acquireLog)

					if resp, ok := result.(*lock.AcquireLockCommandResponse); ok && resp.Error == nil {
						// Release if acquired successfully
						releaseCmd := lock.NewReleaseLockCommand(key, clientID, resp.FenceToken)
						releaseLog := createRaftLog(t, releaseCmd, time.Now())
						sm.Apply(releaseLog)
					}
				}
				done <- true
			}(i)
		}

		for range numGoroutines {
			<-done
		}
	})
}

func TestLockStateMachine_Apply_FenceTokenProgression(t *testing.T) {
	t.Run("fence tokens increment correctly", func(t *testing.T) {
		sm := NewLockStateMachine()

		var lastToken int
		for i := range 10 {
			key := "resource-" + string(rune('a'+i))
			cmd := lock.NewAcquireLockCommand(key, "client-1", 30*time.Second)
			log := createRaftLog(t, cmd, time.Now())
			result := sm.Apply(log)

			resp := result.(*lock.AcquireLockCommandResponse)
			if resp.Error != nil {
				t.Fatalf("failed to acquire lock: %v", *resp.Error)
			}

			expectedToken := i + 1
			if resp.FenceToken != expectedToken {
				t.Errorf("FenceToken = %d, want %d", resp.FenceToken, expectedToken)
			}
			if resp.FenceToken <= lastToken {
				t.Errorf("FenceToken %d is not greater than previous %d", resp.FenceToken, lastToken)
			}
			lastToken = resp.FenceToken
		}
	})
}

func TestLockStateMachine_Snapshot(t *testing.T) {
	t.Run("creates snapshot from empty state", func(t *testing.T) {
		sm := NewLockStateMachine()

		snapshot, err := sm.Snapshot()
		if err != nil {
			t.Fatalf("Snapshot() error = %v", err)
		}
		if snapshot == nil {
			t.Fatal("Snapshot() returned nil")
		}
	})

	t.Run("creates snapshot with locks", func(t *testing.T) {
		sm := NewLockStateMachine()

		// Acquire some locks
		for i := range 5 {
			key := "resource-" + string(rune('a'+i))
			cmd := lock.NewAcquireLockCommand(key, "client-1", 30*time.Second)
			log := createRaftLog(t, cmd, time.Now())
			sm.Apply(log)
		}

		snapshot, err := sm.Snapshot()
		if err != nil {
			t.Fatalf("Snapshot() error = %v", err)
		}
		if snapshot == nil {
			t.Fatal("Snapshot() returned nil")
		}
	})

	t.Run("snapshot returns LockStateMachineSnapshot type", func(t *testing.T) {
		sm := NewLockStateMachine()

		snapshot, err := sm.Snapshot()
		if err != nil {
			t.Fatalf("Snapshot() error = %v", err)
		}

		_, ok := snapshot.(*LockStateMachineSnapshot)
		if !ok {
			t.Errorf("expected *LockStateMachineSnapshot, got %T", snapshot)
		}
	})
}

func TestLockStateMachine_Restore(t *testing.T) {
	t.Run("restores empty state", func(t *testing.T) {
		sm1 := NewLockStateMachine()

		// Take snapshot of empty state
		snapshot, err := sm1.Snapshot()
		if err != nil {
			t.Fatalf("Snapshot() error = %v", err)
		}

		// Get the snapshot data
		lss := snapshot.(*LockStateMachineSnapshot)
		reader := io.NopCloser(bytes.NewReader(lss.data))

		// Restore to a new state machine
		sm2 := NewLockStateMachine()
		err = sm2.Restore(reader)
		if err != nil {
			t.Fatalf("Restore() error = %v", err)
		}
	})

	t.Run("restores state with locks", func(t *testing.T) {
		sm1 := NewLockStateMachine()

		// Acquire some locks
		keys := []string{"resource-a", "resource-b", "resource-c"}
		for _, key := range keys {
			cmd := lock.NewAcquireLockCommand(key, "client-1", 30*time.Second)
			log := createRaftLog(t, cmd, time.Now())
			sm1.Apply(log)
		}

		// Take snapshot
		snapshot, err := sm1.Snapshot()
		if err != nil {
			t.Fatalf("Snapshot() error = %v", err)
		}

		// Restore to a new state machine
		lss := snapshot.(*LockStateMachineSnapshot)
		reader := io.NopCloser(bytes.NewReader(lss.data))

		sm2 := NewLockStateMachine()
		err = sm2.Restore(reader)
		if err != nil {
			t.Fatalf("Restore() error = %v", err)
		}

		// Verify locks are restored - try to acquire with same client (should succeed)
		// and different client (should fail)
		for _, key := range keys {
			// Same client should get existing lock
			cmd := lock.NewAcquireLockCommand(key, "client-1", 30*time.Second)
			log := createRaftLog(t, cmd, time.Now())
			result := sm2.Apply(log)

			resp := result.(*lock.AcquireLockCommandResponse)
			if resp.Error != nil {
				t.Errorf("expected same client to reacquire lock %s: %v", key, *resp.Error)
			}

			// Different client should fail
			cmd2 := lock.NewAcquireLockCommand(key, "client-2", 30*time.Second)
			log2 := createRaftLog(t, cmd2, time.Now())
			result2 := sm2.Apply(log2)

			resp2 := result2.(*lock.AcquireLockCommandResponse)
			if resp2.Error == nil {
				t.Errorf("expected different client to fail acquiring lock %s", key)
			}
		}
	})

	t.Run("restores fence token", func(t *testing.T) {
		sm1 := NewLockStateMachine()

		// Acquire several locks to increment fence token
		for i := range 5 {
			key := "resource-" + string(rune('a'+i))
			cmd := lock.NewAcquireLockCommand(key, "client-1", 30*time.Second)
			log := createRaftLog(t, cmd, time.Now())
			result := sm1.Apply(log)

			// Release to allow next acquire
			resp := result.(*lock.AcquireLockCommandResponse)
			if resp.Error == nil {
				releaseCmd := lock.NewReleaseLockCommand(key, "client-1", resp.FenceToken)
				releaseLog := createRaftLog(t, releaseCmd, time.Now())
				sm1.Apply(releaseLog)
			}
		}

		// Take snapshot
		snapshot, err := sm1.Snapshot()
		if err != nil {
			t.Fatalf("Snapshot() error = %v", err)
		}

		// Restore to a new state machine
		lss := snapshot.(*LockStateMachineSnapshot)
		reader := io.NopCloser(bytes.NewReader(lss.data))

		sm2 := NewLockStateMachine()
		err = sm2.Restore(reader)
		if err != nil {
			t.Fatalf("Restore() error = %v", err)
		}

		// New lock should get fence token > 5
		cmd := lock.NewAcquireLockCommand("new-resource", "client-1", 30*time.Second)
		log := createRaftLog(t, cmd, time.Now())
		result := sm2.Apply(log)

		resp := result.(*lock.AcquireLockCommandResponse)
		if resp.Error != nil {
			t.Fatalf("failed to acquire new lock: %v", *resp.Error)
		}
		if resp.FenceToken <= 5 {
			t.Errorf("FenceToken = %d, want > 5 (restored fence token)", resp.FenceToken)
		}
	})

	t.Run("restore with invalid data returns error", func(t *testing.T) {
		sm := NewLockStateMachine()

		reader := io.NopCloser(bytes.NewReader([]byte("invalid json data")))
		err := sm.Restore(reader)
		if err == nil {
			t.Error("expected error when restoring invalid data")
		}
	})

	t.Run("restore with empty data returns error", func(t *testing.T) {
		sm := NewLockStateMachine()

		reader := io.NopCloser(bytes.NewReader([]byte{}))
		err := sm.Restore(reader)
		if err == nil {
			t.Error("expected error when restoring empty data")
		}
	})
}

func TestLockStateMachine_SnapshotRestore_RoundTrip(t *testing.T) {
	t.Run("full round trip preserves state", func(t *testing.T) {
		sm1 := NewLockStateMachine()

		// Create complex state
		type lockInfo struct {
			key        string
			clientID   string
			fenceToken int
		}
		locks := make([]lockInfo, 0)

		for i := range 10 {
			key := "resource-" + string(rune('a'+i))
			clientID := "client-" + string(rune('1'+i%3))

			cmd := lock.NewAcquireLockCommand(key, clientID, time.Hour)
			log := createRaftLog(t, cmd, time.Now())
			result := sm1.Apply(log)

			resp := result.(*lock.AcquireLockCommandResponse)
			if resp.Error == nil {
				locks = append(locks, lockInfo{
					key:        key,
					clientID:   clientID,
					fenceToken: resp.FenceToken,
				})
			}
		}

		// Take snapshot
		snapshot, err := sm1.Snapshot()
		if err != nil {
			t.Fatalf("Snapshot() error = %v", err)
		}

		// Persist snapshot to buffer
		var buf bytes.Buffer
		sink := &mockSnapshotSink{buf: &buf}
		err = snapshot.Persist(sink)
		if err != nil {
			t.Fatalf("Persist() error = %v", err)
		}

		// Restore to new state machine
		sm2 := NewLockStateMachine()
		reader := io.NopCloser(bytes.NewReader(buf.Bytes()))
		err = sm2.Restore(reader)
		if err != nil {
			t.Fatalf("Restore() error = %v", err)
		}

		// Verify all locks are preserved
		for _, l := range locks {
			// Try to release each lock - should succeed with correct credentials
			releaseCmd := lock.NewReleaseLockCommand(l.key, l.clientID, l.fenceToken)
			releaseLog := createRaftLog(t, releaseCmd, time.Now())
			result := sm2.Apply(releaseLog)

			resp := result.(*lock.LockCommandResponse)
			if resp.Error != nil {
				t.Errorf("failed to release lock %s: %v", l.key, *resp.Error)
			}
		}
	})
}

// mockSnapshotSink implements raft.SnapshotSink for testing.
type mockSnapshotSink struct {
	buf       *bytes.Buffer
	cancelled bool
	closed    bool
}

func (m *mockSnapshotSink) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}

func (m *mockSnapshotSink) Close() error {
	m.closed = true
	return nil
}

func (m *mockSnapshotSink) ID() string {
	return "mock-snapshot-id"
}

func (m *mockSnapshotSink) Cancel() error {
	m.cancelled = true
	return nil
}

func TestLockStateMachine_Apply_SequentialOperations(t *testing.T) {
	t.Run("acquire release acquire uses new fence token", func(t *testing.T) {
		sm := NewLockStateMachine()

		// First acquire
		cmd1 := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		log1 := createRaftLog(t, cmd1, time.Now())
		result1 := sm.Apply(log1)
		resp1 := result1.(*lock.AcquireLockCommandResponse)

		if resp1.Error != nil {
			t.Fatalf("first acquire failed: %v", *resp1.Error)
		}
		firstToken := resp1.FenceToken

		// Release
		releaseCmd := lock.NewReleaseLockCommand("resource-1", "client-1", firstToken)
		releaseLog := createRaftLog(t, releaseCmd, time.Now())
		sm.Apply(releaseLog)

		// Second acquire
		cmd2 := lock.NewAcquireLockCommand("resource-1", "client-2", 30*time.Second)
		log2 := createRaftLog(t, cmd2, time.Now())
		result2 := sm.Apply(log2)
		resp2 := result2.(*lock.AcquireLockCommandResponse)

		if resp2.Error != nil {
			t.Fatalf("second acquire failed: %v", *resp2.Error)
		}

		if resp2.FenceToken <= firstToken {
			t.Errorf("second FenceToken %d should be greater than first %d",
				resp2.FenceToken, firstToken)
		}
	})

	t.Run("renew does not change fence token", func(t *testing.T) {
		sm := NewLockStateMachine()

		// Acquire
		acquireCmd := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		acquireLog := createRaftLog(t, acquireCmd, time.Now())
		acquireResult := sm.Apply(acquireLog)
		acquireResp := acquireResult.(*lock.AcquireLockCommandResponse)

		if acquireResp.Error != nil {
			t.Fatalf("acquire failed: %v", *acquireResp.Error)
		}
		originalToken := acquireResp.FenceToken

		// Renew multiple times
		for i := range 5 {
			renewCmd := lock.NewRenewLockCommand("resource-1", "client-1", originalToken, 60*time.Second)
			renewLog := createRaftLog(t, renewCmd, time.Now().Add(time.Duration(i)*time.Second))
			renewResult := sm.Apply(renewLog)
			renewResp := renewResult.(*lock.RenewLockCommandResponse)

			if renewResp.Error != nil {
				t.Fatalf("renew %d failed: %v", i, *renewResp.Error)
			}
		}

		// Acquire another lock - should use next token
		acquireCmd2 := lock.NewAcquireLockCommand("resource-2", "client-1", 30*time.Second)
		acquireLog2 := createRaftLog(t, acquireCmd2, time.Now())
		acquireResult2 := sm.Apply(acquireLog2)
		acquireResp2 := acquireResult2.(*lock.AcquireLockCommandResponse)

		if acquireResp2.FenceToken != originalToken+1 {
			t.Errorf("FenceToken = %d, want %d (renew should not increment)",
				acquireResp2.FenceToken, originalToken+1)
		}
	})
}

func TestLockStateMachine_Apply_LockExpiration(t *testing.T) {
	t.Run("expired lock can be acquired by another client", func(t *testing.T) {
		sm := NewLockStateMachine()

		// First client acquires with very short TTL (expired by AppliedAt time)
		cmd1 := lock.NewAcquireLockCommand("resource-1", "client-1", 1*time.Millisecond)
		log1 := createRaftLog(t, cmd1, time.Now().Add(-10*time.Millisecond)) // Applied in past
		sm.Apply(log1)

		// Wait to ensure expiration
		time.Sleep(5 * time.Millisecond)

		// Second client can acquire
		cmd2 := lock.NewAcquireLockCommand("resource-1", "client-2", 30*time.Second)
		log2 := createRaftLog(t, cmd2, time.Now())
		result := sm.Apply(log2)

		resp := result.(*lock.AcquireLockCommandResponse)
		if resp.Error != nil {
			t.Errorf("expected to acquire expired lock: %v", *resp.Error)
		}
	})
}

// Benchmarks
func BenchmarkLockStateMachine_Apply_Acquire(b *testing.B) {
	sm := NewLockStateMachine()

	for i := 0; i < b.N; i++ {
		key := "resource-" + string(rune('a'+i%26))
		cmd := lock.NewAcquireLockCommand(key, "client-1", 30*time.Second)
		data, _ := lock.Encode(cmd)
		log := &raft.Log{
			Type:       raft.LogCommand,
			Data:       data,
			AppendedAt: time.Now(),
		}
		sm.Apply(log)
	}
}

func BenchmarkLockStateMachine_Apply_Release(b *testing.B) {
	sm := NewLockStateMachine()

	// Pre-acquire locks
	tokens := make([]int, b.N)
	for i := range b.N {
		key := "resource-" + string(rune('a'+i%26)) + "-" + string(rune('0'+i/26))
		cmd := lock.NewAcquireLockCommand(key, "client-1", time.Hour)
		data, _ := lock.Encode(cmd)
		log := &raft.Log{
			Type:       raft.LogCommand,
			Data:       data,
			AppendedAt: time.Now(),
		}
		result := sm.Apply(log)
		if resp, ok := result.(*lock.AcquireLockCommandResponse); ok && resp.Error == nil {
			tokens[i] = resp.FenceToken
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "resource-" + string(rune('a'+i%26)) + "-" + string(rune('0'+i/26))
		cmd := lock.NewReleaseLockCommand(key, "client-1", tokens[i])
		data, _ := lock.Encode(cmd)
		log := &raft.Log{
			Type:       raft.LogCommand,
			Data:       data,
			AppendedAt: time.Now(),
		}
		sm.Apply(log)
	}
}

func BenchmarkLockStateMachine_Snapshot(b *testing.B) {
	sm := NewLockStateMachine()

	// Create some locks
	for i := range 100 {
		key := "resource-" + string(rune('a'+i%26)) + "-" + string(rune('0'+i/26))
		cmd := lock.NewAcquireLockCommand(key, "client-1", time.Hour)
		data, _ := lock.Encode(cmd)
		log := &raft.Log{
			Type:       raft.LogCommand,
			Data:       data,
			AppendedAt: time.Now(),
		}
		sm.Apply(log)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sm.Snapshot()
	}
}
