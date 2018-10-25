package backend

import (
	"encoding/base64"

	"github.com/hashicorp/errwrap"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getClientSet(c *config) (*kubernetes.Clientset, error) {
	data, err := base64.StdEncoding.DecodeString(c.CA)
	if err != nil {
		return nil, errwrap.Wrapf("Unable to create kubernetes client, unable to decode CA '{{err}}'", err)
	}

	clientConf := &rest.Config{
		Host: c.APIURL,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: data,
		},
		BearerToken: c.Token,
	}
	clientset, err := kubernetes.NewForConfig(clientConf)
	if err != nil {
		return nil, errwrap.Wrapf("Unable to create kubernetes client '{{err}}'", err)
	}
	return clientset, nil
}
