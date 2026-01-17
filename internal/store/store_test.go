package store

import (
	"strings"
	"testing"
	"time"

	"github.com/myamout/locksmith/internal/lock"
)

// newTestStore creates a new LockSessionStore for testing.
func newTestStore() *LockSessionStore {
	return &LockSessionStore{
		locks:      make(map[string]*lock.LockEntry),
		fenceToken: 0,
	}
}

// acquireLock is a helper that acquires a lock and returns the response.
func acquireLock(t *testing.T, store *LockSessionStore, key, clientID string, expiresIn time.Duration) *lock.AcquireLockCommandResponse {
	t.Helper()
	cmd := lock.NewAcquireLockCommand(key, clientID, expiresIn)
	cmd.SetAppliedAt(time.Now())
	resp := store.Action(cmd)
	return resp.(*lock.AcquireLockCommandResponse)
}

func TestLockSessionStore_Action_Acquire(t *testing.T) {
	t.Run("acquire new lock successfully", func(t *testing.T) {
		store := newTestStore()
		cmd := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		cmd.SetAppliedAt(time.Now())

		resp := store.Action(cmd)
		acquireResp, ok := resp.(*lock.AcquireLockCommandResponse)
		if !ok {
			t.Fatalf("expected *lock.AcquireLockCommandResponse, got %T", resp)
		}

		if acquireResp.Error != nil {
			t.Errorf("unexpected error: %v", *acquireResp.Error)
		}
		if acquireResp.FenceToken != 1 {
			t.Errorf("FenceToken = %d, want 1", acquireResp.FenceToken)
		}
		if acquireResp.HeldBy != "" {
			t.Errorf("HeldBy = %q, want empty string for new lock", acquireResp.HeldBy)
		}
	})

	t.Run("acquire same lock by same client returns existing lock", func(t *testing.T) {
		store := newTestStore()

		// First acquire
		cmd1 := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		cmd1.SetAppliedAt(time.Now())
		resp1 := store.Action(cmd1).(*lock.AcquireLockCommandResponse)

		if resp1.Error != nil {
			t.Fatalf("first acquire failed: %v", *resp1.Error)
		}

		// Second acquire by same client
		cmd2 := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		cmd2.SetAppliedAt(time.Now())
		resp2 := store.Action(cmd2).(*lock.AcquireLockCommandResponse)

		if resp2.Error != nil {
			t.Errorf("second acquire should succeed for same client: %v", *resp2.Error)
		}
		if resp2.FenceToken != resp1.FenceToken {
			t.Errorf("FenceToken = %d, want %d (same as original)", resp2.FenceToken, resp1.FenceToken)
		}
	})

	t.Run("acquire lock held by different client fails", func(t *testing.T) {
		store := newTestStore()

		// First client acquires lock
		cmd1 := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		cmd1.SetAppliedAt(time.Now())
		resp1 := store.Action(cmd1).(*lock.AcquireLockCommandResponse)

		if resp1.Error != nil {
			t.Fatalf("first acquire failed: %v", *resp1.Error)
		}

		// Second client tries to acquire same lock
		cmd2 := lock.NewAcquireLockCommand("resource-1", "client-2", 30*time.Second)
		cmd2.SetAppliedAt(time.Now())
		resp2 := store.Action(cmd2).(*lock.AcquireLockCommandResponse)

		if resp2.Error == nil {
			t.Error("expected error when acquiring lock held by another client")
		}
		if resp2.HeldBy != "client-1" {
			t.Errorf("HeldBy = %q, want %q", resp2.HeldBy, "client-1")
		}
	})

	t.Run("acquire expired lock succeeds", func(t *testing.T) {
		store := newTestStore()

		// First client acquires lock with short expiration
		cmd1 := lock.NewAcquireLockCommand("resource-1", "client-1", 1*time.Millisecond)
		cmd1.SetAppliedAt(time.Now().Add(-10 * time.Millisecond)) // Already expired
		store.Action(cmd1)

		// Wait to ensure lock is expired
		time.Sleep(5 * time.Millisecond)

		// Second client can now acquire the expired lock
		cmd2 := lock.NewAcquireLockCommand("resource-1", "client-2", 30*time.Second)
		cmd2.SetAppliedAt(time.Now())
		resp2 := store.Action(cmd2).(*lock.AcquireLockCommandResponse)

		if resp2.Error != nil {
			t.Errorf("expected to acquire expired lock, got error: %v", *resp2.Error)
		}
		if resp2.FenceToken != 2 {
			t.Errorf("FenceToken = %d, want 2 (new token after expired lock)", resp2.FenceToken)
		}
	})

	t.Run("fence token increments with each new lock", func(t *testing.T) {
		store := newTestStore()

		keys := []string{"resource-1", "resource-2", "resource-3"}
		for i, key := range keys {
			cmd := lock.NewAcquireLockCommand(key, "client-1", 30*time.Second)
			cmd.SetAppliedAt(time.Now())
			resp := store.Action(cmd).(*lock.AcquireLockCommandResponse)

			expectedToken := i + 1
			if resp.FenceToken != expectedToken {
				t.Errorf("FenceToken for %s = %d, want %d", key, resp.FenceToken, expectedToken)
			}
		}
	})
}

