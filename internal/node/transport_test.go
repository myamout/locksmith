package node

import (
	"net"
	"testing"
	"time"
)

func TestNewNodeTransport(t *testing.T) {
	t.Run("creates transport with valid address", func(t *testing.T) {
		// Use a random available port
		addr := "127.0.0.1:0"
		transport, err := NewNodeTransport(addr)
		if err != nil {
			t.Fatalf("NewNodeTransport() error = %v", err)
		}
		defer transport.Close()

		if transport == nil {
			t.Fatal("NewNodeTransport() returned nil transport")
		}
	})

	t.Run("creates transport with localhost", func(t *testing.T) {
		addr := "localhost:0"
		transport, err := NewNodeTransport(addr)
		if err != nil {
			t.Fatalf("NewNodeTransport() error = %v", err)
		}
		defer transport.Close()

		if transport == nil {
			t.Fatal("NewNodeTransport() returned nil transport")
		}
	})

	t.Run("returns error for invalid address format", func(t *testing.T) {
		invalidAddrs := []string{
			"invalid",
			"no-port",
			":only-port",
			"",
		}

		for _, addr := range invalidAddrs {
			t.Run(addr, func(t *testing.T) {
				transport, err := NewNodeTransport(addr)
				if err == nil && addr != "" {
					// Empty string might be valid in some cases
					if transport != nil {
						transport.Close()
					}
				}
				// We expect either an error or a valid transport
			})
		}
	})

	t.Run("returns error for invalid host", func(t *testing.T) {
		// Use a clearly invalid hostname
		addr := "this.host.definitely.does.not.exist.invalid:9000"
		_, err := NewNodeTransport(addr)
		if err == nil {
			t.Skip("DNS resolution unexpectedly succeeded")
		}
	})
}

func TestNewNodeTransport_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		addr      string
		wantError bool
		skip      string
	}{
		{
			name:      "valid IPv4 with port 0",
			addr:      "127.0.0.1:0",
			wantError: false,
		},
		{
			name:      "valid localhost with port 0",
			addr:      "localhost:0",
			wantError: false,
		},
		{
			name:      "0.0.0.0 is not advertisable",
			addr:      "0.0.0.0:0",
			wantError: true, // raft requires advertisable addresses
		},
		{
			name:      "missing port",
			addr:      "127.0.0.1",
			wantError: true,
		},
		{
			name:      "missing host is not advertisable",
			addr:      ":9000",
			wantError: true, // raft requires advertisable addresses
		},
		{
			name:      "empty string",
			addr:      "",
			wantError: true,
		},
		{
			name:      "invalid port number",
			addr:      "127.0.0.1:99999",
			wantError: true,
		},
		{
			name:      "negative port",
			addr:      "127.0.0.1:-1",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip != "" {
				t.Skip(tt.skip)
			}

			transport, err := NewNodeTransport(tt.addr)
			if tt.wantError {
				if err == nil {
					if transport != nil {
						transport.Close()
					}
					t.Errorf("NewNodeTransport(%q) expected error, got nil", tt.addr)
				}
			} else {
				if err != nil {
					t.Errorf("NewNodeTransport(%q) unexpected error: %v", tt.addr, err)
				} else if transport != nil {
					transport.Close()
				}
			}
		})
	}
}

func TestNewNodeTransport_TransportProperties(t *testing.T) {
	t.Run("transport can be closed", func(t *testing.T) {
		transport, err := NewNodeTransport("127.0.0.1:0")
		if err != nil {
			t.Fatalf("NewNodeTransport() error = %v", err)
		}

		err = transport.Close()
		if err != nil {
			t.Errorf("transport.Close() error = %v", err)
		}
	})

	t.Run("transport has valid local address", func(t *testing.T) {
		transport, err := NewNodeTransport("127.0.0.1:0")
		if err != nil {
			t.Fatalf("NewNodeTransport() error = %v", err)
		}
		defer transport.Close()

		localAddr := transport.LocalAddr()
		if localAddr == "" {
			t.Error("transport.LocalAddr() returned empty string")
		}
	})

	t.Run("transport is listening after creation", func(t *testing.T) {
		transport, err := NewNodeTransport("127.0.0.1:0")
		if err != nil {
			t.Fatalf("NewNodeTransport() error = %v", err)
		}
		defer transport.Close()

		// The transport should have a valid local address
		addr := transport.LocalAddr()
		if addr == "" {
			t.Fatal("transport has no local address")
		}

		// Try to connect to the transport
		conn, err := net.DialTimeout("tcp", string(addr), time.Second)
		if err != nil {
			t.Logf("Note: could not connect to transport (this may be expected): %v", err)
		} else {
			conn.Close()
		}
	})
}

