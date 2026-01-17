package node

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDefaultNodeConfig(t *testing.T) {
	t.Run("returns non-nil config", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		if cfg == nil {
			t.Fatal("DefaultNodeConfig returned nil")
		}
	})

	t.Run("has correct default heartbeat timeout", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		expected := time.Second * 1
		if cfg.HeartbeatTimeout != expected {
			t.Errorf("HeartbeatTimeout = %v, want %v", cfg.HeartbeatTimeout, expected)
		}
	})

	t.Run("has correct default election timeout", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		expected := time.Second * 1
		if cfg.ElectionTimeout != expected {
			t.Errorf("ElectionTimeout = %v, want %v", cfg.ElectionTimeout, expected)
		}
	})

	t.Run("has correct default leader lease timeout", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		expected := time.Second * 1
		if cfg.LeaderLeaseTimeout != expected {
			t.Errorf("LeaderLeaseTimeout = %v, want %v", cfg.LeaderLeaseTimeout, expected)
		}
	})

	t.Run("has correct default commit timeout", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		expected := time.Second * 1
		if cfg.CommitTimeout != expected {
			t.Errorf("CommitTimeout = %v, want %v", cfg.CommitTimeout, expected)
		}
	})

	t.Run("has correct default snapshot interval", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		expected := time.Second * 180
		if cfg.SnapshotInterval != expected {
			t.Errorf("SnapshotInterval = %v, want %v", cfg.SnapshotInterval, expected)
		}
	})

	t.Run("has correct default snapshot threshold", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		expected := 8192
		if cfg.SnapshotThreshold != expected {
			t.Errorf("SnapshotThreshold = %d, want %d", cfg.SnapshotThreshold, expected)
		}
	})

	t.Run("has empty ID by default", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		if cfg.ID != "" {
			t.Errorf("ID = %q, want empty string", cfg.ID)
		}
	})

	t.Run("has empty Addr by default", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		if cfg.Addr != "" {
			t.Errorf("Addr = %q, want empty string", cfg.Addr)
		}
	})

	t.Run("has empty DataDir by default", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		if cfg.DataDir != "" {
			t.Errorf("DataDir = %q, want empty string", cfg.DataDir)
		}
	})

	t.Run("has false Leader by default", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		if cfg.Leader != false {
			t.Errorf("Leader = %v, want false", cfg.Leader)
		}
	})
}

func TestNodeConfig_ZeroValue(t *testing.T) {
	var cfg NodeConfig

	t.Run("zero value has empty ID", func(t *testing.T) {
		if cfg.ID != "" {
			t.Errorf("ID = %q, want empty string", cfg.ID)
		}
	})

	t.Run("zero value has empty Addr", func(t *testing.T) {
		if cfg.Addr != "" {
			t.Errorf("Addr = %q, want empty string", cfg.Addr)
		}
	})

	t.Run("zero value has empty DataDir", func(t *testing.T) {
		if cfg.DataDir != "" {
			t.Errorf("DataDir = %q, want empty string", cfg.DataDir)
		}
	})

	t.Run("zero value has false Leader", func(t *testing.T) {
		if cfg.Leader != false {
			t.Errorf("Leader = %v, want false", cfg.Leader)
		}
	})

	t.Run("zero value has zero HeartbeatTimeout", func(t *testing.T) {
		if cfg.HeartbeatTimeout != 0 {
			t.Errorf("HeartbeatTimeout = %v, want 0", cfg.HeartbeatTimeout)
		}
	})

	t.Run("zero value has zero ElectionTimeout", func(t *testing.T) {
		if cfg.ElectionTimeout != 0 {
			t.Errorf("ElectionTimeout = %v, want 0", cfg.ElectionTimeout)
		}
	})

	t.Run("zero value has zero LeaderLeaseTimeout", func(t *testing.T) {
		if cfg.LeaderLeaseTimeout != 0 {
			t.Errorf("LeaderLeaseTimeout = %v, want 0", cfg.LeaderLeaseTimeout)
		}
	})

	t.Run("zero value has zero CommitTimeout", func(t *testing.T) {
		if cfg.CommitTimeout != 0 {
			t.Errorf("CommitTimeout = %v, want 0", cfg.CommitTimeout)
		}
	})

	t.Run("zero value has zero SnapshotInterval", func(t *testing.T) {
		if cfg.SnapshotInterval != 0 {
			t.Errorf("SnapshotInterval = %v, want 0", cfg.SnapshotInterval)
		}
	})

	t.Run("zero value has zero SnapshotThreshold", func(t *testing.T) {
		if cfg.SnapshotThreshold != 0 {
			t.Errorf("SnapshotThreshold = %d, want 0", cfg.SnapshotThreshold)
		}
	})
}

