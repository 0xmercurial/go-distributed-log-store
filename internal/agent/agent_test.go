package agent

import (
	"logstore/internal/config"
	"testing"
)

func TestAgent(t *testing.T) {
	serverTLSConfig, err := config.SetupFromTLSConfig(
		config.TLSConfig{
			CertFile: config.ServerCertFile,
			KeyFile:  config.ServerKeyFile,
			CAFile:   config.CAFile,
		},
	)
}
