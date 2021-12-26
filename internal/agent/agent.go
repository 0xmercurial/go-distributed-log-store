package agent

import (
	"crypto/tls"
	"fmt"
	"logstore/internal/discovery"
	"logstore/internal/logcomponents"
	"net"
	"sync"

	"google.golang.org/grpc"
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
