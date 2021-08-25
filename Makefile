VAULT_ADDR=http://127.0.0.1:8200
export VAULT_ADDR
SHA256SUM=$(shell sha256sum vault/plugin/vault-plugin-secrets-kubernetes | awk {'print $$1'})
PLUGIN_NAME=vault-plugin-secrets-kubernetes

build:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -installsuffix cgo -o vault/plugin/${PLUGIN_NAME} .
	#goupx vault/plugin/${PLUGIN_NAME}

login:
	echo "123qwe" | vault login -

add-plugin:
	vault write sys/plugins/catalog/${PLUGIN_NAME} sha256="${SHA256SUM}" command=${PLUGIN_NAME}

enable-plugin:
	vault secrets enable -path=k8s -plugin-name=${PLUGIN_NAME} plugin
	vault secrets list

list-plugins:
	vault read sys/plugins/catalog

configure-plugin:
	vault write k8s/config token=${TOKEN} api-url=${MASTER_URL} CA=${MASTER_CA}
	vault read k8s/config

create-sa:
	vault write k8s/sa/it-deployer namespace=it service-account-name=deployer

get-token:
	vault read k8s/secrets/it-deployer

up:
	docker-compose down
	docker-compose up -d

test:
	go test -v -cover $(shell go list ./... | grep -v /vendor/)

init-plugin: login add-plugin enable-plugin list-plugins

crosscompile:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -ldflags="-s -w" -installsuffix cgo -o vault/plugin/${PLUGIN_NAME}-linux-amd64 .
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -a -ldflags="-s -w" -installsuffix cgo -o vault/plugin/${PLUGIN_NAME}-linux-arm64 .
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -a -ldflags="-s -w" -installsuffix cgo -o vault/plugin/${PLUGIN_NAME}-windows-amd64 .
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -a -ldflags="-s -w" -installsuffix cgo -o vault/plugin/${PLUGIN_NAME}-darwin-amd64 .