func TestLockSessionStore_Action_Acquire_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		clientID      string
		expiresIn     time.Duration
		wantError     bool
		wantFenceToken int
	}{
		{
			name:          "basic acquire",
			key:           "test-key",
			clientID:      "client-123",
			expiresIn:     30 * time.Second,
			wantError:     false,
			wantFenceToken: 1,
		},
		{
			name:          "acquire with empty key",
			key:           "",
			clientID:      "client-456",
			expiresIn:     time.Minute,
			wantError:     false,
			wantFenceToken: 1,
		},
		{
			name:          "acquire with empty client ID",
			key:           "another-key",
			clientID:      "",
			expiresIn:     time.Hour,
			wantError:     false,
			wantFenceToken: 1,
		},
		{
			name:          "acquire with special characters in key",
			key:           "lock/resource/123",
			clientID:      "client-789",
			expiresIn:     time.Second,
			wantError:     false,
			wantFenceToken: 1,
		},
		{
			name:          "acquire with UUID client ID",
			key:           "uuid-key",
			clientID:      "550e8400-e29b-41d4-a716-446655440000",
			expiresIn:     5 * time.Minute,
			wantError:     false,
			wantFenceToken: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore()
			cmd := lock.NewAcquireLockCommand(tt.key, tt.clientID, tt.expiresIn)
			cmd.SetAppliedAt(time.Now())

			resp := store.Action(cmd).(*lock.AcquireLockCommandResponse)

			if tt.wantError && resp.Error == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && resp.Error != nil {
				t.Errorf("unexpected error: %v", *resp.Error)
			}
			if resp.FenceToken != tt.wantFenceToken {
				t.Errorf("FenceToken = %d, want %d", resp.FenceToken, tt.wantFenceToken)
			}
		})
	}
}