func TestNodeConfig_Fields(t *testing.T) {
	tests := []struct {
		name              string
		id                string
		addr              string
		dataDir           string
		leader            bool
		heartbeatTimeout  time.Duration
		electionTimeout   time.Duration
		leaderLeaseTimeout time.Duration
		commitTimeout     time.Duration
		snapshotInterval  time.Duration
		snapshotThreshold int
	}{
		{
			name:              "basic config",
			id:                "node-1",
			addr:              "127.0.0.1:9000",
			dataDir:           "/var/data/raft",
			leader:            true,
			heartbeatTimeout:  500 * time.Millisecond,
			electionTimeout:   time.Second,
			leaderLeaseTimeout: 500 * time.Millisecond,
			commitTimeout:     100 * time.Millisecond,
			snapshotInterval:  5 * time.Minute,
			snapshotThreshold: 1024,
		},
		{
			name:              "empty strings",
			id:                "",
			addr:              "",
			dataDir:           "",
			leader:            false,
			heartbeatTimeout:  0,
			electionTimeout:   0,
			leaderLeaseTimeout: 0,
			commitTimeout:     0,
			snapshotInterval:  0,
			snapshotThreshold: 0,
		},
		{
			name:              "high values",
			id:                "node-high",
			addr:              "192.168.1.100:9999",
			dataDir:           "/very/long/path/to/data/directory",
			leader:            true,
			heartbeatTimeout:  time.Hour,
			electionTimeout:   time.Hour,
			leaderLeaseTimeout: time.Hour,
			commitTimeout:     time.Hour,
			snapshotInterval:  24 * time.Hour,
			snapshotThreshold: 1000000,
		},
		{
			name:              "IPv6 address",
			id:                "node-ipv6",
			addr:              "[::1]:9000",
			dataDir:           "/data",
			leader:            false,
			heartbeatTimeout:  time.Second,
			electionTimeout:   time.Second,
			leaderLeaseTimeout: time.Second,
			commitTimeout:     time.Second,
			snapshotInterval:  time.Minute,
			snapshotThreshold: 100,
		},
		{
			name:              "hostname address",
			id:                "node-hostname",
			addr:              "localhost:9000",
			dataDir:           "./data",
			leader:            true,
			heartbeatTimeout:  2 * time.Second,
			electionTimeout:   2 * time.Second,
			leaderLeaseTimeout: 2 * time.Second,
			commitTimeout:     time.Second,
			snapshotInterval:  10 * time.Minute,
			snapshotThreshold: 4096,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &NodeConfig{
				ID:                 tt.id,
				Addr:               tt.addr,
				DataDir:            tt.dataDir,
				Leader:             tt.leader,
				HeartbeatTimeout:   tt.heartbeatTimeout,
				ElectionTimeout:    tt.electionTimeout,
				LeaderLeaseTimeout: tt.leaderLeaseTimeout,
				CommitTimeout:      tt.commitTimeout,
				SnapshotInterval:   tt.snapshotInterval,
				SnapshotThreshold:  tt.snapshotThreshold,
			}

			if cfg.ID != tt.id {
				t.Errorf("ID = %q, want %q", cfg.ID, tt.id)
			}
			if cfg.Addr != tt.addr {
				t.Errorf("Addr = %q, want %q", cfg.Addr, tt.addr)
			}
			if cfg.DataDir != tt.dataDir {
				t.Errorf("DataDir = %q, want %q", cfg.DataDir, tt.dataDir)
			}
			if cfg.Leader != tt.leader {
				t.Errorf("Leader = %v, want %v", cfg.Leader, tt.leader)
			}
			if cfg.HeartbeatTimeout != tt.heartbeatTimeout {
				t.Errorf("HeartbeatTimeout = %v, want %v", cfg.HeartbeatTimeout, tt.heartbeatTimeout)
			}
			if cfg.ElectionTimeout != tt.electionTimeout {
				t.Errorf("ElectionTimeout = %v, want %v", cfg.ElectionTimeout, tt.electionTimeout)
			}
			if cfg.LeaderLeaseTimeout != tt.leaderLeaseTimeout {
				t.Errorf("LeaderLeaseTimeout = %v, want %v", cfg.LeaderLeaseTimeout, tt.leaderLeaseTimeout)
			}
			if cfg.CommitTimeout != tt.commitTimeout {
				t.Errorf("CommitTimeout = %v, want %v", cfg.CommitTimeout, tt.commitTimeout)
			}
			if cfg.SnapshotInterval != tt.snapshotInterval {
				t.Errorf("SnapshotInterval = %v, want %v", cfg.SnapshotInterval, tt.snapshotInterval)
			}
			if cfg.SnapshotThreshold != tt.snapshotThreshold {
				t.Errorf("SnapshotThreshold = %d, want %d", cfg.SnapshotThreshold, tt.snapshotThreshold)
			}
		})
	}
}

