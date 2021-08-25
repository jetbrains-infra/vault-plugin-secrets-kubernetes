package backend

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

const (
	defaultLeaseTTLHr = 1
	maxLeaseTTLHr     = 12
)

func getTestBackend(t *testing.T) (logical.Backend, logical.Storage) {
	b := New()

	c := &logical.BackendConfig{
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: defaultLeaseTTLHr * time.Hour,
			MaxLeaseTTLVal:     maxLeaseTTLHr * time.Hour,
		},
		StorageView: &logical.InmemStorage{},
	}
	b.testMode = true
	err := b.Setup(context.Background(), c)
	if err != nil {
		t.Fatalf("unable to create backend: %v", err)
	}

	return b, c.StorageView
}

func assertNoErrorRequest(t *testing.T, b logical.Backend, r *logical.Request) *logical.Response {
	resp, err := b.HandleRequest(context.Background(), r)
	if err != nil {
		t.Errorf("Should not errors here, but get error '%s'", err)
	} else if resp != nil && resp.IsError() {
		t.Errorf("Should not errors here, but get error '%s'", resp.Error())
	}
	return resp
}

func assertEquals(t *testing.T, value1 interface{}, value2 interface{}, message string) {
	if value1 != value2 {
		t.Errorf("%s != %s, %s", value1, value2, message)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Error should be nil, but get %s", err.Error())
	}
}