func TestLockSessionStore_Action_Release(t *testing.T) {
	t.Run("release lock successfully", func(t *testing.T) {
		store := newTestStore()

		// Acquire a lock first
		acquireResp := acquireLock(t, store, "resource-1", "client-1", 30*time.Second)
		if acquireResp.Error != nil {
			t.Fatalf("failed to acquire lock: %v", *acquireResp.Error)
		}

		// Release the lock
		releaseCmd := lock.NewReleaseLockCommand("resource-1", "client-1", acquireResp.FenceToken)
		resp := store.Action(releaseCmd)
		releaseResp, ok := resp.(*lock.LockCommandResponse)
		if !ok {
			t.Fatalf("expected *lock.LockCommandResponse, got %T", resp)
		}

		if releaseResp.Error != nil {
			t.Errorf("unexpected error: %v", *releaseResp.Error)
		}
		if !releaseResp.Success() {
			t.Error("expected Success() to return true")
		}
	})

	t.Run("release non-existent lock fails", func(t *testing.T) {
		store := newTestStore()

		releaseCmd := lock.NewReleaseLockCommand("non-existent", "client-1", 1)
		resp := store.Action(releaseCmd).(*lock.LockCommandResponse)

		if resp.Error == nil {
			t.Error("expected error when releasing non-existent lock")
		}
		if resp.Success() {
			t.Error("expected Success() to return false")
		}
	})

	t.Run("release lock with wrong client ID fails", func(t *testing.T) {
		store := newTestStore()

		// Acquire lock
		acquireResp := acquireLock(t, store, "resource-1", "client-1", 30*time.Second)

		// Try to release with wrong client ID
		releaseCmd := lock.NewReleaseLockCommand("resource-1", "client-2", acquireResp.FenceToken)
		resp := store.Action(releaseCmd).(*lock.LockCommandResponse)

		if resp.Error == nil {
			t.Error("expected error when releasing with wrong client ID")
		}
	})

	t.Run("release lock with wrong fence token fails", func(t *testing.T) {
		store := newTestStore()

		// Acquire lock
		acquireLock(t, store, "resource-1", "client-1", 30*time.Second)

		// Try to release with wrong fence token
		releaseCmd := lock.NewReleaseLockCommand("resource-1", "client-1", 999)
		resp := store.Action(releaseCmd).(*lock.LockCommandResponse)

		if resp.Error == nil {
			t.Error("expected error when releasing with wrong fence token")
		}
	})

	t.Run("release already released lock fails", func(t *testing.T) {
		store := newTestStore()

		// Acquire and release
		acquireResp := acquireLock(t, store, "resource-1", "client-1", 30*time.Second)
		releaseCmd := lock.NewReleaseLockCommand("resource-1", "client-1", acquireResp.FenceToken)
		store.Action(releaseCmd)

		// Try to release again
		resp := store.Action(releaseCmd).(*lock.LockCommandResponse)

		if resp.Error == nil {
			t.Error("expected error when releasing already released lock")
		}
	})

	t.Run("can acquire lock after release", func(t *testing.T) {
		store := newTestStore()

		// Acquire and release
		acquireResp1 := acquireLock(t, store, "resource-1", "client-1", 30*time.Second)
		releaseCmd := lock.NewReleaseLockCommand("resource-1", "client-1", acquireResp1.FenceToken)
		store.Action(releaseCmd)

		// Acquire again (possibly by different client)
		acquireResp2 := acquireLock(t, store, "resource-1", "client-2", 30*time.Second)

		if acquireResp2.Error != nil {
			t.Errorf("expected to acquire released lock, got error: %v", *acquireResp2.Error)
		}
		if acquireResp2.FenceToken != 2 {
			t.Errorf("FenceToken = %d, want 2 (incremented after release)", acquireResp2.FenceToken)
		}
	})
}

func TestLockSessionStore_Action_Release_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		setupLock      bool
		releaseKey     string
		releaseClient  string
		releaseFence   int
		lockKey        string
		lockClient     string
		lockFenceToken int
		wantError      bool
	}{
		{
			name:           "successful release",
			setupLock:      true,
			lockKey:        "key-1",
			lockClient:     "client-1",
			lockFenceToken: 1,
			releaseKey:     "key-1",
			releaseClient:  "client-1",
			releaseFence:   1,
			wantError:      false,
		},
		{
			name:          "release without existing lock",
			setupLock:     false,
			releaseKey:    "key-1",
			releaseClient: "client-1",
			releaseFence:  1,
			wantError:     true,
		},
		{
			name:           "release with wrong key",
			setupLock:      true,
			lockKey:        "key-1",
			lockClient:     "client-1",
			lockFenceToken: 1,
			releaseKey:     "wrong-key",
			releaseClient:  "client-1",
			releaseFence:   1,
			wantError:      true,
		},
		{
			name:           "release with wrong client",
			setupLock:      true,
			lockKey:        "key-1",
			lockClient:     "client-1",
			lockFenceToken: 1,
			releaseKey:     "key-1",
			releaseClient:  "wrong-client",
			releaseFence:   1,
			wantError:      true,
		},
		{
			name:           "release with wrong fence token",
			setupLock:      true,
			lockKey:        "key-1",
			lockClient:     "client-1",
			lockFenceToken: 1,
			releaseKey:     "key-1",
			releaseClient:  "client-1",
			releaseFence:   999,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore()

			if tt.setupLock {
				// Manually set up lock in store to control fence token
				store.locks[tt.lockKey] = &lock.LockEntry{
					Key:        tt.lockKey,
					ClientID:   tt.lockClient,
					FenceToken: tt.lockFenceToken,
					ExpiresAt:  time.Now().Add(30 * time.Second),
					CreatedAt:  time.Now(),
				}
			}

			releaseCmd := lock.NewReleaseLockCommand(tt.releaseKey, tt.releaseClient, tt.releaseFence)
			resp := store.Action(releaseCmd).(*lock.LockCommandResponse)

			if tt.wantError && resp.Error == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && resp.Error != nil {
				t.Errorf("unexpected error: %v", *resp.Error)
			}
		})
	}
}

