package backend

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/logical"
)

func TestSecretsUpdateNotFound(t *testing.T) {
	b, s := getTestBackend(t)

	request := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/test", secretsStoragePrefix),
		Data:      map[string]interface{}{},
		Storage:   s,
	}

	e := "ServiceAccount 'test' not found"
	resp, _ := b.HandleRequest(context.Background(), request)
	if resp.Error().Error() != e {
		t.Errorf("Error must be '%s', get '%s'", e, resp.Error())
	}
}

func TestSecretsUpdateWithoutConfig(t *testing.T) {
	b, s := getTestBackend(t)

	request := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/test", saStoragePrefix),
		Data: map[string]interface{}{
			"namespace":            "test",
			"service-account-name": "test",
		},
		Storage: s,
	}
	assertNoErrorRequest(t, b, request)

	request = &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/test", secretsStoragePrefix),
		Data: map[string]interface{}{
			"ttl": 500,
		},
		Storage: s,
	}

	e := "Please configure plugin with 'config' path"
	resp, _ := b.HandleRequest(context.Background(), request)
	if resp.Error().Error() != e {
		t.Errorf("Error must be '%s', get '%s'", e, resp.Error())
	}
}

func TestSecretsUpdate(t *testing.T) {
	b, s := getTestBackend(t)

	request := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data: map[string]interface{}{
			"api-url": "https://localhost:8443",
			"token":   "123qwe",
			"CA":      "abc",
			"ttl":     100,
			"max-ttl": 200,
		},
		Storage: s,
	}
	assertNoErrorRequest(t, b, request)

	request = &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/test", saStoragePrefix),
		Data: map[string]interface{}{
			"namespace":            "test",
			"service-account-name": "test",
		},
		Storage: s,
	}
	assertNoErrorRequest(t, b, request)

	request = &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/test", secretsStoragePrefix),
		Data: map[string]interface{}{
			"ttl": 500,
		},
		Storage: s,
	}

	e := "Max TTL configured to '200', you try to create TTL '500'"
	resp, _ := b.HandleRequest(context.Background(), request)
	if resp.Error().Error() != e {
		t.Errorf("Error must be '%s', get '%s'", e, resp.Error())
	}

	request = &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/test", secretsStoragePrefix),
		Data:      nil,
		Storage:   s,
	}

	resp = assertNoErrorRequest(t, b, request)
	assertEquals(t, resp.Data["token"].(string), "test", "")
	assertEquals(t, resp.Data["namespace"].(string), "test", "")
	assertEquals(t, resp.Data["CA_base64"].(string), "test", "")
}
