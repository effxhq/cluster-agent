package nodes

import (
	"context"

	"github.com/effxhq/cluster-agent/internal/appconf"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

func Setup(ctx context.Context, coreFactory *core.Factory) {
	// TODO: determine if nodes are enabled

	nodeController := coreFactory.Core().V1().Node()
	nodeController.Informer()
	nodeController.Cache()

	nodeController.OnChange(ctx, appconf.Name, func(id string, node *corev1.Node) (*corev1.Node, error) {
		if node == nil {
			// delete from cache
			return nil, nil
		}

		zap_plugin.FromContext(ctx).Info("node", zap.String("id", id))
		return node, nil
	})
}