func TestLockSessionStore_Action_Renew(t *testing.T) {
	t.Run("renew lock successfully", func(t *testing.T) {
		store := newTestStore()

		// Acquire a lock
		acquireResp := acquireLock(t, store, "resource-1", "client-1", 30*time.Second)
		if acquireResp.Error != nil {
			t.Fatalf("failed to acquire lock: %v", *acquireResp.Error)
		}

		// Renew the lock
		renewCmd := lock.NewRenewLockCommand("resource-1", "client-1", acquireResp.FenceToken, 60*time.Second)
		renewCmd.SetAppliedAt(time.Now())
		resp := store.Action(renewCmd)
		renewResp, ok := resp.(*lock.RenewLockCommandResponse)
		if !ok {
			t.Fatalf("expected *lock.RenewLockCommandResponse, got %T", resp)
		}

		if renewResp.Error != nil {
			t.Errorf("unexpected error: %v", *renewResp.Error)
		}
		if renewResp.ExpiresAt.IsZero() {
			t.Error("expected non-zero ExpiresAt")
		}
	})

	t.Run("renew non-existent lock fails", func(t *testing.T) {
		store := newTestStore()

		renewCmd := lock.NewRenewLockCommand("non-existent", "client-1", 1, 30*time.Second)
		renewCmd.SetAppliedAt(time.Now())
		resp := store.Action(renewCmd).(*lock.RenewLockCommandResponse)

		if resp.Error == nil {
			t.Error("expected error when renewing non-existent lock")
		}
	})

	t.Run("renew expired lock fails", func(t *testing.T) {
		store := newTestStore()

		// Create an already expired lock
		store.locks["resource-1"] = &lock.LockEntry{
			Key:        "resource-1",
			ClientID:   "client-1",
			FenceToken: 1,
			ExpiresAt:  time.Now().Add(-10 * time.Second), // Already expired
			CreatedAt:  time.Now().Add(-20 * time.Second),
		}

		renewCmd := lock.NewRenewLockCommand("resource-1", "client-1", 1, 30*time.Second)
		renewCmd.SetAppliedAt(time.Now())
		resp := store.Action(renewCmd).(*lock.RenewLockCommandResponse)

		if resp.Error == nil {
			t.Error("expected error when renewing expired lock")
		}
	})

	t.Run("renew lock with wrong client ID fails", func(t *testing.T) {
		store := newTestStore()

		// Acquire lock
		acquireResp := acquireLock(t, store, "resource-1", "client-1", 30*time.Second)

		// Try to renew with wrong client ID
		renewCmd := lock.NewRenewLockCommand("resource-1", "client-2", acquireResp.FenceToken, 30*time.Second)
		renewCmd.SetAppliedAt(time.Now())
		resp := store.Action(renewCmd).(*lock.RenewLockCommandResponse)

		if resp.Error == nil {
			t.Error("expected error when renewing with wrong client ID")
		}
	})

	t.Run("renew lock with wrong fence token fails", func(t *testing.T) {
		store := newTestStore()

		// Acquire lock
		acquireLock(t, store, "resource-1", "client-1", 30*time.Second)

		// Try to renew with wrong fence token
		renewCmd := lock.NewRenewLockCommand("resource-1", "client-1", 999, 30*time.Second)
		renewCmd.SetAppliedAt(time.Now())
		resp := store.Action(renewCmd).(*lock.RenewLockCommandResponse)

		if resp.Error == nil {
			t.Error("expected error when renewing with wrong fence token")
		}
	})

	t.Run("renew updates expiration time", func(t *testing.T) {
		store := newTestStore()

		// Acquire lock with short expiration
		acquireResp := acquireLock(t, store, "resource-1", "client-1", 5*time.Second)
		originalExpiry := acquireResp.ExpiresAt

		// Wait a bit
		time.Sleep(10 * time.Millisecond)

		// Renew with longer expiration
		renewCmd := lock.NewRenewLockCommand("resource-1", "client-1", acquireResp.FenceToken, 60*time.Second)
		renewCmd.SetAppliedAt(time.Now())
		renewResp := store.Action(renewCmd).(*lock.RenewLockCommandResponse)

		if renewResp.Error != nil {
			t.Fatalf("failed to renew: %v", *renewResp.Error)
		}
		if !renewResp.ExpiresAt.After(originalExpiry) {
			t.Errorf("ExpiresAt = %v, want after %v", renewResp.ExpiresAt, originalExpiry)
		}
	})
}

