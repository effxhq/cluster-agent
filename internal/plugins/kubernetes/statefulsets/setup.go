package statefulsets

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
	allowed, err := httpClient.IsResourceAllowed(ctx, "stateful_sets")
	if err != nil {
		zap_plugin.FromContext(ctx).Info("statefulsets", zap.Error(err))
	}

	if !allowed {
		return
	}

	statefulSetController := appsFactory.Apps().V1().StatefulSet()
	statefulSetController.Informer()
	statefulSetController.Cache()

	statefulSetController.OnChange(ctx, appconf.Name, func(id string, statefulset *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
		if statefulset == nil {
			// delete from cache
			return nil, nil
		}

		statefulset.TypeMeta = metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		}

		zap_plugin.FromContext(ctx).Info("statefulset", zap.String("id", id))

		err := httpClient.PostResource(ctx, statefulset)

		if err != nil {
			return nil, err
		}

		return statefulset, nil
	})
}
