package pool

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/semaphore"
)

type WorkerPool struct {
	sem    *semaphore.Weighted
	cancel context.CancelFunc
	count  int
	mu     sync.Mutex
}

func New(count int) *WorkerPool {
	return &WorkerPool{
		sem:   semaphore.NewWeighted(int64(count)),
		count: count,
	}
}

func (w *WorkerPool) watcher(ctx context.Context) {
	defer fmt.Println("Exiting pool")
	for {
		if err := w.sem.Acquire(ctx, 1); err != nil {
			return
		}
	}
}

func (w *WorkerPool) worker(
	ctx context.Context,
	cb func(context.Context),
	errCb func(any),
) {
	defer func() {
		if err := recover(); err != nil && errCb != nil {
			errCb(err)
		}
	}()

	defer w.sem.Release(1)

	cb(ctx)
}

func (w *WorkerPool) Start(
	ctx context.Context,
	cb func(context.Context),
	errCb func(any),
) {
	ctx, cancel := context.WithCancel(ctx)

	w.mu.Lock()
	w.cancel = cancel
	w.mu.Unlock()

	for i := 0; i < w.count; i++ {
		if err := w.sem.Acquire(ctx, 1); err != nil {
			return
		}

		go w.worker(ctx, cb, errCb)
	}

	go w.watcher(ctx)
}

func (w *WorkerPool) Close() error {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil // after this, we can call start again with new context
	}
	w.mu.Unlock()

	// We can ignore error here, as context.Background
	_ = w.sem.Acquire(context.Background(), int64(w.count))
	defer w.sem.Release(int64(w.count))
	return nil
}
