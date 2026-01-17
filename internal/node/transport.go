package node

import (
	"net"
	"time"

	"github.com/hashicorp/raft"
)

const (
	maxPool = 3
	timeout = time.Second * 10
)

func NewNodeTransport(addr string) (*raft.NetworkTransport, error) {
	bind, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	return raft.NewTCPTransport(
		addr,
		bind,
		maxPool,
		timeout,
		nil)
}
