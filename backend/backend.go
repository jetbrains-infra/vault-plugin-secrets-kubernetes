package backend

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

)

type kubeBackend struct {
	*framework.Backend
	testMode bool
	saMutex sync.RWMutex
}

// New creates and returns new instance of Kubernetes secrets manager backend
func New() *kubeBackend {
	var b kubeBackend

	b.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,

		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{"login"},
		},
		WALRollback:       b.walRollback,
		WALRollbackMinAge: 5 * time.Minute,
		Paths: []*framework.Path{
			pathConfig(&b),
			pathServiceAccounts(&b),
			pathServiceAccountsList(&b),
			pathSecrets(&b),
			// TODO P1 pathConfigRotateToken
		},
		Secrets: []*framework.Secret{
			secretAccessTokens(&b),
		},
	}

	b.testMode = false

	return &b
}

// Factory creates and returns new backend with BackendConfig
func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := New()
	if err := b.Setup(ctx, c); err != nil {
		return nil, err
	}
	return b, nil
}
