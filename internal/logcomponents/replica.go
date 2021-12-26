package logcomponents

import (
	"logstore/internal/log/proto"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Replica struct {
	DialOptions []grpc.DialOption
	LocalServer proto.LogClient
	logger      *zap.Logger

	mu      sync.Mutex
	servers map[string]chan struct{}
	closed  bool
	close   chan struct{}
}

func (r *Replica) init() {
	if r.logger == nil {
		r.logger = zap.L().Named("replica")
	}
	if r.servers == nil {
		r.servers = make(map[string]chan struct{})
	}
	if r.close == nil {
		r.close = make(chan struct{})
	}
}

func (r *Replica) replicate(addr string, leave chan struct{}) {
	clientConn, err := grpc.Dial(addr, r.DialOptions...)
}

func (r *Replica) Join(name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()

	if r.closed {
		return nil
	}

	if _, ok := r.servers[name]; ok {
		return nil
	}
	r.servers[name] = make(chan struct{})
	go r.replicate(addr, r.servers[name])

	return nil
}
