package daemonsets

import (
	"context"

	"github.com/effxhq/cluster-agent/internal/appconf"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apps"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
)

func Setup(ctx context.Context, appsFactory *apps.Factory) {
	// TODO: determine if daemonsets are enabled

	daemonSetController := appsFactory.Apps().V1().DaemonSet()
	daemonSetController.Informer()
	daemonSetController.Cache()

	daemonSetController.OnChange(ctx, appconf.Name, func(id string, daemonset *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
		if daemonset == nil {
			// delete from cache
			return nil, nil
		}

		zap_plugin.FromContext(ctx).Info("daemonset", zap.String("id", id))
		return daemonset, nil
	})
}
