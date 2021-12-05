package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

type TLSConfig struct {
	CertFile      string
	KeyFile       string
	CAFile        string
	ServerAddress string
	Server        bool
}

func SetupFromTLSConfig(config TLSConfig) (*tls.Config, error) {
	var err error
	tlsConfig := &tls.Config{}
	//Setting certs. client + server need this to verify certs
	if config.CertFile != "" && config.KeyFile != "" {
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(
			config.CertFile, config.KeyFile,
		)
		if err != nil {
			return nil, err
		}
	}

	if config.CAFile != "" {
		b, err := ioutil.ReadFile(config.CAFile)
		if err != nil {
			return nil, err
		}
		ca := x509.NewCertPool()
		success := ca.AppendCertsFromPEM(b)
		if !success {
			return nil, fmt.Errorf(
				"failed to parse root certificate: %q",
				config.CAFile,
			)
		}
		if config.Server {
			tlsConfig.ClientCAs = ca
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			//Else client, use RootCAs to verify server cert
			tlsConfig.RootCAs = ca
		}
		tlsConfig.ServerName = config.ServerAddress
	}
	return tlsConfig, nil
}
