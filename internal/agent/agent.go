package agent

import (
	"crypto/tls"
	"fmt"
	"logstore/internal/authz"
	"logstore/internal/discovery"
	"logstore/internal/logcomponents"
	"logstore/internal/server"
	"net"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

/*
Agent struct is used for node coordination/operation
*/
type Agent struct {
	Config
	log        *logcomponents.Log
	server     *grpc.Server
	membership *discovery.Membership
	replicator *logcomponents.Replica

	shutdown     bool
	shutdowns    chan struct{}
	shutdownLock sync.Mutex
}

func New(config Config) (*Agent, error) {
	a := &Agent{
		Config:    config,
		shutdowns: make(chan struct{}),
	}
	setup := []func() error{
		a.setupLogger,
		a.setupLog,
		a.setupServer,
		//a.setupMembership,
	}

	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func (a *Agent) setupLogger() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}

func (a *Agent) setupLog() error {
	var err error
	a.log, err = logcomponents.NewLog(
		a.Config.DataDir,
		logcomponents.Config{},
	)

	return err
}

func (a *Agent) setupServer() error {
	authorizer := authz.New(
		a.Config.ACLModelFile,
		a.Config.ACLPolicyFile,
	)

	serverConfig := &server.Config{
		CommitLog:  a.log,
		Authorizer: authorizer,
	}

	var opts []grpc.ServerOption
	if a.Config.ServerTLSConfig != nil {
		creds := credentials.NewTLS(a.Config.ServerTLSConfig)
		opts = append(opts, grpc.Creds(creds))
	}

	var err error
	a.server, err = server.NewGRPCServer(serverConfig, opts...)
	if err != nil {
		return err
	}
	rpcAddr, err := a.RPCAddr()
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}
	go func() {
		if err := a.server.Serve(listener); err != nil {
			//_ = a.Shutdown()
		}
	}()
	return err
}

type Config struct {
	ServerTLSConfig *tls.Config
	PeerTLSConfig   *tls.Config
	DataDir         string
	BindAddr        string
	RPCPort         int
	NodeName        string
	StartJoinAddrs  []string
	ACLModelFile    string
	ACLPolicyFile   string
}

func (c Config) RPCAddr() (string, error) {
	host, _, err := net.SplitHostPort(c.BindAddr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", host, c.RPCPort), nil
}
