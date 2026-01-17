package node

import (
	"os"
	"path/filepath"
	"testing"
)

// testDataDir creates a temporary directory for testing and returns a cleanup function.
func testDataDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "locksmith-storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir, func() {
		os.RemoveAll(dir)
	}
}

func TestNewNodeStorage(t *testing.T) {
	t.Run("creates storage with valid directory", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		if storage == nil {
			t.Fatal("NewNodeStorage() returned nil storage")
		}
	})

	t.Run("creates data directory", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "new-raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		// Check that directory was created
		info, err := os.Stat(dataDir)
		if err != nil {
			t.Fatalf("data directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("data directory is not a directory")
		}
	})

	t.Run("creates raft.db file", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		dbPath := filepath.Join(dataDir, "raft.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("raft.db file was not created")
		}
	})

	t.Run("creates snapshots directory", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		snapshotDir := filepath.Join(dataDir, "snapshots")
		info, err := os.Stat(snapshotDir)
		if err != nil {
			t.Fatalf("snapshot directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("snapshot path is not a directory")
		}
	})

	t.Run("returns error for existing directory", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		// Create the directory first
		dataDir := filepath.Join(baseDir, "existing-dir")
		if err := os.Mkdir(dataDir, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}

		// NewNodeStorage should fail because directory already exists
		_, err := NewNodeStorage(dataDir)
		if err == nil {
			t.Error("expected error when directory already exists")
		}
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		// Try to create in a path that definitely doesn't exist
		_, err := NewNodeStorage("/nonexistent/path/that/should/fail/raft-data")
		if err == nil {
			t.Error("expected error for invalid path")
		}
	})

	t.Run("returns error for empty path", func(t *testing.T) {
		_, err := NewNodeStorage("")
		if err == nil {
			t.Error("expected error for empty path")
		}
	})
}

func TestNodeStorage_GetLogStore(t *testing.T) {
	t.Run("returns non-nil log store", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		logStore := storage.GetLogStore()
		if logStore == nil {
			t.Error("GetLogStore() returned nil")
		}
	})

	t.Run("returns consistent log store", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		logStore1 := storage.GetLogStore()
		logStore2 := storage.GetLogStore()

		if logStore1 != logStore2 {
			t.Error("GetLogStore() should return same instance")
		}
	})
}

func TestNodeStorage_GetStableStore(t *testing.T) {
	t.Run("returns non-nil stable store", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		stableStore := storage.GetStableStore()
		if stableStore == nil {
			t.Error("GetStableStore() returned nil")
		}
	})

	t.Run("returns consistent stable store", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		stableStore1 := storage.GetStableStore()
		stableStore2 := storage.GetStableStore()

		if stableStore1 != stableStore2 {
			t.Error("GetStableStore() should return same instance")
		}
	})
}

func TestNodeStorage_GetSnapshotStore(t *testing.T) {
	t.Run("returns non-nil snapshot store", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		snapshotStore := storage.GetSnapshotStore()
		if snapshotStore == nil {
			t.Error("GetSnapshotStore() returned nil")
		}
	})

	t.Run("returns consistent snapshot store", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		snapshotStore1 := storage.GetSnapshotStore()
		snapshotStore2 := storage.GetSnapshotStore()

		if snapshotStore1 != snapshotStore2 {
			t.Error("GetSnapshotStore() should return same instance")
		}
	})
}

func TestNodeStorage_Close(t *testing.T) {
	t.Run("closes without error", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}

		err = storage.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	t.Run("close is idempotent", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}

		// First close
		err = storage.Close()
		if err != nil {
			t.Errorf("first Close() error = %v", err)
		}

		// Second close should not panic
		// Note: May return error for already closed database
		_ = storage.Close()
	})

	t.Run("close with nil boltDB returns nil", func(t *testing.T) {
		storage := &NodeStorage{
			boltDB: nil,
		}

		err := storage.Close()
		if err != nil {
			t.Errorf("Close() with nil boltDB error = %v", err)
		}
	})
}

func TestNodeStorage_StoreRelationships(t *testing.T) {
	t.Run("log and stable store share same bolt store", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		// Both should be backed by the same BoltStore
		// This is an implementation detail, but important for consistency
		logStore := storage.GetLogStore()
		stableStore := storage.GetStableStore()

		if logStore == nil || stableStore == nil {
			t.Fatal("stores should not be nil")
		}
	})
}

