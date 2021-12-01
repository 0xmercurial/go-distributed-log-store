package server

import (
	"context"
	"io/ioutil"
	"logstore/internal/log/proto"
	log "logstore/internal/logcomponents"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func setupTest(t *testing.T, fn func(*Config)) (
	client proto.LogClient,
	config *Config,
	teardown func(),
) {
	t.Helper() //marks func as a helper (logs will be ignored)
	listener, err := net.Listen("tcp", ":0")
	assert.NoError(t, err)

	dialOps := []grpc.DialOption{grpc.WithInsecure()}
	clientConn, err := grpc.Dial(listener.Addr().String(), dialOps...)
	assert.NoError(t, err)

	dir, err := ioutil.TempDir("", "srv-test")
	assert.NoError(t, err)

	commitLog, err := log.NewLog(dir, log.Config{})
	assert.NoError(t, err)

	config = &Config{
		CommitLog: commitLog,
	}
	if fn != nil {
		fn(config)
	}

	srv, err := NewGRPCServer(config)
	assert.NoError(t, err)

	go func() {
		srv.Serve(listener)
	}()

	client = proto.NewLogClient(clientConn)

	return client, config, func() { //returning an anon func that shutsdown srv
		srv.Stop()
		clientConn.Close()
		listener.Close()
		commitLog.Remove()
	}
}

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client proto.LogClient,
		config *Config,
	){
		"unary success": testUnaryAppendRead,
		//"stream success":     testStreamAppendRead,
		//"read out of bounds": testOOBRead,
	} {
		t.Run(scenario, func(t *testing.T) {
			client, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, client, config)
		})
	}
}

func testUnaryAppendRead(t *testing.T, client proto.LogClient, config *Config) {
	ctx := context.Background()

	want := &proto.Record{
		Value: []byte("record"),
	}
	appReq := &proto.AppendRequest{
		Record: want,
	}
	append, err := client.Append(ctx, appReq)
	assert.NoError(t, err)

	readReq := &proto.ReadRequest{
		Offset: append.Offset,
	}
	read, err := client.Read(ctx, readReq)
	assert.NoError(t, err)
	assert.Equal(t, want.Value, read.Record.Value)
	assert.Equal(t, want.Offset, read.Record.Offset)
}
