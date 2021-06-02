package statefulsets

import (
	"context"

	"github.com/effxhq/cluster-agent/internal/appconf"
	client_plugin "github.com/effxhq/cluster-agent/internal/plugins/client"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apps"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
)

func Setup(ctx context.Context, appsFactory *apps.Factory, httpClient client_plugin.HTTPClient) {
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

		err := httpClient.PostResource(ctx, statefulset)

		if err != nil {
			return nil, err
		}

		return statefulset, nil
	})
}
