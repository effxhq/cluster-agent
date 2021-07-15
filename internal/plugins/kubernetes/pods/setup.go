package pods

import (
	"context"
	"sync"
	"time"

	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/effxhq/cluster-agent/internal/appconf"
	client_plugin "github.com/effxhq/cluster-agent/internal/plugins/client"
	zap_plugin "github.com/effxhq/cluster-agent/internal/plugins/zap"
)

type tracker struct {
	mu         *sync.Mutex
	index      map[string]*corev1.Pod
	queue      []string
	httpClient client_plugin.HTTPClient
}

func (t *tracker) Enqueue(name string, pod *corev1.Pod) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.index[name] == nil {
		t.queue = append(t.queue, name)
	}
	t.index[name] = pod
}

func (t *tracker) Dequeue(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.index[name] != nil {
		delete(t.index, name)
	}
}

func (t *tracker) Once(ctx context.Context) error {
	logger := zap_plugin.FromContext(ctx)

	t.mu.Lock()
	if len(t.queue) == 0 {
		t.mu.Unlock()
		return nil
	}

	name := t.queue[0]
	pod := t.index[name]

	t.queue = t.queue[1:]
	if pod != nil {
		t.queue = append(t.queue, name)
	}
	t.mu.Unlock()

	if pod == nil {
		return nil
	}

	logger.Info("pod", zap.String("id", name))
	return t.httpClient.PostResource(ctx, pod)
}

func (t *tracker) Run(ctx context.Context) {
	logger := zap_plugin.FromContext(ctx)

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <- tick.C:
			err := t.Once(ctx)
			if err != nil {
				logger.Error("failed to heartbeat pod", zap.Error(err))
			}
		}
	}
}

func Setup(ctx context.Context, coreFactory *core.Factory, httpClient client_plugin.HTTPClient) {
	podController := coreFactory.Core().V1().Pod()

	t := &tracker{
		mu: &sync.Mutex{},
		index: make(map[string]*corev1.Pod),
		queue: make([]string, 0),
		httpClient: httpClient,
	}

	go t.Run(ctx)

	podController.OnChange(ctx, appconf.Name, func(id string, pod *corev1.Pod) (*corev1.Pod, error) {
		if pod == nil {
			t.Dequeue(id)
			return nil, nil
		}

		pod.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		}

		t.Enqueue(id, pod)
		return pod, nil
	})
}
