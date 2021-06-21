package daemonsets

import (
	"context"

	"github.com/rancher/wrangler-api/pkg/generated/controllers/apps"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/effxhq/cluster-agent/internal/appconf"
	client_plugin "github.com/effxhq/cluster-agent/internal/plugins/client"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
)

func Setup(ctx context.Context, appsFactory *apps.Factory, httpClient client_plugin.HTTPClient) {
	allowed, err := httpClient.IsResourceAllowed(ctx, "daemon_sets")
	if err != nil {
		zap_plugin.FromContext(ctx).Info("daemonset", zap.Error(err))
	}

	if !allowed {
		return
	}

	daemonSetController := appsFactory.Apps().V1().DaemonSet()
	daemonSetController.Informer()
	daemonSetController.Cache()

	daemonSetController.OnChange(ctx, appconf.Name, func(id string, daemonset *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
		if daemonset == nil {
			// delete from cache
			return nil, nil
		}

		daemonset.TypeMeta = metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "DaemonSet",
		}

		zap_plugin.FromContext(ctx).Info("daemonset", zap.String("id", id))

		err := httpClient.PostResource(ctx, daemonset)

		if err != nil {
			return nil, err
		}

		return daemonset, nil
	})
}
