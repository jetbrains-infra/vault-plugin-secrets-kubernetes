# vault-plugin-secrets-k8s
Vault secrets manager plugin for kubernetes

# Description
TBD

# How to setup 
TBD

# How to build and run locally
Requirements:
* make sha256sum (apt-get install make coreutils)
* golang ~1.10
* docker
* docker-compose
* vault CLI utility

```bash
$ glide up -v
$ make test
$ make build up init-plugin
$ VAULT_ADDR=http://127.0.0.1:8200 vault login  # token = 123qwe
$ VAULT_ADDR=http://127.0.0.1:8200 vault path-help k8s/config
Request:        config
Matching Route: ^config$

Configure the Kubernetes backend

## PARAMETERS

    CA (string)
        Kubernetes apiserver Certificate Authority (base64 encoded)

    api-url (string)
        URL to kubernetes apiserver https endpoint

    max-ttl (duration (sec))
        Maximum time a secret is valid for. If <= 0, will use system default.

    token (string)
        ServiceAccount token with permissions to list, create, delete Secrets

    ttl (duration (sec))
        Default lease for generated secrets. If <= 0, will use system default.

## DESCRIPTION

The Kubernetes backend requires credentials for managing Secrets in cluster. This endpoint is used to configure those
 credentials as well as default values for the backend in general
```
