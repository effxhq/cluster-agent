package pods

import (
	"context"

	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/effxhq/cluster-agent/internal/appconf"
	client_plugin "github.com/effxhq/cluster-agent/internal/plugins/client"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
)

func Setup(ctx context.Context, coreFactory *core.Factory, httpClient client_plugin.HTTPClient) {
	podController := coreFactory.Core().V1().Pod()
	podController.Informer()
	podController.Cache()

	podController.OnChange(ctx, appconf.Name, func(id string, pod *corev1.Pod) (*corev1.Pod, error) {
		if pod == nil {
			// delete from cache
			return nil, nil
		}

		pod.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		}

		zap_plugin.FromContext(ctx).Info("pod", zap.String("id", id))

		err := httpClient.PostResource(ctx, pod)

		if err != nil {
			return nil, err
		}

		return pod, nil
	})
}
