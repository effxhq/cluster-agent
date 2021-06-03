package main

import (
	"github.com/effxhq/go-lifecycle"

	client_plugin "github.com/effxhq/cluster-agent/internal/plugins/client"
	kubernetes_plugin "github.com/effxhq/cluster-agent/internal/plugins/kubernetes"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
)

func main() {
	app := new(lifecycle.Application)

	client, err := client_plugin.NewHTTPClient()

	app.Initialize(
		zap_plugin.Plugin(),
		lifecycle.PluginFuncs{
			InitializeFunc: func(app *lifecycle.Application) error {
				return err
			},
		},
		kubernetes_plugin.Plugin(client),
	)

	app.Start()
}
