# vault-plugin-secrets-kubernetes
Vault secrets manager plugin for kubernetes

# Description
This is a Secret Manager plugin for Kubernetes. Instead of https://github.com/hashicorp/vault-plugin-auth-kubernetes this plugin adds ability to generate kubernetes tokens with ttl.
Secrets will be created for selected ServiceAccount.
This plugin is useful for creating secure deployments, as the created tokens expire after TTL and actually there are no tokens with higher privileges for the namespace for the rest of the time.

## Limitations
* Works only with RBAC
* Creates only Secrets, you can't create ServceAccounts, Roles or RoleBindings through plugin
* Main token rotation is not implemented yet

# How to setup 
## Kubernetes part
First of all we need to create special ServiceAccount, Role and RoleBinding. This Role has only access to create/get/delete Secrets.
```bash
$ kubectl create -f example/clusterrole.yaml               # ClusterRole
$ kubectl create -f example/sa.yaml                 # ServiceAccount
$ kubectl create -f example/clusterrolebinding.yaml        # ClusterRoleBinding
$ # Lets get all needed credentials
$ kubectl describe sa vault
...
  Tokens:              vault-token-c8wgn <-- This is secret name
...
$ kubectl describe secret vault-token-c8wgn
...
Data
====
token: <SA token will be here>
...
$ export TOKEN=<SA token>
$ kubectl get secret vault-token-c8wgn -o yaml
... 
data:
  ca.crt: <one line of base64 encoded CA>
...
$ export MASTER_CA=<master CA>
$ kubectl cluster-info
  Kubernetes master is running at https://my-cluster-domain-name:6443 
$ export MASTER_URL=https://my-cluster-domain-name:6443
```
After this step your should have:
* Vault ServiceAccount token ```$TOKEN```
* Kubernetes CA base64 encoded in one line string  ```$MASTER_CA```
* Kubernetes API URL ```$MASTER_URL```

## Vault part
* First of all put plugin binary to your vault plugins directory (https://www.vaultproject.io/docs/configuration/index.html#plugin_directory)
* Add and enable plugin
```bash
$ export PLUGIN_NAME=vault-plugin-secrets-kubernetes
$ export SHA256SUM=$(sha256sum vault/plugin/vault-plugin-secrets-kubernetes | awk {'print $1'})
$ vault login
$ # Add plugin to catalog
$ vault write sys/plugins/catalog/${PLUGIN_NAME} sha256="${SHA256SUM}" command=${PLUGIN_NAME} 
$ # Enable plugin 
$ vault secrets enable -path=k8s -plugin-name=${PLUGIN_NAME} plugin 
$ vault secrets list    # Check for plugin in catalog 
```
* Configure plugin
```bash
$ vault write k8s/config token=${TOKEN} api-url=${MASTER_URL} CA=${MASTER_CA}
$ vault read k8s/config
```
If write was successful, that means vault successfully checked the login to Kubernetes and we ready to use the plugin.
# How to use
## Kubernetes part
Create ServiceAccount with required Role
```bash
$ kubectl create -f example/deploy-bot-role.yaml               # Role
$ kubectl create -f example/deploy-bot-sa.yaml                 # ServiceAccount
$ kubectl create -f example/deploy-bot-rolebinding.yaml        # RoleBinding 
$ cat deploy-bot-sa.yaml | grep name
  name: deploy-bot  <-- Save SA name
$ kubectl get sa deploy-bot -o yaml | grep namespace
   namespace: my-namespace  <-- Save namespace name
```
## Vault part
Notice: sa means ServiceAccount
```bash
$ vault write k8s/sa/deploy-bot namespace=my-namespace service-account-name=deploy-bot
$ vault write k8s/secrets/deploy-bot ttl=60 # Create secret for deploy-bot with TTL 60 seconds
```
## Gettings help
```bash
$ vault path-help k8s/config
$ vault path-help k8s/sa/name
$ vault path-help k8s/secrets/name
```

# Work with multiple clusters and namespaces
Just enable plugin with different paths:
```bash
$ vault secrets enable -path=us-west2 -plugin-name=${PLUGIN_NAME} plugin 
$ vault secrets enable -path=us-east1 -plugin-name=${PLUGIN_NAME} plugin 
```
Use namespace name in plugin sa name
```bash
$ vault write k8s/sa/my-namespace-deploy-bot namespace=my-namespace service-account-name=deploy-bot
```

# How to build and run locally
Requirements:
* make, sha256sum (apt-get install make coreutils)
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
