package deployments

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
	// TODO: determine if deployments are enabled

	deploymentController := appsFactory.Apps().V1().Deployment()
	deploymentController.Informer()
	deploymentController.Cache()

	deploymentController.OnChange(ctx, appconf.Name, func(id string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
		if deployment == nil {
			// delete from cache
			return nil, nil
		}

		deployment.TypeMeta = metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind: "Deployment",
		}

		zap_plugin.FromContext(ctx).Info("deployment", zap.String("id", id))

		err := httpClient.PostResource(ctx, deployment)

		if err != nil {
			return nil, err
		}

		return deployment, nil
	})
}