func TestLockSessionStore_Action_Renew_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		lockKey      string
		lockClient   string
		lockFence    int
		lockExpired  bool
		renewKey     string
		renewClient  string
		renewFence   int
		renewExpires time.Duration
		wantError    bool
	}{
		{
			name:         "successful renew",
			lockKey:      "key-1",
			lockClient:   "client-1",
			lockFence:    1,
			lockExpired:  false,
			renewKey:     "key-1",
			renewClient:  "client-1",
			renewFence:   1,
			renewExpires: 30 * time.Second,
			wantError:    false,
		},
		{
			name:         "renew with wrong key",
			lockKey:      "key-1",
			lockClient:   "client-1",
			lockFence:    1,
			lockExpired:  false,
			renewKey:     "wrong-key",
			renewClient:  "client-1",
			renewFence:   1,
			renewExpires: 30 * time.Second,
			wantError:    true,
		},
		{
			name:         "renew with wrong client",
			lockKey:      "key-1",
			lockClient:   "client-1",
			lockFence:    1,
			lockExpired:  false,
			renewKey:     "key-1",
			renewClient:  "wrong-client",
			renewFence:   1,
			renewExpires: 30 * time.Second,
			wantError:    true,
		},
		{
			name:         "renew with wrong fence token",
			lockKey:      "key-1",
			lockClient:   "client-1",
			lockFence:    1,
			lockExpired:  false,
			renewKey:     "key-1",
			renewClient:  "client-1",
			renewFence:   999,
			renewExpires: 30 * time.Second,
			wantError:    true,
		},
		{
			name:         "renew expired lock",
			lockKey:      "key-1",
			lockClient:   "client-1",
			lockFence:    1,
			lockExpired:  true,
			renewKey:     "key-1",
			renewClient:  "client-1",
			renewFence:   1,
			renewExpires: 30 * time.Second,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore()

			// Set up lock
			expiresAt := time.Now().Add(30 * time.Second)
			if tt.lockExpired {
				expiresAt = time.Now().Add(-10 * time.Second)
			}
			store.locks[tt.lockKey] = &lock.LockEntry{
				Key:        tt.lockKey,
				ClientID:   tt.lockClient,
				FenceToken: tt.lockFence,
				ExpiresAt:  expiresAt,
				CreatedAt:  time.Now(),
			}

			renewCmd := lock.NewRenewLockCommand(tt.renewKey, tt.renewClient, tt.renewFence, tt.renewExpires)
			renewCmd.SetAppliedAt(time.Now())
			resp := store.Action(renewCmd).(*lock.RenewLockCommandResponse)

			if tt.wantError && resp.Error == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && resp.Error != nil {
				t.Errorf("unexpected error: %v", *resp.Error)
			}
		})
	}
}

func TestLockSessionStore_Action_InvalidCommand(t *testing.T) {
	t.Run("invalid command type returns error", func(t *testing.T) {
		store := newTestStore()

		// Create a custom command type that implements LockCommand but is not recognized
		var invalidCmd lock.LockCommand = &mockInvalidCommand{}

		resp := store.Action(invalidCmd)
		cmdResp, ok := resp.(*lock.LockCommandResponse)
		if !ok {
			t.Fatalf("expected *lock.LockCommandResponse, got %T", resp)
		}

		if cmdResp.Error == nil {
			t.Error("expected error for invalid command type")
		}
	})
}

// mockInvalidCommand implements LockCommand but is not a recognized command type.
type mockInvalidCommand struct{}

func (m *mockInvalidCommand) Type() lock.LockCommandType {
	return lock.LockCommandType(999) // Unknown type
}

func (m *mockInvalidCommand) Key() string {
	return "mock-key"
}

func TestLockSessionStore_FenceTokenMonotonicity(t *testing.T) {
	t.Run("fence token is monotonically increasing", func(t *testing.T) {
		store := newTestStore()

		var lastToken int
		for i := range 100 {
			key := "resource"
			clientID := "client-1"

			// Acquire lock
			acquireResp := acquireLock(t, store, key, clientID, 30*time.Second)
			if acquireResp.Error != nil {
				t.Fatalf("failed to acquire lock on iteration %d: %v", i, *acquireResp.Error)
			}

			if acquireResp.FenceToken <= lastToken {
				t.Errorf("iteration %d: FenceToken %d is not greater than previous %d",
					i, acquireResp.FenceToken, lastToken)
			}
			lastToken = acquireResp.FenceToken

			// Release lock
			releaseCmd := lock.NewReleaseLockCommand(key, clientID, acquireResp.FenceToken)
			store.Action(releaseCmd)
		}
	})

	t.Run("fence token increments across different keys", func(t *testing.T) {
		store := newTestStore()

		tokens := make([]int, 0, 10)
		for i := range 10 {
			key := "resource-" + string(rune('a'+i))
			acquireResp := acquireLock(t, store, key, "client-1", 30*time.Second)
			if acquireResp.Error != nil {
				t.Fatalf("failed to acquire lock for %s: %v", key, *acquireResp.Error)
			}
			tokens = append(tokens, acquireResp.FenceToken)
		}

		// Verify tokens are strictly increasing
		for i := 1; i < len(tokens); i++ {
			if tokens[i] <= tokens[i-1] {
				t.Errorf("token[%d]=%d should be greater than token[%d]=%d",
					i, tokens[i], i-1, tokens[i-1])
			}
		}
	})
}