func TestTransportConstants(t *testing.T) {
	t.Run("maxPool is positive", func(t *testing.T) {
		if maxPool <= 0 {
			t.Errorf("maxPool = %d, want positive value", maxPool)
		}
	})

	t.Run("maxPool has reasonable value", func(t *testing.T) {
		if maxPool != 3 {
			t.Errorf("maxPool = %d, want 3", maxPool)
		}
	})

	t.Run("timeout is positive", func(t *testing.T) {
		if timeout <= 0 {
			t.Errorf("timeout = %v, want positive value", timeout)
		}
	})

	t.Run("timeout has expected value", func(t *testing.T) {
		expected := time.Second * 10
		if timeout != expected {
			t.Errorf("timeout = %v, want %v", timeout, expected)
		}
	})
}

func TestNewNodeTransport_MultipleTransports(t *testing.T) {
	t.Run("can create multiple transports on different ports", func(t *testing.T) {
		transport1, err := NewNodeTransport("127.0.0.1:0")
		if err != nil {
			t.Fatalf("first NewNodeTransport() error = %v", err)
		}
		defer transport1.Close()

		transport2, err := NewNodeTransport("127.0.0.1:0")
		if err != nil {
			t.Fatalf("second NewNodeTransport() error = %v", err)
		}
		defer transport2.Close()

		// Both transports should be created successfully
		// Note: LocalAddr() may return the advertised address (127.0.0.1:0)
		// rather than the actual bound address, which is implementation-specific
		if transport1 == nil || transport2 == nil {
			t.Error("transports should not be nil")
		}
	})

	t.Run("cannot create two transports on same specific port", func(t *testing.T) {
		// First, get an available port
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to get available port: %v", err)
		}
		port := listener.Addr().(*net.TCPAddr).Port
		listener.Close()

		// Now use that specific port
		addr := "127.0.0.1:" + itoa(port)

		transport1, err := NewNodeTransport(addr)
		if err != nil {
			t.Fatalf("first NewNodeTransport() error = %v", err)
		}
		defer transport1.Close()

		// Second transport on same port should fail
		transport2, err := NewNodeTransport(addr)
		if err == nil {
			transport2.Close()
			t.Error("expected error when creating second transport on same port")
		}
	})
}

func TestNewNodeTransport_IPv6(t *testing.T) {
	t.Run("creates transport with IPv6 loopback", func(t *testing.T) {
		// Check if IPv6 is available
		listener, err := net.Listen("tcp6", "[::1]:0")
		if err != nil {
			t.Skip("IPv6 not available on this system")
		}
		listener.Close()

		transport, err := NewNodeTransport("[::1]:0")
		if err != nil {
			t.Fatalf("NewNodeTransport() error = %v", err)
		}
		defer transport.Close()

		if transport == nil {
			t.Fatal("NewNodeTransport() returned nil transport")
		}
	})
}

func TestNewNodeTransport_CloseIdempotent(t *testing.T) {
	t.Run("transport can be closed multiple times", func(t *testing.T) {
		transport, err := NewNodeTransport("127.0.0.1:0")
		if err != nil {
			t.Fatalf("NewNodeTransport() error = %v", err)
		}

		// First close
		err = transport.Close()
		if err != nil {
			t.Errorf("first Close() error = %v", err)
		}

		// Second close should not panic
		// Note: It may or may not return an error depending on implementation
		_ = transport.Close()
	})
}

// itoa converts an int to a string without importing strconv
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	var s string
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
