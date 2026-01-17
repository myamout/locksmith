package statemachine

import (
	"bytes"
	"errors"
	"testing"
)

func TestLockStateMachineSnapshot_Persist(t *testing.T) {
	t.Run("persist writes data to sink", func(t *testing.T) {
		data := []byte(`{"locks":{},"fenceToken":0}`)
		snapshot := &LockStateMachineSnapshot{data: data}

		var buf bytes.Buffer
		sink := &mockSnapshotSink{buf: &buf}

		err := snapshot.Persist(sink)
		if err != nil {
			t.Fatalf("Persist() error = %v", err)
		}

		if !bytes.Equal(buf.Bytes(), data) {
			t.Errorf("written data = %q, want %q", buf.String(), string(data))
		}
	})

	t.Run("persist closes sink on success", func(t *testing.T) {
		snapshot := &LockStateMachineSnapshot{data: []byte(`{}`)}

		var buf bytes.Buffer
		sink := &mockSnapshotSink{buf: &buf}

		err := snapshot.Persist(sink)
		if err != nil {
			t.Fatalf("Persist() error = %v", err)
		}

		if !sink.closed {
			t.Error("sink should be closed after successful persist")
		}
		if sink.cancelled {
			t.Error("sink should not be cancelled on success")
		}
	})

	t.Run("persist cancels sink on write error", func(t *testing.T) {
		snapshot := &LockStateMachineSnapshot{data: []byte(`{}`)}

		sink := &errorSnapshotSink{writeErr: errors.New("write error")}

		err := snapshot.Persist(sink)
		if err == nil {
			t.Fatal("expected error from Persist")
		}

		if !sink.cancelled {
			t.Error("sink should be cancelled on write error")
		}
		if sink.closed {
			t.Error("sink should not be closed on write error")
		}
	})

	t.Run("persist with empty data", func(t *testing.T) {
		snapshot := &LockStateMachineSnapshot{data: []byte{}}

		var buf bytes.Buffer
		sink := &mockSnapshotSink{buf: &buf}

		err := snapshot.Persist(sink)
		if err != nil {
			t.Fatalf("Persist() error = %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("expected empty buffer, got %d bytes", buf.Len())
		}
	})

	t.Run("persist with large data", func(t *testing.T) {
		// Create a large payload
		largeData := make([]byte, 1024*1024) // 1MB
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		snapshot := &LockStateMachineSnapshot{data: largeData}

		var buf bytes.Buffer
		sink := &mockSnapshotSink{buf: &buf}

		err := snapshot.Persist(sink)
		if err != nil {
			t.Fatalf("Persist() error = %v", err)
		}

		if !bytes.Equal(buf.Bytes(), largeData) {
			t.Error("large data not persisted correctly")
		}
	})

	t.Run("persist preserves JSON structure", func(t *testing.T) {
		data := []byte(`{"locks":{"key1":{"key":"key1","clientID":"client-1","fenceToken":1}},"fenceToken":1}`)
		snapshot := &LockStateMachineSnapshot{data: data}

		var buf bytes.Buffer
		sink := &mockSnapshotSink{buf: &buf}

		err := snapshot.Persist(sink)
		if err != nil {
			t.Fatalf("Persist() error = %v", err)
		}

		if !bytes.Equal(buf.Bytes(), data) {
			t.Errorf("JSON data not preserved: got %q, want %q", buf.String(), string(data))
		}
	})
}

func TestLockStateMachineSnapshot_Release(t *testing.T) {
	t.Run("release does not panic", func(t *testing.T) {
		snapshot := &LockStateMachineSnapshot{data: []byte(`{}`)}

		// Should not panic
		snapshot.Release()
	})

	t.Run("release can be called multiple times", func(t *testing.T) {
		snapshot := &LockStateMachineSnapshot{data: []byte(`{}`)}

		// Should not panic when called multiple times
		for range 10 {
			snapshot.Release()
		}
	})

	t.Run("release on nil data does not panic", func(t *testing.T) {
		snapshot := &LockStateMachineSnapshot{data: nil}

		// Should not panic
		snapshot.Release()
	})
}

func TestLockStateMachineSnapshot_Data(t *testing.T) {
	t.Run("data is preserved", func(t *testing.T) {
		testData := []byte(`{"test":"data"}`)
		snapshot := &LockStateMachineSnapshot{data: testData}

		if !bytes.Equal(snapshot.data, testData) {
			t.Errorf("data = %q, want %q", string(snapshot.data), string(testData))
		}
	})

	t.Run("nil data is allowed", func(t *testing.T) {
		snapshot := &LockStateMachineSnapshot{data: nil}

		if snapshot.data != nil {
			t.Errorf("data = %v, want nil", snapshot.data)
		}
	})
}

// errorSnapshotSink is a mock that returns errors.
type errorSnapshotSink struct {
	writeErr  error
	closeErr  error
	cancelled bool
	closed    bool
}

func (e *errorSnapshotSink) Write(p []byte) (n int, err error) {
	if e.writeErr != nil {
		return 0, e.writeErr
	}
	return len(p), nil
}

func (e *errorSnapshotSink) Close() error {
	e.closed = true
	return e.closeErr
}

func (e *errorSnapshotSink) ID() string {
	return "error-sink"
}

func (e *errorSnapshotSink) Cancel() error {
	e.cancelled = true
	return nil
}

// Benchmarks
func BenchmarkLockStateMachineSnapshot_Persist(b *testing.B) {
	data := []byte(`{"locks":{"key1":{"key":"key1","clientID":"client-1","fenceToken":1,"expiresAt":"2024-01-01T00:00:00Z","createdAt":"2024-01-01T00:00:00Z"}},"fenceToken":1}`)
	snapshot := &LockStateMachineSnapshot{data: data}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		sink := &mockSnapshotSink{buf: &buf}
		_ = snapshot.Persist(sink)
	}
}

func BenchmarkLockStateMachineSnapshot_Persist_Large(b *testing.B) {
	// Create snapshot with 1000 locks
	largeData := make([]byte, 100*1024) // 100KB
	for i := range largeData {
		largeData[i] = byte('a' + (i % 26))
	}
	snapshot := &LockStateMachineSnapshot{data: largeData}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		sink := &mockSnapshotSink{buf: &buf}
		_ = snapshot.Persist(sink)
	}
}
