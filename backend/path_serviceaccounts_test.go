package backend

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/vault/logical"
)

func TestRoleCreate(t *testing.T) {
	b, s := getTestBackend(t)

	request := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      fmt.Sprintf("%s/test", saStoragePrefix),
		Data:      map[string]interface{}{},
		Storage:   s,
	}

	e := "namespace is required"
	resp, _ := b.HandleRequest(context.Background(), request)
	if resp.Error().Error() != e {
		t.Errorf("Error must be '%s', get '%s'", e, resp.Error())
	}

	request.Data = map[string]interface{}{
		"namespace": "test",
	}

	e = "service-account-name is required"
	resp, _ = b.HandleRequest(context.Background(), request)
	if resp.Error().Error() != e {
		t.Errorf("Error must be '%s', get '%s'", e, resp.Error())
	}

	request.Data = map[string]interface{}{
		"namespace":            "test",
		"service-account-name": "test",
	}

	assertNoErrorRequest(t, b, request)
}

func TestRoleUpdate(t *testing.T) {
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

	request.Data = map[string]interface{}{
		"namespace": "test1",
	}
	assertNoErrorRequest(t, b, request)
	resp := assertNoErrorRequest(t, b, &logical.Request{
		Operation: logical.ReadOperation,
		Path:      fmt.Sprintf("%s/test", saStoragePrefix),
		Data:      nil,
		Storage:   s,
	})
	assertEquals(t, resp.Data["namespace"], "test1", "Namespace should be updated from test to test1")
	assertEquals(t, resp.Data["service-account-name"], "test", "ServiceAccount-name should not be updated")

	request.Data = map[string]interface{}{
		"service-account-name": "test1",
	}
	assertNoErrorRequest(t, b, request)
	resp = assertNoErrorRequest(t, b, &logical.Request{
		Operation: logical.ReadOperation,
		Path:      fmt.Sprintf("%s/test", saStoragePrefix),
		Data:      nil,
		Storage:   s,
	})

	assertEquals(t, resp.Data["service-account-name"], "test1", "ServiceAccount-name should be updated from test to test1")
	assertEquals(t, resp.Data["namespace"], "test1", "Namespace should not be updated")
}

func TestRoleReadNotFound(t *testing.T) {
	b, s := getTestBackend(t)

	request := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      fmt.Sprintf("%s/notfound", saStoragePrefix),
		Data:      nil,
		Storage:   s,
	}

	resp, err := b.HandleRequest(context.Background(), request)
	if resp != nil || err != nil {
		t.Errorf("Response and error should be nil, get response: '%v', err: '%s'", resp, err)
	}
}

func TestRoleDelete(t *testing.T) {
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
		Operation: logical.DeleteOperation,
		Path:      fmt.Sprintf("%s/test", saStoragePrefix),
		Data:      nil,
		Storage:   s,
	}

	assertNoErrorRequest(t, b, request)

	request = &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      fmt.Sprintf("%s/notfound", saStoragePrefix),
		Data:      nil,
		Storage:   s,
	}

	assertNoErrorRequest(t, b, request)
}
