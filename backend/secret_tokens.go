package backend

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/errwrap"

	"github.com/mitchellh/mapstructure"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	secretTypeAccessToken = "secret_token"
	secretPrefix          = "vault"
	symbolsForGenerator   = "abcdefghijklmnopqrstuvwxyz0123456789"
	secretWALKind         = "secret"
)

func secretAccessTokens(b *kubeBackend) *framework.Secret {
	return &framework.Secret{
		Type: secretTypeAccessToken,
		Fields: map[string]*framework.FieldSchema{
			"token": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Token of the secret",
			},
			"namespace": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Namespace of the secret",
			},
			"ca": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "CA of the api server",
			},
		},

		Renew:  b.secretAccessTokenRenew,
		Revoke: b.secretAccessTokenRevoke,
	}
}

type walSecret struct {
	Name      string
	Namespace string
}

func (b *kubeBackend) createSecret(ctx context.Context, s logical.Storage, c *config, sa *ServiceAccount) (*logical.Response, error) {
	name := fmt.Sprintf("%s-%s-%s", secretPrefix, sa.ServiceAccountName, generatePostfix(8))

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": sa.ServiceAccountName,
			},
		},
		Type: "kubernetes.io/service-account-token",
	}
	// Write to the WAL that this user will be created. We do this before
	// the user is created because if switch the order then the WAL put
	// can fail, which would put us in an awkward position: we have a user
	// we need to rollback but can't put the WAL entry to do the rollback.
	walID, err := framework.PutWAL(ctx, s, secretWALKind, &walSecret{
		Name:      name,
		Namespace: sa.Namespace,
	})
	if err != nil {
		return nil, err
	}

	var token, namespace string
	var CABase64 interface{}
	var resp *v1.Secret

	if !b.testMode {
		clientSet, err := getClientSet(c)
		if err != nil {
			return nil, err
		}
		_, err = clientSet.CoreV1().Secrets(sa.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, errwrap.Wrapf("Unable to create secret, {{err}}", err)
		}
		// Do 5 tries to get secret, due to it may not generated after first try
		for range []int{0, 1, 2, 3, 4} {
			secretResp, err := clientSet.CoreV1().Secrets(sa.Namespace).Get(ctx, secret.Name, metav1.GetOptions{})
			if err != nil {
				return nil, errwrap.Wrapf("Unable to get secret, {{err}}", err)
			}
			if len(secretResp.Data) == 0 {
				time.Sleep(time.Second)
				continue
			}
			resp = secretResp
			token = string(secretResp.Data["token"])
			namespace = string(secretResp.Data["namespace"])
			CABase64 = secretResp.Data["ca.crt"]
			break
		}
		if resp == nil || len(resp.Data) == 0 {
			return nil, errors.New("unable to get secret with 5 tries, Data was empty")
		}
	} else {
		token = "test"
		namespace = "test"
		CABase64 = "test"
	}

	// Remove the WAL entry, we succeeded! If we fail, we don't return
	// the secret because it'll get rolled back anyways, so we have to return
	// an error here.
	if err := framework.DeleteWAL(ctx, s, walID); err != nil {
		return nil, errwrap.Wrapf("failed to commit WAL entry: {{err}}", err)
	}

	return b.Secret(secretTypeAccessToken).Response(map[string]interface{}{
		"token":     token,
		"namespace": namespace,
		"CA_base64": CABase64,
	}, map[string]interface{}{
		"secret-name": name,
		"namespace":   sa.Namespace,
	}), nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generatePostfix(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = symbolsForGenerator[rand.Intn(len(symbolsForGenerator))]
	}
	return string(b)
}

func (b *kubeBackend) secretAccessTokenRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	c, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	resp := &logical.Response{Secret: req.Secret}
	resp.Secret.TTL = c.TTL
	resp.Secret.MaxTTL = c.MaxTTL
	return resp, nil
}

func (b *kubeBackend) secretAccessTokenRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	c, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	clientSet, err := getClientSet(c)
	if err != nil {
		return nil, err
	}

	namespace := req.Secret.InternalData["namespace"].(string)
	name := req.Secret.InternalData["secret-name"].(string)

	err = clientSet.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})

	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (b *kubeBackend) walRollback(ctx context.Context, r *logical.Request, kind string, data interface{}) error {
	var entry walSecret
	if err := mapstructure.Decode(data, &entry); err != nil {
		return err
	}

	switch kind {
	case secretWALKind:
		r.Secret = &logical.Secret{
			InternalData: map[string]interface{}{
				"secret-name": entry.Name,
				"namespace":   entry.Namespace,
			},
		}
		_, err := b.secretAccessTokenRevoke(ctx, r, nil)
		return err
	default:
		return fmt.Errorf("unknown kind to rollback %s", kind)
	}
}
