package pods

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
)

type HeartbeatFunc func(ctx context.Context, id string)

func NewHeartbeat(period time.Duration, fn HeartbeatFunc) *Heartbeat {
	return &Heartbeat{
		Clock:      clockwork.NewRealClock(),
		period:     period,
		fn:         fn,
		mu:         &sync.Mutex{},
		timestamps: make(map[string]int64),
		queue:      make([]string, 0),
	}
}

// Heartbeat will call the provided HeartbeatFunc on the supplied period for each enqueued element.
type Heartbeat struct {
	Clock clockwork.Clock

	period time.Duration
	fn     HeartbeatFunc

	mu         *sync.Mutex
	timestamps map[string]int64
	queue      []string
}

func (h *Heartbeat) heartbeat(ctx context.Context, id string) {
	// ensure we re-enqueue for later processing when we're done
	defer h.Enqueue(id, false)
	h.fn(ctx, id)
}

// Poll reads from the head of the queue until we reach an entry with a greater timestamp.
func (h *Heartbeat) Poll(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := h.Clock.Now().UnixNano()
	i := 0

	for ; h.timestamps[h.queue[i]] <= now; i++ {
		go h.heartbeat(ctx, h.queue[i])
	}

	ids := h.queue[:i]

	// "pop" off processed ids
	h.queue = h.queue[i:]
	for _, id := range ids {
		delete(h.timestamps, id)
	}
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

// Dequeue removes the heartbeat associated with the id from the list. This is roughly a log(n) runtime. There is _some_
// scanning involved, but it's only in the event that two heartbeats are scheduled at the same time.
func (h *Heartbeat) Dequeue(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	timeForID, ok := h.timestamps[id]
	if !ok {
		return
	}

	pos := sort.Search(len(h.queue), func(i int) bool {
		// smallest index where this is true
		return timeForID <= h.timestamps[h.queue[i]]
	})

	for {
		if h.queue[pos] == id {
			// found it! prune
			h.queue = append(h.queue[0:pos], h.queue[pos+1:]...)
			delete(h.timestamps, id)
			return

		} else if h.timestamps[h.queue[pos]] > timeForID {
			panic("expected entry to exist, but it doesn't appear to. system in a bad state")
		}

		pos++
	}
}

// Enqueue adds the provided id to the list of elements to process. If runNow is true, then the element is enqueued
// using the current timestamp rather than one in the future. This operation should be close to log(n) runtime.
func (h *Heartbeat) Enqueue(id string, runNow bool) {
	// dequeue to ensure the element is not already queued.
	h.Dequeue(id)

	h.mu.Lock()
	defer h.mu.Unlock()

	now := h.Clock.Now()
	if !runNow {
		now = now.Add(h.period)
	}
	timeForID := now.UnixNano()

	pos := sort.Search(len(h.queue), func(i int) bool {
		return timeForID < h.timestamps[h.queue[i]]
	})

	h.timestamps[id] = timeForID
	if pos == len(h.queue) {
		h.queue = append(h.queue, id)
	} else {
		h.queue = append(h.queue[:pos], append([]string{id}, h.queue[pos:]...)...)
	}
}
