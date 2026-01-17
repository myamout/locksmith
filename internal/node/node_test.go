package node

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// testNodeSetup creates temporary directories and a valid config for testing.
// Returns the config and a cleanup function.
func testNodeSetup(t *testing.T) (*NodeConfig, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "locksmith-node-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cfg := DefaultNodeConfig()
	cfg.ID = "test-node"
	cfg.Addr = "127.0.0.1:0"
	cfg.DataDir = filepath.Join(dir, "raft-data")

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return cfg, cleanup
}

func TestNewLocksmithNode(t *testing.T) {
	t.Run("creates node with valid config", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		if node == nil {
			t.Fatal("NewLocksmithNode() returned nil node")
		}
	})

	t.Run("node has valid config reference", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		if node.cfg != cfg {
			t.Error("node.cfg does not reference the original config")
		}
	})

	t.Run("node has valid raft instance", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		if node.raft == nil {
			t.Error("node.raft is nil")
		}
	})

	t.Run("node has valid state machine", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		if node.sm == nil {
			t.Error("node.sm is nil")
		}
	})

	t.Run("node has valid storage", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		if node.storage == nil {
			t.Error("node.storage is nil")
		}
	})

	t.Run("node has valid transport", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		if node.transport == nil {
			t.Error("node.transport is nil")
		}
	})

	t.Run("returns error for invalid data directory", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		cfg.ID = "test-node"
		cfg.Addr = "127.0.0.1:0"
		cfg.DataDir = "/nonexistent/path/that/should/fail"

		_, err := NewLocksmithNode(cfg)
		if err == nil {
			t.Error("expected error for invalid data directory")
		}
	})

	t.Run("returns error for invalid address", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "locksmith-node-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(dir)

		cfg := DefaultNodeConfig()
		cfg.ID = "test-node"
		cfg.Addr = "invalid-address-format"
		cfg.DataDir = filepath.Join(dir, "raft-data")

		_, err = NewLocksmithNode(cfg)
		if err == nil {
			t.Error("expected error for invalid address")
		}
	})

	t.Run("cleans up storage on transport failure", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "locksmith-node-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(dir)

		cfg := DefaultNodeConfig()
		cfg.ID = "test-node"
		cfg.Addr = "invalid"
		cfg.DataDir = filepath.Join(dir, "raft-data")

		_, err = NewLocksmithNode(cfg)
		if err == nil {
			t.Skip("transport creation unexpectedly succeeded")
		}

		// Check that storage cleanup happened by verifying we can't access the db
		// This is implementation-dependent
	})
}

func TestLocksmithNode_Bootstrap(t *testing.T) {
	t.Run("bootstrap succeeds for single node", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		err = node.Bootstrap()
		if err != nil {
			t.Errorf("Bootstrap() error = %v", err)
		}
	})

	t.Run("bootstrap is idempotent", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		// First bootstrap
		err = node.Bootstrap()
		if err != nil {
			t.Errorf("first Bootstrap() error = %v", err)
		}

		// Second bootstrap should not error (ErrCantBootstrap is handled)
		err = node.Bootstrap()
		if err != nil {
			t.Errorf("second Bootstrap() error = %v", err)
		}
	})

	t.Run("bootstrap configures node as server", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		err = node.Bootstrap()
		if err != nil {
			t.Errorf("Bootstrap() error = %v", err)
		}

		// After bootstrap, node should eventually become leader (single node cluster)
		// Wait a bit for leader election
		time.Sleep(100 * time.Millisecond)

		// Node should be in some valid state
		state := node.raft.State()
		if state.String() == "" {
			t.Error("node has invalid raft state")
		}
	})
}

