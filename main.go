package main

import (
	"github.com/effxhq/go-lifecycle"

	kubernetes_plugin "github.com/effxhq/cluster-agent/internal/plugins/kubernetes"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
)

func main() {
	app := new(lifecycle.Application)

	app.Initialize(
		zap_plugin.Plugin(),
		kubernetes_plugin.Plugin(),
	)

	app.Start()
}
