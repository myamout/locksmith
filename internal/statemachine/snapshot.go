package statemachine

import (
	"github.com/hashicorp/raft"
)

type LockStateMachineSnapshot struct {
	data []byte
}

func (ls *LockStateMachineSnapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := sink.Write(ls.data); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (ls *LockStateMachineSnapshot) Release() {}
