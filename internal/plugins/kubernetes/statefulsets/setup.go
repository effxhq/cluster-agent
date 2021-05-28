package statefulsets

import (
	"context"

	"github.com/effxhq/cluster-agent/internal/appconf"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apps"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
)

func Setup(ctx context.Context, appsFactory *apps.Factory) {
	// TODO: determine if statefulsets are enabled

	statefulSetController := appsFactory.Apps().V1().StatefulSet()
	statefulSetController.Informer()
	statefulSetController.Cache()

	statefulSetController.OnChange(ctx, appconf.Name, func(id string, statefulset *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
		if statefulset == nil {
			// delete from cache
			return nil, nil
		}

		zap_plugin.FromContext(ctx).Info("statefulset", zap.String("id", id))
		return statefulset, nil
	})
}
