package kubernetes_plugin

import (
	"github.com/effxhq/go-lifecycle"
)

func Plugin() lifecycle.Plugin {
	return &lifecycle.PluginFuncs{}
}
