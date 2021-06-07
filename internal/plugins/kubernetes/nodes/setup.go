package nodes

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
	// TODO: determine if nodes are enabled

	nodeController := coreFactory.Core().V1().Node()
	nodeController.Informer()
	nodeController.Cache()

	nodeController.OnChange(ctx, appconf.Name, func(id string, node *corev1.Node) (*corev1.Node, error) {
		if node == nil {
			// delete from cache
			return nil, nil
		}

		node.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind: "Node",
		}

		zap_plugin.FromContext(ctx).Info("node", zap.String("id", id))

		err := httpClient.PostResource(ctx, node)

		if err != nil {
			return nil, err
		}

		return node, nil
	})
}
