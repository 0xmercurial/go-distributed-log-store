CONFIG_PATH="secrets"

.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}
	export CONFIG_DIR=${CONFIG_PATH}

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
		tls-config/client-csr.json | cfssljson -bare server

	mv *.pem *.csr ${CONFIG_PATH}
	
.PHONY: test
test:
	export CONFIG_DIR=${CONFIG_PATH}
	go test -race ./...


.PHONY: compile
compile:
	protoc internal/log/proto/*.proto \
		--go_out=plugins=grpc:.
