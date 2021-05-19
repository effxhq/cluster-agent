package deployments

import (
	"context"

	"github.com/effxhq/cluster-agent/internal/appconf"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apps"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
)

func Setup(ctx context.Context, appsFactory *apps.Factory) {
	// TODO: determine if deployments are enabled

	deploymentController := appsFactory.Apps().V1().Deployment()
	deploymentController.Informer()
	deploymentController.Cache()

	deploymentController.OnChange(ctx, appconf.Name, func(id string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
		if deployment == nil {
			// delete from cache
			return nil, nil
		}

		zap_plugin.FromContext(ctx).Info("deployment", zap.String("id", id))
		return deployment, nil
	})
}
