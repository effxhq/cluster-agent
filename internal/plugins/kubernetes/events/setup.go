package events

import (
	"context"

	"github.com/effxhq/cluster-agent/internal/appconf"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

func Setup(ctx context.Context, coreFactory *core.Factory) {
	// TODO: determine if events are enabled

	eventController := coreFactory.Core().V1().Event()
	eventController.Informer()
	eventController.Cache()

	eventController.OnChange(ctx, appconf.Name, func(id string, event *corev1.Event) (*corev1.Event, error) {
		if event == nil {
			// delete from cache?
			return nil, nil
		}

		zap_plugin.FromContext(ctx).Info("event", zap.String("id", id))
		return event, nil
	})
}
