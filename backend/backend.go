package backend

import (
	"context"
	"sync"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

type kubeBackend struct {
	*framework.Backend
	testMode bool
	saMutex  sync.RWMutex
	log      hclog.Logger
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

func NewFactory(log hclog.Logger) logical.Factory {
	return func(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
		b := New()
    b.log = log

		if err := b.Setup(ctx, c); err != nil {
			return nil, err
		}
		return b, nil
	}
}
