package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"logstore/internal/config"
	"logstore/internal/log/proto"
	"logstore/internal/portutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestAgent(t *testing.T) {
	serverTLSConfig, err := config.SetupFromTLSConfig(
		config.TLSConfig{
			CertFile:      config.ServerCertFile,
			KeyFile:       config.ServerKeyFile,
			CAFile:        config.CAFile,
			Server:        true,
			ServerAddress: "127.0.0.1",
		},
	)
	assert.NoError(t, err)

	peerTLSConfig, err := config.SetupFromTLSConfig(
		config.TLSConfig{
			CertFile:      config.RootClientCertFile,
			KeyFile:       config.RootClientKeyFile,
			CAFile:        config.CAFile,
			Server:        false,
			ServerAddress: "127.0.0.1",
		},
	)
	assert.NoError(t, err)

	var agents []*Agent
	for i := 0; i < 3; i++ {
		ports := portutil.Get(2)
		bindAddr := fmt.Sprintf("%s:%d", "127.0.01", ports[0])
		rpcPort := ports[1]

		dataDir, err := ioutil.TempDir("", "agent-test-log")
		assert.NoError(t, err)

		var startJoinAddrs []string
		if i != 0 {
			startJoinAddrs = append(
				startJoinAddrs,
				agents[0].Config.BindAddr,
			)
		}

		agent, err := New(
			Config{
				NodeName:        fmt.Sprintf("%d", i),
				StartJoinAddrs:  startJoinAddrs,
				BindAddr:        bindAddr,
				RPCPort:         rpcPort,
				DataDir:         dataDir,
				ACLModelFile:    config.ACLModelFile,
				ACLPolicyFile:   config.ACLPolicyFile,
				ServerTLSConfig: serverTLSConfig,
				PeerTLSConfig:   peerTLSConfig,
			},
		)
		assert.NoError(t, err)
		agents = append(agents, agent)
	}
	defer func() {
		for _, agent := range agents {
			err := agent.Shutdown()
			assert.NoError(t, err)
			assert.NoError(
				t,
				os.RemoveAll(agent.Config.DataDir),
			)
		}
	}()
	time.Sleep(3 * time.Second)

	leader := client(t, agents[0], peerTLSConfig)
	appendResponse, err := leader.Append(
		context.Background(),
		&proto.AppendRequest{
			Record: &proto.Record{
				Value: []byte("record"),
			},
		},
	)
	assert.NoError(t, err)

	readResponse, err := leader.Read(
		context.Background(),
		&proto.ReadRequest{
			Offset: appendResponse.Offset,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, readResponse.Record.Value, []byte("record"))

	time.Sleep(3 * time.Second)

	follower := client(t, agents[1], peerTLSConfig)
	readResponse, err = follower.Read(
		context.Background(),
		&proto.ReadRequest{
			Offset: appendResponse.Offset,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, readResponse.Record.Value, []byte("record"))
}

func client(
	t *testing.T,
	agent *Agent,
	tlsConfig *tls.Config,
) proto.LogClient {
	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
	}
	rpcAddr, err := agent.Config.RPCAddr()
	assert.NoError(t, err)
	conn, err := grpc.Dial(
		fmt.Sprintf("%s", rpcAddr),
		opts...,
	)
	assert.NoError(t, err)
	client := proto.NewLogClient(conn)
	return client
}