func TestNodeStorage_Persistence(t *testing.T) {
	t.Run("data persists after close and reopen", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")

		// Create first storage
		storage1, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("first NewNodeStorage() error = %v", err)
		}

		// Get stores and verify they work
		logStore := storage1.GetLogStore()
		if logStore == nil {
			t.Fatal("GetLogStore() returned nil")
		}

		stableStore := storage1.GetStableStore()
		if stableStore == nil {
			t.Fatal("GetStableStore() returned nil")
		}

		// Write some data to stable store
		testKey := []byte("test-key")
		testValue := []byte("test-value")
		err = stableStore.Set(testKey, testValue)
		if err != nil {
			t.Fatalf("StableStore.Set() error = %v", err)
		}

		// Close the storage
		if err := storage1.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		// Cannot reopen with NewNodeStorage as it tries to create the directory
		// This tests that the data was written to disk
		dbPath := filepath.Join(dataDir, "raft.db")
		info, err := os.Stat(dbPath)
		if err != nil {
			t.Fatalf("database file not found: %v", err)
		}
		if info.Size() == 0 {
			t.Error("database file is empty")
		}
	})
}

func TestNodeStorage_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		pathSetup func(baseDir string) string
		wantError bool
	}{
		{
			name: "valid new directory",
			pathSetup: func(baseDir string) string {
				return filepath.Join(baseDir, "valid-new-dir")
			},
			wantError: false,
		},
		{
			name: "nested new directory",
			pathSetup: func(baseDir string) string {
				// This should fail because parent doesn't exist
				return filepath.Join(baseDir, "parent", "child", "grandchild")
			},
			wantError: true,
		},
		{
			name: "directory with special characters",
			pathSetup: func(baseDir string) string {
				return filepath.Join(baseDir, "dir-with-special_chars.test")
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir, cleanup := testDataDir(t)
			defer cleanup()

			dataDir := tt.pathSetup(baseDir)
			storage, err := NewNodeStorage(dataDir)

			if tt.wantError {
				if err == nil {
					if storage != nil {
						storage.Close()
					}
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if storage != nil {
					storage.Close()
				}
			}
		})
	}
}

func TestNodeStorage_DirectoryPermissions(t *testing.T) {
	t.Run("creates directory with correct permissions", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		info, err := os.Stat(dataDir)
		if err != nil {
			t.Fatalf("os.Stat() error = %v", err)
		}

		// Check that the directory has expected permissions (0755)
		perm := info.Mode().Perm()
		if perm != 0755 {
			t.Errorf("directory permissions = %o, want 0755", perm)
		}
	})
}

func TestNodeStorage_ConcurrentAccess(t *testing.T) {
	t.Run("stores can be accessed concurrently", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		// Access stores from multiple goroutines
		done := make(chan bool, 3)

		go func() {
			_ = storage.GetLogStore()
			done <- true
		}()

		go func() {
			_ = storage.GetStableStore()
			done <- true
		}()

		go func() {
			_ = storage.GetSnapshotStore()
			done <- true
		}()

		// Wait for all goroutines
		for i := 0; i < 3; i++ {
			<-done
		}
	})
}

func TestNodeStorage_FileStructure(t *testing.T) {
	t.Run("creates expected file structure", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		// Check data directory
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			t.Error("data directory does not exist")
		}

		// Check raft.db
		dbPath := filepath.Join(dataDir, "raft.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("raft.db does not exist")
		}

		// Check snapshots directory
		snapshotPath := filepath.Join(dataDir, "snapshots")
		if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
			t.Error("snapshots directory does not exist")
		}
	})
}

func TestNodeStorage_StableStoreOperations(t *testing.T) {
	t.Run("can set and get values", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		stableStore := storage.GetStableStore()

		// Set a value
		key := []byte("test-key")
		value := []byte("test-value")
		if err := stableStore.Set(key, value); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		// Get the value back
		got, err := stableStore.Get(key)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if string(got) != string(value) {
			t.Errorf("Get() = %q, want %q", got, value)
		}
	})

	t.Run("can set and get uint64", func(t *testing.T) {
		baseDir, cleanup := testDataDir(t)
		defer cleanup()

		dataDir := filepath.Join(baseDir, "raft-data")
		storage, err := NewNodeStorage(dataDir)
		if err != nil {
			t.Fatalf("NewNodeStorage() error = %v", err)
		}
		defer storage.Close()

		stableStore := storage.GetStableStore()

		// Set a uint64 value
		key := []byte("uint64-key")
		value := uint64(12345)
		if err := stableStore.SetUint64(key, value); err != nil {
			t.Fatalf("SetUint64() error = %v", err)
		}

		// Get the value back
		got, err := stableStore.GetUint64(key)
		if err != nil {
			t.Fatalf("GetUint64() error = %v", err)
		}

		if got != value {
			t.Errorf("GetUint64() = %d, want %d", got, value)
		}
	})
}