func TestLockSessionStore_MultipleLocks(t *testing.T) {
	t.Run("can hold multiple locks simultaneously", func(t *testing.T) {
		store := newTestStore()

		keys := []string{"resource-1", "resource-2", "resource-3", "resource-4", "resource-5"}
		tokens := make(map[string]int)

		// Acquire all locks
		for _, key := range keys {
			resp := acquireLock(t, store, key, "client-1", 30*time.Second)
			if resp.Error != nil {
				t.Fatalf("failed to acquire lock %s: %v", key, *resp.Error)
			}
			tokens[key] = resp.FenceToken
		}

		// Verify all locks are held
		if len(store.locks) != len(keys) {
			t.Errorf("expected %d locks, got %d", len(keys), len(store.locks))
		}

		// Release all locks
		for _, key := range keys {
			releaseCmd := lock.NewReleaseLockCommand(key, "client-1", tokens[key])
			resp := store.Action(releaseCmd).(*lock.LockCommandResponse)
			if resp.Error != nil {
				t.Errorf("failed to release lock %s: %v", key, *resp.Error)
			}
		}

		// Verify all locks are released
		if len(store.locks) != 0 {
			t.Errorf("expected 0 locks after release, got %d", len(store.locks))
		}
	})

	t.Run("different clients can hold different locks", func(t *testing.T) {
		store := newTestStore()

		// Client 1 acquires resource-1
		resp1 := acquireLock(t, store, "resource-1", "client-1", 30*time.Second)
		if resp1.Error != nil {
			t.Fatalf("client-1 failed to acquire resource-1: %v", *resp1.Error)
		}

		// Client 2 acquires resource-2
		resp2 := acquireLock(t, store, "resource-2", "client-2", 30*time.Second)
		if resp2.Error != nil {
			t.Fatalf("client-2 failed to acquire resource-2: %v", *resp2.Error)
		}

		// Verify both locks are held by correct clients
		if store.locks["resource-1"].ClientID != "client-1" {
			t.Errorf("resource-1 held by %s, want client-1", store.locks["resource-1"].ClientID)
		}
		if store.locks["resource-2"].ClientID != "client-2" {
			t.Errorf("resource-2 held by %s, want client-2", store.locks["resource-2"].ClientID)
		}
	})
}

