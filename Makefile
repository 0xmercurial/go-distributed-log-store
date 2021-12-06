.EXPORT_ALL_VARIABLES:
DIR=secrets
CONFIG_PATH=$(shell pwd)
CONFIG_DIR=${CONFIG_PATH}/${DIR}

.PHONY: init
init:
	mkdir -p ${CONFIG_DIR}

.PHONY: gencert
gencert:
	cfssl gencert \
		-initca tls-config/ca-csr.json | cfssljson -bare ca

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=tls-config/ca-config.json \
		-profile=server \
		tls-config/server-csr.json | cfssljson -bare server

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=tls-config/ca-config.json \
		-profile=client \
		tls-config/client-csr.json | cfssljson -bare client

	mv *.pem *.csr secrets

.PHONY: test
test:

	go test -race ./...


.PHONY: compile
compile:
	protoc internal/log/proto/*.proto \
		--go_out=plugins=grpc:.
