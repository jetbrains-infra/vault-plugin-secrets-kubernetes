package backend

import (
	"context"
	"fmt"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

const (
	saStoragePrefix = "sa"
)

// ServiceAccount bind to Kubernetes ServiceAccount with ServiceAccountName and Namespace, all permissions are
// stored in Kubernetes and not managed by vault plugin
type ServiceAccount struct {
	Name               string
	Namespace          string
	ServiceAccountName string
}

func (r *ServiceAccount) save(ctx context.Context, s logical.Storage) error {
	entry, err := logical.StorageEntryJSON(fmt.Sprintf("%s/%s", saStoragePrefix, r.Name), r)
	if err != nil {
		return err
	}

	return s.Put(ctx, entry)
}

func (r *ServiceAccount) toResponse() *logical.Response {
	return &logical.Response{
		Data: map[string]interface{}{
			"namespace":            r.Namespace,
			"service-account-name": r.ServiceAccountName,
		},
	}
}

func getServiceAccount(ctx context.Context, name string, s logical.Storage) (*ServiceAccount, error) {
	entry, err := s.Get(ctx, fmt.Sprintf("%s/%s", saStoragePrefix, name))
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	sa := &ServiceAccount{}
	if err := entry.DecodeJSON(sa); err != nil {
		return nil, err
	}
	return sa, nil
}

func pathServiceAccounts(b *kubeBackend) *framework.Path {
	return &framework.Path{
		Pattern: fmt.Sprintf("%s/%s", saStoragePrefix, framework.GenericNameRegex("name")),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Required. Name of the Vault object",
			},
			"namespace": {
				Type:        framework.TypeString,
				Description: "Required. ServiceAccount's namespace",
			},
			"service-account-name": {
				Type:        framework.TypeString,
				Description: "Required. Name of ServiceAccount in Kubernetes namespace",
			},
		},
		// ExistenceCheck: b.pathRoleSetExistenceCheck,
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.DeleteOperation: b.pathServiceAccountDelete,
			logical.ReadOperation:   b.pathServiceAccountRead,
			// logical.CreateOperation: b.pathServiceAccountCreateUpdate,
			logical.UpdateOperation: b.pathServiceAccountCreateUpdate,
		},
		HelpSynopsis:    pathServiceAccountHelpSyn,
		HelpDescription: pathServiceAccountHelpDesc,
	}
}

func (b *kubeBackend) pathServiceAccountCreateUpdate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	new := false
	nameRaw, ok := d.GetOk("name")
	if !ok {
		return logical.ErrorResponse("name is required"), nil
	}

	name := nameRaw.(string)

	sa, err := getServiceAccount(ctx, name, req.Storage)
	if err != nil {
		return nil, err
	}

	if sa == nil {
		new = true
		sa = &ServiceAccount{
			Name: nameRaw.(string),
		}
	}

	namespaceRaw, ok := d.GetOk("namespace")
	if ok {
		sa.Namespace = namespaceRaw.(string)
	} else if !ok && new {
		return logical.ErrorResponse("namespace is required"), nil
	}

	saNameRaw, ok := d.GetOk("service-account-name")
	if ok {
		sa.ServiceAccountName = saNameRaw.(string)
	} else if !ok && new {
		return logical.ErrorResponse("service-account-name is required"), nil
	}

	if err := sa.save(ctx, req.Storage); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *kubeBackend) pathServiceAccountRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	nameRaw, ok := d.GetOk("name")
	if !ok {
		return logical.ErrorResponse("name is required"), nil
	}
	sa, err := getServiceAccount(ctx, nameRaw.(string), req.Storage)
	if err != nil {
		return nil, err
	}
	if sa == nil {
		return nil, nil
	}
	return sa.toResponse(), nil
}

func (b *kubeBackend) pathServiceAccountDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	nameRaw, ok := d.GetOk("name")
	if !ok {
		return logical.ErrorResponse("name is required"), nil
	}
	name := nameRaw.(string)
	sa, err := getServiceAccount(ctx, name, req.Storage)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("unable to get sa '%s': {{err}}", name), err)
	}
	if sa == nil {
		return nil, nil
	}
	if err := req.Storage.Delete(ctx, fmt.Sprintf("%s/%s", saStoragePrefix, sa.Name)); err != nil {
		return nil, err
	}
	return nil, nil
}

const pathServiceAccountHelpSyn = `Read/write ServiceAccount bindings Vault <-> Kubernetes.`
const pathServiceAccountHelpDesc = `
This path allow you create service account, which bind Kubernetes ServiceAccount. Vault will
create Secrets for this ServiceAccount in Kubernetes.`
