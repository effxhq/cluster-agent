package kubernetes_plugin

import (
	"github.com/effxhq/go-lifecycle"
	"github.com/pkg/errors"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apps"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	"github.com/rancher/wrangler/pkg/start"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	client_plugin "github.com/effxhq/cluster-agent/internal/plugins/client"
	"github.com/effxhq/cluster-agent/internal/plugins/kubernetes/daemonsets"
	"github.com/effxhq/cluster-agent/internal/plugins/kubernetes/deployments"
	"github.com/effxhq/cluster-agent/internal/plugins/kubernetes/events"
	"github.com/effxhq/cluster-agent/internal/plugins/kubernetes/nodes"
	"github.com/effxhq/cluster-agent/internal/plugins/kubernetes/pods"
	"github.com/effxhq/cluster-agent/internal/plugins/kubernetes/statefulsets"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel:       "level",
			logrus.FieldKeyTime:        "ts",
			logrus.FieldKeyFunc:        "caller",
			logrus.FieldKeyMsg:         "msg",
			logrus.FieldKeyLogrusError: "err",
		},
	})
}

func Plugin(httpClient client_plugin.HTTPClient) lifecycle.Plugin {
	var kubeClient *kubernetes.Clientset
	var appsFactory *apps.Factory
	var coreFactory *core.Factory

	return &lifecycle.PluginFuncs{
		InitializeFunc: func(app *lifecycle.Application) error {
			// TODO: determine if kubernetes is enabled

			// replace with in cluster configuration now that we
			cfg, err := rest.InClusterConfig()
			if err != nil {
				logrus.Fatalf("Error building kubeconfig: %s", err.Error())
			}

			kubeClient, err = kubernetes.NewForConfig(cfg)
			if err != nil {
				return errors.Wrap(err, "failed to setup kubernetes client")
			}

			appsFactory, err = apps.NewFactoryFromConfig(cfg)
			if err != nil {
				return errors.Wrap(err, "failed to setup kubernetes client")
			}

			coreFactory, err = core.NewFactoryFromConfig(cfg)
			if err != nil {
				return errors.Wrap(err, "failed to start core client")
			}

			ctx := app.Context()

			daemonsets.Setup(ctx, appsFactory, httpClient)
			deployments.Setup(ctx, appsFactory, httpClient)
			statefulsets.Setup(ctx, appsFactory, httpClient)
			events.Setup(ctx, coreFactory, httpClient)
			nodes.Setup(ctx, coreFactory, httpClient)
			pods.Setup(ctx, coreFactory, httpClient)

			return nil
		},
		StartFunc: func(app *lifecycle.Application) error {
			// nothing to start
			if appsFactory == nil {
				return nil
			}

			// start shared informers
			err := start.All(app.Context(), 2, appsFactory, coreFactory)
			if err != nil {
				return errors.Wrap(err, "failed to start shared informers")
			}

			return nil
		},
	}
}