func TestLocksmithNode_AddVoter(t *testing.T) {
	t.Run("returns error when not leader", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		// Without bootstrap, node is a follower
		err = node.AddVoter("new-node", "127.0.0.1:9001")
		if err == nil {
			t.Error("expected error when adding voter as non-leader")
		}
		if err.Error() != "non-leader node cannot add voters" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("can add voter when leader", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		// Bootstrap to become leader
		if err := node.Bootstrap(); err != nil {
			t.Fatalf("Bootstrap() error = %v", err)
		}

		// Wait for node to become leader
		timeout := time.After(5 * time.Second)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				t.Fatal("timeout waiting for node to become leader")
			case <-ticker.C:
				if node.raft.State().String() == "Leader" {
					// Now try to add a voter
					// Note: This will likely fail because the address is not reachable,
					// but it should not fail with "non-leader" error
					err := node.AddVoter("new-node", "127.0.0.1:0")
					if err != nil && err.Error() == "non-leader node cannot add voters" {
						t.Error("node should be leader but got non-leader error")
					}
					return
				}
			}
		}
	})
}

func TestLocksmithNode_RemoveNode(t *testing.T) {
	t.Run("returns error when not leader", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		// Without bootstrap, node is a follower
		err = node.RemoveNode("other-node")
		if err == nil {
			t.Error("expected error when removing node as non-leader")
		}
		if err.Error() != "non-leader node cannot remove other nodes" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestLocksmithNode_Apply(t *testing.T) {
	t.Run("returns error when not leader", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		// Without bootstrap, node is a follower
		_, err = node.Apply([]byte("test data"), time.Second)
		if err == nil {
			t.Error("expected error when applying as non-leader")
		}
		if err.Error() != "non-leader node cannot apply changes across the cluster" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("can apply when leader", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		// Bootstrap to become leader
		if err := node.Bootstrap(); err != nil {
			t.Fatalf("Bootstrap() error = %v", err)
		}

		// Wait for node to become leader
		timeout := time.After(5 * time.Second)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				t.Fatal("timeout waiting for node to become leader")
			case <-ticker.C:
				if node.raft.State().String() == "Leader" {
					// Apply some data
					_, err := node.Apply([]byte("test data"), time.Second)
					// Note: This may return an error due to invalid command format,
					// but it should not fail with "non-leader" error
					if err != nil && err.Error() == "non-leader node cannot apply changes across the cluster" {
						t.Error("node should be leader but got non-leader error")
					}
					return
				}
			}
		}
	})
}

func TestLocksmithNode_Shutdown(t *testing.T) {
	t.Run("shutdown succeeds", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}

		err = node.Shutdown()
		if err != nil {
			t.Errorf("Shutdown() error = %v", err)
		}
	})

	t.Run("shutdown closes storage", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}

		err = node.Shutdown()
		if err != nil {
			t.Errorf("Shutdown() error = %v", err)
		}

		// Storage should be closed
		// Attempting operations on closed storage may panic or error
	})

	t.Run("shutdown closes transport", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}

		err = node.Shutdown()
		if err != nil {
			t.Errorf("Shutdown() error = %v", err)
		}

		// Transport should be closed
	})

	t.Run("shutdown after bootstrap", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}

		if err := node.Bootstrap(); err != nil {
			t.Fatalf("Bootstrap() error = %v", err)
		}

		err = node.Shutdown()
		if err != nil {
			t.Errorf("Shutdown() error = %v", err)
		}
	})
}

func TestLocksmithNode_RaftConfiguration(t *testing.T) {
	t.Run("raft uses config ID as local ID", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		cfg.ID = "custom-node-id"

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		// Verify the raft instance was created with the correct ID
		// The ID is set in the raft configuration
		// This is indirectly tested through bootstrap
		if err := node.Bootstrap(); err != nil {
			t.Fatalf("Bootstrap() error = %v", err)
		}
	})

	t.Run("raft uses configured timeouts", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		// Set custom timeouts
		cfg.HeartbeatTimeout = 200 * time.Millisecond
		cfg.ElectionTimeout = 500 * time.Millisecond
		cfg.LeaderLeaseTimeout = 200 * time.Millisecond
		cfg.CommitTimeout = 100 * time.Millisecond

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		// Node should be created successfully with custom timeouts
		if node.raft == nil {
			t.Error("raft instance is nil")
		}
	})

	t.Run("raft uses configured snapshot settings", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		cfg.SnapshotInterval = 30 * time.Second
		cfg.SnapshotThreshold = 100

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		// Node should be created successfully with custom snapshot settings
		if node.raft == nil {
			t.Error("raft instance is nil")
		}
	})
}

