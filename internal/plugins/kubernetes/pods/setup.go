package pods

import (
	"context"
	"strings"
	"time"

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
	cache := podController.Cache()

	logger := zap_plugin.FromContext(ctx)

	heartbeat := NewHeartbeat(time.Minute, func(ctx context.Context, id string) {
		parts := strings.SplitN(id, "/", 2)
		log := logger.With(
			zap.String("kind", "pod"),
			zap.String("namespace", parts[0]),
			zap.String("name", parts[0]),
		)

		pod, err := cache.Get(parts[0], parts[1])
		if err != nil {
			log.Error("failed to get element from cache", zap.Error(err))
			return
		}

		pod.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		}

		logger.Info("heartbeating pod", zap.String("id", id))

		err = httpClient.PostResource(ctx, pod)
		if err != nil {
			logger.Error("", zap.Error(err))
			return
		}
	})

	podController.OnChange(ctx, appconf.Name, func(id string, pod *corev1.Pod) (*corev1.Pod, error) {
		if pod == nil {
			// delete from cache
			heartbeat.Dequeue(id)
			return nil, nil
		}

		heartbeat.Enqueue(id, true)

		return pod, nil
	})
}
