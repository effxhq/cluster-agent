package pods

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
)

type item struct {
	cancelled bool
	timestamp int64
}

type HeartbeatFunc func(ctx context.Context, id string)

func NewHeartbeat(period time.Duration, fn HeartbeatFunc) *Heartbeat {
	return &Heartbeat{
		Clock:  clockwork.NewRealClock(),
		period: period,
		fn:     fn,
		mu:     &sync.Mutex{},
		items:  make(map[string]*item),
		queue:  make([]string, 0),
	}
}

// Heartbeat will call the provided HeartbeatFunc on the supplied period for each enqueued element.
type Heartbeat struct {
	Clock clockwork.Clock

	period time.Duration
	fn     HeartbeatFunc

	mu    *sync.Mutex
	items map[string]*item
	queue []string
}

// Dequeue removes the heartbeat associated with the id from the list. This is roughly a log(n) runtime. There is _some_
// scanning involved, but it's only in the event that two heartbeats are scheduled at the same time.
func (h *Heartbeat) Dequeue(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.items[id]; !ok {
		// doesn't exist
		return
	}

	h.items[id].cancelled = true
}

// Enqueue adds the provided id to the list of elements to process. If runNow is true, then the element is enqueued
// using the current timestamp rather than one in the future. This operation should be close to log(n) runtime.
func (h *Heartbeat) Enqueue(ctx context.Context, id string) {
	h.mu.Lock()

	_, alreadyExists := h.items[id]
	if !alreadyExists {
		h.items[id] = &item{
			cancelled: false,
			timestamp: 0,
		}
	}

	h.mu.Unlock()

	if alreadyExists {
		return
	}

	h.heartbeat(ctx, id)
}

func (h *Heartbeat) enqueue(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.items[id].cancelled {
		delete(h.items, id)
	} else {
		timeForID := h.Clock.Now().Add(h.period).UnixNano()

		pos := sort.Search(len(h.queue), func(i int) bool {
			return timeForID < h.items[h.queue[i]].timestamp
		})

		h.items[id].timestamp = timeForID
		if pos == len(h.queue) {
			h.queue = append(h.queue, id)
		} else {
			h.queue = append(h.queue[:pos], append([]string{id}, h.queue[pos:]...)...)
		}
	}
}

func (h *Heartbeat) heartbeat(ctx context.Context, id string) {
	defer h.enqueue(id)

	if !h.items[id].cancelled {
		h.fn(ctx, id)
	}
}

// Poll reads from the head of the queue until we reach an entry with a greater timestamp.
func (h *Heartbeat) Poll(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := h.Clock.Now().UnixNano()
	i := 0

	for ; h.items[h.queue[i]].timestamp <= now; i++ {
		go h.heartbeat(ctx, h.queue[i])
	}

	h.queue = h.queue[i:]
}

// Start forks a go routine until the provided context is cancelled. Every 200 ms, it attempts to poll the underlying
// queue for entries that need to be processed.
func (h *Heartbeat) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.Poll(ctx)
			}
		}
	}()
}
