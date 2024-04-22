package main

import (
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/plugin"
	"github.com/jetbrains-infra/vault-plugin-secrets-kubernetes/backend"

	"github.com/hashicorp/vault/api"
)

func main() {
	loggerOpts := &hclog.LoggerOptions{
		Name:  "vault-kubernetes",
		Level: hclog.Info,
	}

	pluginLogPath := "/tmp/vault-k8s.log"
	fp, err := os.OpenFile(pluginLogPath, os.O_WRONLY|os.O_CREATE, 0640)
	if err == nil {
		loggerOpts.Output = fp
		loggerOpts.Level = hclog.Trace
	} else {
		log.Fatalf("Failed to open plugin log file %s", pluginLogPath)
	}

	logger := hclog.New(loggerOpts)
	logger.Info("Plugin started")
	defer func() {
		logger.Info("Plugin stopped")
	}()
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	err = flags.Parse(os.Args[1:])
	if err != nil {
		logger.Error("Unable to parse arguments %+v, %s", os.Args[1:], err)
		os.Exit(1)
	}

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	err = plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: backend.NewFactory(logger),
		TLSProviderFunc:    tlsProviderFunc,
		Logger:             logger,
	})
	if err != nil {
		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}
}