func TestLockSessionStore_EdgeCases(t *testing.T) {
	t.Run("empty key works", func(t *testing.T) {
		store := newTestStore()

		resp := acquireLock(t, store, "", "client-1", 30*time.Second)
		if resp.Error != nil {
			t.Errorf("failed to acquire lock with empty key: %v", *resp.Error)
		}
	})

	t.Run("empty client ID works", func(t *testing.T) {
		store := newTestStore()

		resp := acquireLock(t, store, "resource-1", "", 30*time.Second)
		if resp.Error != nil {
			t.Errorf("failed to acquire lock with empty client ID: %v", *resp.Error)
		}
	})

	t.Run("very long key works", func(t *testing.T) {
		store := newTestStore()

		longKey := strings.Repeat("a", 1000)

		resp := acquireLock(t, store, longKey, "client-1", 30*time.Second)
		if resp.Error != nil {
			t.Errorf("failed to acquire lock with long key: %v", *resp.Error)
		}
	})

	t.Run("special characters in key", func(t *testing.T) {
		store := newTestStore()

		specialKeys := []string{
			"key/with/slashes",
			"key:with:colons",
			"key.with.dots",
			"key-with-dashes",
			"key_with_underscores",
			"key with spaces",
			"key\twith\ttabs",
			"key\nwith\nnewlines",
		}

		for _, key := range specialKeys {
			resp := acquireLock(t, store, key, "client-1", 30*time.Second)
			if resp.Error != nil {
				t.Errorf("failed to acquire lock with key %q: %v", key, *resp.Error)
			}
		}
	})

	t.Run("zero expiration time", func(t *testing.T) {
		store := newTestStore()

		cmd := lock.NewAcquireLockCommand("resource-1", "client-1", 0)
		cmd.SetAppliedAt(time.Now())
		resp := store.Action(cmd).(*lock.AcquireLockCommandResponse)

		// Lock is acquired but will be immediately expired
		if resp.Error != nil {
			t.Errorf("failed to acquire lock with zero expiration: %v", *resp.Error)
		}
	})

	t.Run("negative expiration is set but lock expires immediately", func(t *testing.T) {
		store := newTestStore()

		cmd := lock.NewAcquireLockCommand("resource-1", "client-1", -time.Second)
		cmd.SetAppliedAt(time.Now())
		resp := store.Action(cmd).(*lock.AcquireLockCommandResponse)

		if resp.Error != nil {
			t.Errorf("failed to acquire lock with negative expiration: %v", *resp.Error)
		}

		// Another client should be able to acquire since it's expired
		cmd2 := lock.NewAcquireLockCommand("resource-1", "client-2", 30*time.Second)
		cmd2.SetAppliedAt(time.Now())
		resp2 := store.Action(cmd2).(*lock.AcquireLockCommandResponse)

		if resp2.Error != nil {
			t.Errorf("failed to acquire expired lock: %v", *resp2.Error)
		}
	})
}

func TestLockSessionStore_ResponseTypes(t *testing.T) {
	t.Run("acquire returns correct response type", func(t *testing.T) {
		store := newTestStore()
		cmd := lock.NewAcquireLockCommand("key", "client", time.Second)
		cmd.SetAppliedAt(time.Now())

		resp := store.Action(cmd)
		if _, ok := resp.(*lock.AcquireLockCommandResponse); !ok {
			t.Errorf("expected *lock.AcquireLockCommandResponse, got %T", resp)
		}
	})

	t.Run("release returns correct response type", func(t *testing.T) {
		store := newTestStore()

		// First acquire
		acquireResp := acquireLock(t, store, "key", "client", time.Second)

		// Then release
		releaseCmd := lock.NewReleaseLockCommand("key", "client", acquireResp.FenceToken)
		resp := store.Action(releaseCmd)
		if _, ok := resp.(*lock.LockCommandResponse); !ok {
			t.Errorf("expected *lock.LockCommandResponse, got %T", resp)
		}
	})

	t.Run("renew returns correct response type", func(t *testing.T) {
		store := newTestStore()

		// First acquire
		acquireResp := acquireLock(t, store, "key", "client", time.Second)

		// Then renew
		renewCmd := lock.NewRenewLockCommand("key", "client", acquireResp.FenceToken, time.Second)
		renewCmd.SetAppliedAt(time.Now())
		resp := store.Action(renewCmd)
		if _, ok := resp.(*lock.RenewLockCommandResponse); !ok {
			t.Errorf("expected *lock.RenewLockCommandResponse, got %T", resp)
		}
	})

	t.Run("invalid command returns LockCommandResponse", func(t *testing.T) {
		store := newTestStore()
		resp := store.Action(&mockInvalidCommand{})
		if _, ok := resp.(*lock.LockCommandResponse); !ok {
			t.Errorf("expected *lock.LockCommandResponse, got %T", resp)
		}
	})
}