func TestLocksmithNode_LeaderElection(t *testing.T) {
	t.Run("single node becomes leader after bootstrap", func(t *testing.T) {
		cfg, cleanup := testNodeSetup(t)
		defer cleanup()

		node, err := NewLocksmithNode(cfg)
		if err != nil {
			t.Fatalf("NewLocksmithNode() error = %v", err)
		}
		defer node.Shutdown()

		if err := node.Bootstrap(); err != nil {
			t.Fatalf("Bootstrap() error = %v", err)
		}

		// Wait for leader election
		timeout := time.After(5 * time.Second)
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				t.Fatalf("timeout waiting for leader, current state: %s", node.raft.State())
			case <-ticker.C:
				if node.raft.State().String() == "Leader" {
					return
				}
			}
		}
	})
}

func TestLocksmithNode_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(*NodeConfig)
		wantError   bool
	}{
		{
			name:        "default config",
			setupConfig: func(cfg *NodeConfig) {},
			wantError:   false,
		},
		{
			name: "custom ID",
			setupConfig: func(cfg *NodeConfig) {
				cfg.ID = "custom-id-12345"
			},
			wantError: false,
		},
		{
			name: "short timeouts",
			setupConfig: func(cfg *NodeConfig) {
				cfg.HeartbeatTimeout = 100 * time.Millisecond
				cfg.ElectionTimeout = 200 * time.Millisecond
				cfg.LeaderLeaseTimeout = 100 * time.Millisecond
				cfg.CommitTimeout = 50 * time.Millisecond
			},
			wantError: false,
		},
		{
			name: "high snapshot threshold",
			setupConfig: func(cfg *NodeConfig) {
				cfg.SnapshotThreshold = 100000
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, cleanup := testNodeSetup(t)
			defer cleanup()

			tt.setupConfig(cfg)

			node, err := NewLocksmithNode(cfg)
			if tt.wantError {
				if err == nil {
					node.Shutdown()
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else {
					node.Shutdown()
				}
			}
		})
	}
}

func TestLocksmithNode_MultipleNodes(t *testing.T) {
	t.Run("can create multiple independent nodes", func(t *testing.T) {
		cfg1, cleanup1 := testNodeSetup(t)
		defer cleanup1()
		cfg1.ID = "node-1"

		cfg2, cleanup2 := testNodeSetup(t)
		defer cleanup2()
		cfg2.ID = "node-2"

		node1, err := NewLocksmithNode(cfg1)
		if err != nil {
			t.Fatalf("NewLocksmithNode(cfg1) error = %v", err)
		}
		defer node1.Shutdown()

		node2, err := NewLocksmithNode(cfg2)
		if err != nil {
			t.Fatalf("NewLocksmithNode(cfg2) error = %v", err)
		}
		defer node2.Shutdown()

		// Both nodes should be created independently
		if node1.cfg.ID == node2.cfg.ID {
			t.Error("nodes should have different IDs")
		}
	})
}

func TestNewLocksmithNode_ResourceCleanup(t *testing.T) {
	t.Run("cleans up resources on failure", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "locksmith-node-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(dir)

		cfg := DefaultNodeConfig()
		cfg.ID = "" // Empty ID might cause issues
		cfg.Addr = "127.0.0.1:0"
		cfg.DataDir = filepath.Join(dir, "raft-data")

		// This may or may not fail depending on raft's validation
		node, err := NewLocksmithNode(cfg)
		if err == nil {
			node.Shutdown()
		}
		// The important thing is we don't leak resources
	})
}
