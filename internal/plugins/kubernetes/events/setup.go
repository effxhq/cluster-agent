package events

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
	eventController := coreFactory.Core().V1().Event()
	eventController.Informer()
	eventController.Cache()

	eventController.OnChange(ctx, appconf.Name, func(id string, event *corev1.Event) (*corev1.Event, error) {
		if event == nil {
			// delete from cache?
			return nil, nil
		}

		event.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Event",
		}

		zap_plugin.FromContext(ctx).Info("event", zap.String("id", id))

		err := httpClient.PostResource(ctx, event)

		if err != nil {
			return nil, err
		}

		return event, nil
	})
}
