package link

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const writeTimeout = 5 * time.Second

// ClickRecorder queues clicks and writes them in the background so the redirect path never waits in the DB. Clicks are dropped when the queue is full.
type ClickRecorder struct {
	repo    Repository
	queue   chan Click
	wg      sync.WaitGroup
	dropped atomic.Int64
	onDrop  func()
}

func NewClickRecorder(repo Repository, bufferSize int) *ClickRecorder {
	r := &ClickRecorder{
		repo:  repo,
		queue: make(chan Click, bufferSize),
	}

	r.wg.Add(1)

	go r.run()
	return r
}

// Dropped reports how many clicks were discarded because the queue was full.
func (r *ClickRecorder) Dropped() int64 {
	return r.dropped.Load()
}

// OnDrop registers a callback invoked whenever a click is dropped.
func (r *ClickRecorder) OnDrop(fn func()) {
	r.onDrop = fn
}

// Record queues c. It never blocks: if the queue is full, the click is dropped.
func (r *ClickRecorder) Record(c Click) {
	select {
	case r.queue <- c:
	default:
		r.dropped.Add(1)
		if r.onDrop != nil {
			r.onDrop()
		}
	}
}

// Close stops accepting clicks and waits for queued ones to be written.
func (r *ClickRecorder) Close() {
	close(r.queue)
	r.wg.Wait()
}

func (r *ClickRecorder) run() {
	defer r.wg.Done()

	for c := range r.queue {
		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		if err := r.repo.RecordClick(ctx, &c); err != nil {
			log.Printf("recording click: %v", err)
		}
		cancel()
	}
}
