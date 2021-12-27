package agent

import (
	"logstore/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
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

	}
}