func TestNodeConfig_JSONSerialization(t *testing.T) {
	t.Run("marshal and unmarshal preserves all fields", func(t *testing.T) {
		original := &NodeConfig{
			ID:                 "node-1",
			Addr:               "127.0.0.1:9000",
			DataDir:            "/var/data/raft",
			Leader:             true,
			HeartbeatTimeout:   500 * time.Millisecond,
			ElectionTimeout:    time.Second,
			LeaderLeaseTimeout: 500 * time.Millisecond,
			CommitTimeout:      100 * time.Millisecond,
			SnapshotInterval:   5 * time.Minute,
			SnapshotThreshold:  1024,
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}

		var restored NodeConfig
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if restored.ID != original.ID {
			t.Errorf("ID = %q, want %q", restored.ID, original.ID)
		}
		if restored.Addr != original.Addr {
			t.Errorf("Addr = %q, want %q", restored.Addr, original.Addr)
		}
		if restored.DataDir != original.DataDir {
			t.Errorf("DataDir = %q, want %q", restored.DataDir, original.DataDir)
		}
		if restored.Leader != original.Leader {
			t.Errorf("Leader = %v, want %v", restored.Leader, original.Leader)
		}
		if restored.HeartbeatTimeout != original.HeartbeatTimeout {
			t.Errorf("HeartbeatTimeout = %v, want %v", restored.HeartbeatTimeout, original.HeartbeatTimeout)
		}
		if restored.ElectionTimeout != original.ElectionTimeout {
			t.Errorf("ElectionTimeout = %v, want %v", restored.ElectionTimeout, original.ElectionTimeout)
		}
		if restored.LeaderLeaseTimeout != original.LeaderLeaseTimeout {
			t.Errorf("LeaderLeaseTimeout = %v, want %v", restored.LeaderLeaseTimeout, original.LeaderLeaseTimeout)
		}
		if restored.CommitTimeout != original.CommitTimeout {
			t.Errorf("CommitTimeout = %v, want %v", restored.CommitTimeout, original.CommitTimeout)
		}
		if restored.SnapshotInterval != original.SnapshotInterval {
			t.Errorf("SnapshotInterval = %v, want %v", restored.SnapshotInterval, original.SnapshotInterval)
		}
		if restored.SnapshotThreshold != original.SnapshotThreshold {
			t.Errorf("SnapshotThreshold = %d, want %d", restored.SnapshotThreshold, original.SnapshotThreshold)
		}
	})

	t.Run("marshal default config", func(t *testing.T) {
		cfg := DefaultNodeConfig()
		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		if len(data) == 0 {
			t.Error("json.Marshal() returned empty data")
		}
	})

	t.Run("unmarshal from JSON string", func(t *testing.T) {
		jsonStr := `{
			"id": "test-node",
			"addr": "localhost:8080",
			"dataDir": "/tmp/data",
			"leader": true,
			"heartbeatTimeout": 1000000000,
			"electionTimeout": 2000000000,
			"leaderLeaseTimeout": 500000000,
			"commitTimeout": 100000000,
			"snapshotInterval": 300000000000,
			"snapshotThreshold": 2048
		}`

		var cfg NodeConfig
		if err := json.Unmarshal([]byte(jsonStr), &cfg); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if cfg.ID != "test-node" {
			t.Errorf("ID = %q, want %q", cfg.ID, "test-node")
		}
		if cfg.Addr != "localhost:8080" {
			t.Errorf("Addr = %q, want %q", cfg.Addr, "localhost:8080")
		}
		if cfg.DataDir != "/tmp/data" {
			t.Errorf("DataDir = %q, want %q", cfg.DataDir, "/tmp/data")
		}
		if cfg.Leader != true {
			t.Errorf("Leader = %v, want true", cfg.Leader)
		}
		if cfg.HeartbeatTimeout != time.Second {
			t.Errorf("HeartbeatTimeout = %v, want %v", cfg.HeartbeatTimeout, time.Second)
		}
		if cfg.ElectionTimeout != 2*time.Second {
			t.Errorf("ElectionTimeout = %v, want %v", cfg.ElectionTimeout, 2*time.Second)
		}
		if cfg.SnapshotThreshold != 2048 {
			t.Errorf("SnapshotThreshold = %d, want 2048", cfg.SnapshotThreshold)
		}
	})
}

