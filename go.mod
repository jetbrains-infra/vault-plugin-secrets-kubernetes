module github.com/jetbrains-infra/vault-plugin-secrets-kubernetes

go 1.16

require (
	github.com/hashicorp/errwrap v1.1.0
	github.com/hashicorp/go-hclog v0.16.1
	github.com/hashicorp/vault/api v1.1.1
	github.com/hashicorp/vault/sdk v0.2.1
	github.com/mitchellh/mapstructure v1.3.2
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
)