func TestLockSessionStore_LockEntryIntegrity(t *testing.T) {
	t.Run("lock entry has correct fields after acquire", func(t *testing.T) {
		store := newTestStore()
		now := time.Now()

		cmd := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		cmd.SetAppliedAt(now)
		store.Action(cmd)

		entry := store.locks["resource-1"]
		if entry == nil {
			t.Fatal("expected lock entry to exist")
		}
		if entry.Key != "resource-1" {
			t.Errorf("Key = %q, want %q", entry.Key, "resource-1")
		}
		if entry.ClientID != "client-1" {
			t.Errorf("ClientID = %q, want %q", entry.ClientID, "client-1")
		}
		if entry.FenceToken != 1 {
			t.Errorf("FenceToken = %d, want 1", entry.FenceToken)
		}
		expectedExpiry := now.Add(30 * time.Second)
		if !entry.ExpiresAt.Equal(expectedExpiry) {
			t.Errorf("ExpiresAt = %v, want %v", entry.ExpiresAt, expectedExpiry)
		}
		if !entry.CreatedAt.Equal(now) {
			t.Errorf("CreatedAt = %v, want %v", entry.CreatedAt, now)
		}
	})

	t.Run("lock entry is removed after release", func(t *testing.T) {
		store := newTestStore()

		acquireResp := acquireLock(t, store, "resource-1", "client-1", 30*time.Second)
		if store.locks["resource-1"] == nil {
			t.Fatal("expected lock entry to exist after acquire")
		}

		releaseCmd := lock.NewReleaseLockCommand("resource-1", "client-1", acquireResp.FenceToken)
		store.Action(releaseCmd)

		if store.locks["resource-1"] != nil {
			t.Error("expected lock entry to be nil after release")
		}
	})

	t.Run("lock entry expiry is updated after renew", func(t *testing.T) {
		store := newTestStore()

		acquireResp := acquireLock(t, store, "resource-1", "client-1", 30*time.Second)
		originalExpiry := store.locks["resource-1"].ExpiresAt

		renewCmd := lock.NewRenewLockCommand("resource-1", "client-1", acquireResp.FenceToken, 60*time.Second)
		renewCmd.SetAppliedAt(time.Now())
		store.Action(renewCmd)

		newExpiry := store.locks["resource-1"].ExpiresAt
		if !newExpiry.After(originalExpiry) {
			t.Errorf("ExpiresAt = %v, should be after %v", newExpiry, originalExpiry)
		}
	})
}

func TestLockSessionStore_AcquireLockResponse_Fields(t *testing.T) {
	t.Run("successful acquire has correct fields", func(t *testing.T) {
		store := newTestStore()
		cmd := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		cmd.SetAppliedAt(time.Now())

		resp := store.Action(cmd).(*lock.AcquireLockCommandResponse)

		if resp.Error != nil {
			t.Errorf("Error should be nil, got %v", *resp.Error)
		}
		if resp.FenceToken != 1 {
			t.Errorf("FenceToken = %d, want 1", resp.FenceToken)
		}
		if resp.HeldBy != "" {
			t.Errorf("HeldBy = %q, want empty", resp.HeldBy)
		}
		if resp.ExpiresAt.IsZero() {
			t.Error("ExpiresAt should not be zero")
		}
	})

	t.Run("failed acquire has correct fields", func(t *testing.T) {
		store := newTestStore()

		// First acquire succeeds
		cmd1 := lock.NewAcquireLockCommand("resource-1", "client-1", 30*time.Second)
		cmd1.SetAppliedAt(time.Now())
		store.Action(cmd1)

		// Second acquire by different client fails
		cmd2 := lock.NewAcquireLockCommand("resource-1", "client-2", 30*time.Second)
		cmd2.SetAppliedAt(time.Now())
		resp := store.Action(cmd2).(*lock.AcquireLockCommandResponse)

		if resp.Error == nil {
			t.Error("Error should not be nil for failed acquire")
		}
		if resp.HeldBy != "client-1" {
			t.Errorf("HeldBy = %q, want %q", resp.HeldBy, "client-1")
		}
		if resp.ExpiresAt.IsZero() {
			t.Error("ExpiresAt should indicate when lock expires")
		}
	})
}

func TestLockSessionStore_ReAcquireByExpiration(t *testing.T) {
	t.Run("lock can be acquired after expiration", func(t *testing.T) {
		store := newTestStore()

		// Client 1 acquires with very short TTL
		cmd1 := lock.NewAcquireLockCommand("resource-1", "client-1", 50*time.Millisecond)
		cmd1.SetAppliedAt(time.Now())
		resp1 := store.Action(cmd1).(*lock.AcquireLockCommandResponse)
		if resp1.Error != nil {
			t.Fatalf("first acquire failed: %v", *resp1.Error)
		}

		// Wait for lock to expire
		time.Sleep(100 * time.Millisecond)

		// Client 2 can now acquire the same resource
		cmd2 := lock.NewAcquireLockCommand("resource-1", "client-2", 30*time.Second)
		cmd2.SetAppliedAt(time.Now())
		resp2 := store.Action(cmd2).(*lock.AcquireLockCommandResponse)

		if resp2.Error != nil {
			t.Errorf("second acquire should succeed after expiration: %v", *resp2.Error)
		}
		if resp2.FenceToken <= resp1.FenceToken {
			t.Errorf("new FenceToken %d should be greater than old %d",
				resp2.FenceToken, resp1.FenceToken)
		}
	})
}
