package logcomponents

import (
	"context"
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
	if err != nil {
		r.logError(err, "failed to dial", addr)
		return
	}
	defer clientConn.Close()

	client := proto.NewLogClient(clientConn)

	ctx := context.Background()
	stream, err := client.ReadStream(
		ctx,
		&proto.ReadRequest{Offset: 0},
	)
	if err != nil {
		r.logError(err, "failed to read", addr)
		return
	}

	records := make(chan *proto.Record)
	go func() {
		for {
			recv, err := stream.Recv()
			if err != nil {
				r.logError(err, "failed to receive msg", addr)
				return
			}
			records <- recv.Record
		}
	}()

	for {
		select {
		case <-r.close:
			return
		case <-leave:
			return
		case record := <-records:
			_, err = r.LocalServer.Append(
				ctx,
				&proto.AppendRequest{
					Record: record,
				},
			)
			if err != nil {
				r.logError(err, "failed to append", addr)
				return
			}
		}
	}
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

func (r *Replica) Leave(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()

	if _, ok := r.servers[name]; !ok {
		return nil
	}
	close(r.servers[name])
	delete(r.servers, name)
	return nil
}

func (r *Replica) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()

	if r.closed {
		return nil
	}
	r.closed = true
	close(r.close)
	return nil
}

func (r *Replica) logError(err error, msg, addr string) {
	r.logger.Error(
		msg,
		zap.String("addr", addr),
		zap.Error(err),
	)
}
