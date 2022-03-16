.EXPORT_ALL_VARIABLES:
DIR=secrets
CONFIG_PATH=$(shell pwd)
GP = $(GOPATH)
CONFIG_DIR=${CONFIG_PATH}/${DIR}

.PHONY: init
init:
	mkdir -p ${CONFIG_DIR}
	echo ${GOPATH}

.PHONY: gencert
gencert:
	${GOPATH}/bin/cfssl gencert \
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
		-cn="root" \
		tls-config/client-csr.json | cfssljson -bare root-client

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=tls-config/ca-config.json \
		-profile=client \
		-cn="nobody" \
		tls-config/client-csr.json | cfssljson -bare nobody-client

	mv *.pem *.csr secrets

.PHONY: test
test:

	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

	cd internal/server && go test -v -debug=true


.PHONY: compile
compile:
	protoc internal/log/proto/*.proto \
		--go_out=plugins=grpc:.
