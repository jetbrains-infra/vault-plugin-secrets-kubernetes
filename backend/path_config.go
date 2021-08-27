package backend

import (
	"context"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const ConfigStorageKey = "config"
const ConfigPath = "config"

func pathConfig(b *kubeBackend) *framework.Path {
	return &framework.Path{
		Pattern: ConfigPath,
		Fields: map[string]*framework.FieldSchema{
			"token": {
				Type:        framework.TypeString,
				Description: `ServiceAccount token with permissions to list, create, delete Secrets`,
			},
			"api-url": {
				Type:        framework.TypeString,
				Description: `URL to kubernetes apiserver https endpoint`,
			},
			"CA": {
				Type:        framework.TypeString,
				Description: `Kubernetes apiserver Certificate Authority (base64 encoded)`,
			},
			"ttl": {
				Type:        framework.TypeDurationSecond,
				Description: "Default lease for generated secrets. If <= 0, will use system default.",
			},
			"max-ttl": {
				Type:        framework.TypeDurationSecond,
				Description: "Maximum time a secret is valid for. If <= 0, will use system default.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathConfigRead,
			logical.UpdateOperation: b.pathConfigWrite,
			logical.DeleteOperation: b.pathConfigDelete,
		},

		HelpSynopsis:    pathConfigHelpSyn,
		HelpDescription: pathConfigHelpDesc,
	}
}

func (b *kubeBackend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	cfg, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"api-url": cfg.APIURL,
			"ttl":     int64(cfg.TTL / time.Second),
			"max-ttl": int64(cfg.MaxTTL / time.Second),
			"CA":      cfg.CA,
		},
	}, nil
}

func (b *kubeBackend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	cfg, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		cfg = &config{
			TTL:    1800 * time.Second,
			MaxTTL: 3600 * time.Second,
		}
	}

	tokenRaw, ok := data.GetOk("token")
	if ok {
		cfg.Token = tokenRaw.(string)
	}

	apiURL, ok := data.GetOk("api-url")
	if ok {
		cfg.APIURL = apiURL.(string)
	}

	CARaw, ok := data.GetOk("CA")
	if ok {
		cfg.CA = CARaw.(string)
	}

	// Update token TTL.
	ttlRaw, ok := data.GetOk("ttl")
	if ok {
		cfg.TTL = time.Duration(ttlRaw.(int)) * time.Second
	}

	// Update token Max TTL.
	maxTTLRaw, ok := data.GetOk("max-ttl")
	if ok {
		cfg.MaxTTL = time.Duration(maxTTLRaw.(int)) * time.Second
	}

	entry, err := logical.StorageEntryJSON(ConfigStorageKey, cfg)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *kubeBackend) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, ConfigStorageKey); err != nil {
		return nil, err
	}

	return nil, nil
}

type config struct {
	Token  string
	APIURL string
	CA     string

	TTL    time.Duration
	MaxTTL time.Duration
}

func getConfig(ctx context.Context, s logical.Storage) (*config, error) {
	var cfg config
	cfgRaw, err := s.Get(ctx, ConfigStorageKey)
	if err != nil {
		return nil, err
	}
	if cfgRaw == nil {
		return nil, nil
	}

	if err = cfgRaw.DecodeJSON(&cfg); err != nil {
		return nil, err
	}

	return &cfg, err
}

const pathConfigHelpSyn = `Configure the Kubernetes backend`

const pathConfigHelpDesc = `
The Kubernetes backend requires credentials for managing Secrets in cluster. This endpoint is used to configure those
 credentials as well as default values for the backend in general.`
