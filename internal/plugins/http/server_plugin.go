package http_plugin

import (
	"bytes"
	"net"
	"net/http"
	"time"

	"github.com/effxhq/go-lifecycle"
	"github.com/pkg/errors"
)

func ServerPlugin() lifecycle.Plugin {
	return &lifecycle.PluginFuncs{
		StartFunc: func(app *lifecycle.Application) error {
			listener, err := net.Listen("tcp", "0.0.0.0:8080")
			if err != nil {
				return errors.Wrap(err, "failed to start tcp listener")
			}

			mux := http.NewServeMux()
			mux.Handle("/", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				http.ServeContent(writer, request, "", time.Now(), bytes.NewReader(nil))
			}))

			go http.Serve(listener, mux)
			return nil
		},
	}
}
