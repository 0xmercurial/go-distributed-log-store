package server

import (
	"context"
	"io/ioutil"
	"logstore/internal/authz"
	tlscf "logstore/internal/config"
	"logstore/internal/log/proto"
	log "logstore/internal/logcomponents"

	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func setupTest(t *testing.T, fn func(*Config)) (
	rootClient proto.LogClient,
	nobodyClient proto.LogClient,
	config *Config,
	teardown func(),
) {
	t.Helper() //marks func as a helper (logs will be ignored)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	//Client Setup
	newClient := func(certPath, keyPath string) (
		*grpc.ClientConn,
		proto.LogClient,
		[]grpc.DialOption,
	) {
		config := tlscf.TLSConfig{
			CertFile: certPath,
			KeyFile:  keyPath,
			CAFile:   tlscf.CAFile,
			Server:   false,
		}
		tlsConfig, err := tlscf.SetupFromTLSConfig(config)
		assert.NoError(t, err)

		tlsCreds := credentials.NewTLS(tlsConfig)
		opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
		conn, err := grpc.Dial(listener.Addr().String(), opts...)
		assert.NoError(t, err)
		client := proto.NewLogClient(conn)
		return conn, client, opts
	}

	rootConn, rootClient, _ := newClient(
		tlscf.RootClientCertFile,
		tlscf.RootClientKeyFile,
	)

	nobodyConn, nobodyClient, _ := newClient(
		tlscf.NobodyClientCertFile,
		tlscf.NobodyClientKeyFile,
	)
	//Server Setup
	serverInputConf := tlscf.TLSConfig{
		CertFile:      tlscf.ServerCertFile,
		KeyFile:       tlscf.ServerKeyFile,
		CAFile:        tlscf.CAFile,
		ServerAddress: listener.Addr().String(),
		Server:        true,
	}
	serverTLSConfig, err := tlscf.SetupFromTLSConfig(serverInputConf)
	assert.NoError(t, err)
	serverCreds := credentials.NewTLS(serverTLSConfig)

	dir, err := ioutil.TempDir("", "srv-test")
	assert.NoError(t, err)

	commitLog, err := log.NewLog(dir, log.Config{})
	assert.NoError(t, err)

	authorizer := authz.New(tlscf.ACLModelFile, tlscf.ACLPolicyFile)
	config = &Config{
		CommitLog:  commitLog,
		Authorizer: authorizer,
	}
	if fn != nil {
		fn(config)
	}

	srv, err := NewGRPCServer(
		config,
		grpc.Creds(serverCreds),
	)
	assert.NoError(t, err)

	go func() {
		srv.Serve(listener)
	}()

	return rootClient, nobodyClient, config, func() {
		srv.Stop()
		rootConn.Close()
		nobodyConn.Close()
		listener.Close()
	}
}

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		rootClient proto.LogClient,
		nobodyClient proto.LogClient,
		config *Config,
	){
		"unary success":      testUnaryAppendRead,
		"stream success":     testStreamAppendRead,
		"read out of bounds": testOOBRead,
		"unauthz failure":    testNoAuthZ,
	} {
		t.Run(scenario, func(t *testing.T) {
			rootClient, nobodyClient, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, rootClient, nobodyClient, config)
		})
	}
}

func testUnaryAppendRead(
	t *testing.T,
	client, _ proto.LogClient,
	config *Config,
) {
	ctx := context.Background()

	expected := &proto.Record{
		Value: []byte("record"),
	}
	appReq := &proto.AppendRequest{
		Record: expected,
	}
	append, err := client.Append(ctx, appReq)
	assert.NoError(t, err)

	readReq := &proto.ReadRequest{
		Offset: append.Offset,
	}
	read, err := client.Read(ctx, readReq)
	assert.NoError(t, err)
	assert.Equal(t, expected.Value, read.Record.Value)
	assert.Equal(t, expected.Offset, read.Record.Offset)
}

func testStreamAppendRead(
	t *testing.T,
	client, _ proto.LogClient,
	config *Config,
) {
	ctx := context.Background()

	records := []*proto.Record{
		{
			Value:  []byte("uno"),
			Offset: 0,
		}, {
			Value:  []byte("dos"),
			Offset: 1,
		},
	}

	{
		stream, err := client.AppendStream(ctx)
		assert.NoError(t, err)

		for offset, record := range records {
			apdReq := &proto.AppendRequest{
				Record: record,
			}
			err = stream.Send(apdReq)
			assert.NoError(t, err)

			res, err := stream.Recv()
			assert.NoError(t, err)
			if res.Offset != uint64(offset) {
				t.Fatalf(
					"actual offset: %d, expected %d",
					res.Offset,
					offset,
				)
			}
		}
	}

	{
		readReq := &proto.ReadRequest{Offset: 0}
		stream, err := client.ReadStream(ctx, readReq)
		assert.NoError(t, err)

		for i, record := range records {
			res, err := stream.Recv()

			assert.NoError(t, err)
			expected := &proto.Record{
				Value:  record.Value,
				Offset: uint64(i),
			}
			assert.Equal(t, res.Record, expected)
		}
	}
}

func testOOBRead(
	t *testing.T,
	client, _ proto.LogClient,
	config *Config,
) {
	ctx := context.Background()

	record := &proto.Record{
		Value: []byte("record"),
	}
	appReq := &proto.AppendRequest{
		Record: record,
	}
	append, err := client.Append(ctx, appReq)
	assert.NoError(t, err)

	readReq := &proto.ReadRequest{
		Offset: append.Offset + 1,
	}
	read, err := client.Read(ctx, readReq)
	if read != nil {
		t.Error("read not nil")
	}

	expected := grpc.Code(proto.ErrOffOutOfRange{}.GRPCStatus().Err())
	actual := grpc.Code(err)
	if actual != expected {
		t.Fatalf("actual err: %v, expected: %v", actual, expected)
	}

}

func testNoAuthZ(
	t *testing.T,
	_,
	client proto.LogClient,
	config *Config,
) {
	ctx := context.Background()

	record := &proto.Record{
		Value: []byte("record"),
	}
	appReq := &proto.AppendRequest{
		Record: record,
	}
	append, err := client.Append(ctx, appReq)
	assert.Nil(t, append)

	actualCode, expectedCode := status.Code(err), codes.PermissionDenied
	if actualCode != expectedCode {
		t.Fatalf("actual: %d, expected: %d", actualCode, expectedCode)
	}

	readReq := &proto.ReadRequest{
		Offset: 0,
	}
	read, err := client.Read(ctx, readReq)
	assert.Nil(t, read)

	actualCode, expectedCode = status.Code(err), codes.PermissionDenied
	if actualCode != expectedCode {
		t.Fatalf("actual: %d, expected: %d", actualCode, expectedCode)
	}
}