func TestNodeConfig_Pointer(t *testing.T) {
	t.Run("DefaultNodeConfig returns pointer", func(t *testing.T) {
		cfg := DefaultNodeConfig()

		// Modify through pointer
		cfg.ID = "modified-id"
		cfg.Leader = true

		if cfg.ID != "modified-id" {
			t.Errorf("ID = %q, want %q", cfg.ID, "modified-id")
		}
		if cfg.Leader != true {
			t.Errorf("Leader = %v, want true", cfg.Leader)
		}
	})

	t.Run("each call returns new instance", func(t *testing.T) {
		cfg1 := DefaultNodeConfig()
		cfg2 := DefaultNodeConfig()

		cfg1.ID = "node-1"
		cfg2.ID = "node-2"

		if cfg1.ID == cfg2.ID {
			t.Error("modifying one config should not affect the other")
		}
	})
}

func TestNodeConfig_JSONTags(t *testing.T) {
	t.Run("JSON field names are correct", func(t *testing.T) {
		cfg := &NodeConfig{
			ID:                 "test-id",
			Addr:               "test-addr",
			DataDir:            "test-dir",
			Leader:             true,
			HeartbeatTimeout:   time.Second,
			ElectionTimeout:    time.Second,
			LeaderLeaseTimeout: time.Second,
			CommitTimeout:      time.Second,
			SnapshotInterval:   time.Second,
			SnapshotThreshold:  100,
		}

		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}

		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		expectedFields := []string{
			"id", "addr", "dataDir", "leader",
			"heartbeatTimeout", "electionTimeout", "leaderLeaseTimeout",
			"commitTimeout", "snapshotInterval", "snapshotThreshold",
		}

		for _, field := range expectedFields {
			if _, ok := m[field]; !ok {
				t.Errorf("expected JSON field %q not found", field)
			}
		}
	})
}

func TestNodeConfig_DurationValues(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{name: "nanosecond", duration: time.Nanosecond},
		{name: "microsecond", duration: time.Microsecond},
		{name: "millisecond", duration: time.Millisecond},
		{name: "second", duration: time.Second},
		{name: "minute", duration: time.Minute},
		{name: "hour", duration: time.Hour},
		{name: "negative", duration: -time.Second},
		{name: "zero", duration: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &NodeConfig{
				HeartbeatTimeout:   tt.duration,
				ElectionTimeout:    tt.duration,
				LeaderLeaseTimeout: tt.duration,
				CommitTimeout:      tt.duration,
				SnapshotInterval:   tt.duration,
			}

			if cfg.HeartbeatTimeout != tt.duration {
				t.Errorf("HeartbeatTimeout = %v, want %v", cfg.HeartbeatTimeout, tt.duration)
			}
			if cfg.ElectionTimeout != tt.duration {
				t.Errorf("ElectionTimeout = %v, want %v", cfg.ElectionTimeout, tt.duration)
			}
			if cfg.LeaderLeaseTimeout != tt.duration {
				t.Errorf("LeaderLeaseTimeout = %v, want %v", cfg.LeaderLeaseTimeout, tt.duration)
			}
			if cfg.CommitTimeout != tt.duration {
				t.Errorf("CommitTimeout = %v, want %v", cfg.CommitTimeout, tt.duration)
			}
			if cfg.SnapshotInterval != tt.duration {
				t.Errorf("SnapshotInterval = %v, want %v", cfg.SnapshotInterval, tt.duration)
			}
		})
	}
}

func TestNodeConfig_SnapshotThresholdValues(t *testing.T) {
	tests := []struct {
		name      string
		threshold int
	}{
		{name: "zero", threshold: 0},
		{name: "one", threshold: 1},
		{name: "small", threshold: 100},
		{name: "default", threshold: 8192},
		{name: "large", threshold: 1000000},
		{name: "max int32", threshold: 2147483647},
		{name: "negative", threshold: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &NodeConfig{
				SnapshotThreshold: tt.threshold,
			}

			if cfg.SnapshotThreshold != tt.threshold {
				t.Errorf("SnapshotThreshold = %d, want %d", cfg.SnapshotThreshold, tt.threshold)
			}
		})
	}
}
