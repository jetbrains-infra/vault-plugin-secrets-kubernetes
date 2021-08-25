package backend

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const secretsStoragePrefix = "secrets"

func pathSecrets(b *kubeBackend) *framework.Path {
	return &framework.Path{
		Pattern: fmt.Sprintf("%s/%s", secretsStoragePrefix, framework.GenericNameRegex("name")),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Required. Name of the service account",
			},
			"ttl": {
				Type:        framework.TypeDurationSecond,
				Description: "Optional. Secret time to live",
			},
		},
		// ExistenceCheck: ,
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathSecretsUpdate,
			logical.UpdateOperation: b.pathSecretsUpdate,
		},
		HelpSynopsis:    pathSecretsHelpSyn,
		HelpDescription: pathSecretsHelpDesc,
	}
}

func (b *kubeBackend) pathSecretsUpdate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	b.saMutex.Lock()
	defer b.saMutex.Unlock()
	saName := d.Get("name").(string)
	sa, err := getServiceAccount(ctx, saName, req.Storage)
	if err != nil {
		return nil, err
	}

	if sa == nil {
		return logical.ErrorResponse(fmt.Sprintf("ServiceAccount '%s' not found", saName)), nil
	}

	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return logical.ErrorResponse("Please configure plugin with 'config' path"), nil
	}

	var ttl int64
	ttlRaw, ok := d.GetOk("ttl")
	if ok {
		ttl = int64(ttlRaw.(int))
		if ttl > int64(config.MaxTTL.Seconds()) {
			return logical.ErrorResponse(fmt.Sprintf("Max TTL configured to '%d', you try to create TTL '%d'", int64(config.MaxTTL.Seconds()), ttl)), nil
		}
	} else {
		ttl = int64(config.TTL.Seconds())
	}

	resp, err := b.createSecret(ctx, req.Storage, config, sa)
	if err != nil {
		return nil, err
	}
	resp.Secret.TTL = time.Duration(ttl) * time.Second
	resp.Secret.MaxTTL = config.MaxTTL
	return resp, nil
}

const pathSecretsHelpSyn = `Generate Secret for selected Service Account`
const pathSecretsHelpDesc = `
This path allow you to generate Secret with token for selected Service Account, also you will get kubernetes CA_base64
 and namespace for using it in CI/CD`
