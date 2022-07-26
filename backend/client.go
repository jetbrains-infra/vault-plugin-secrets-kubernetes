package backend

import (
	"encoding/base64"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getClientSet(c *config) (*kubernetes.Clientset, error) {
	data, err := base64.StdEncoding.DecodeString(c.CA)
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes client, unable to decode CA: %s", err)
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
		return nil, fmt.Errorf("unable to create kubernetes client: %s", err)
	}
	return clientset, nil
}
