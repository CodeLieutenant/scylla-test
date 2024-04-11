package pool

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

type Runnable interface {
	Run(context.Context) error
	Error(any)
}

type RunFunc func(context.Context) error

var _ Runnable = RunFunc(nil)

func (f RunFunc) Run(ctx context.Context) error {
	return f(ctx)
}

func (f RunFunc) Error(any) {}

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

func Single() *WorkerPool {
	return New(1)
}

func (w *WorkerPool) watcher(ctx context.Context, runnable Runnable) {
	for {
		if err := w.sem.Acquire(ctx, 1); err != nil {
			return
		}

		go w.worker(ctx, runnable)
	}
}

func (w *WorkerPool) worker(ctx context.Context, runnable Runnable) {
	defer func() {
		if err := recover(); err != nil {
			runnable.Error(err)
		}
	}()

	defer w.sem.Release(1)

	if err := runnable.Run(ctx); err != nil {
		runnable.Error(err)
	}
}

func (w *WorkerPool) Start(ctx context.Context, runnable Runnable) {
	ctx, cancel := context.WithCancel(ctx)

	w.mu.Lock()
	w.cancel = cancel
	w.mu.Unlock()

	for i := 0; i < w.count; i++ {
		if err := w.sem.Acquire(ctx, 1); err != nil {
			return
		}

		go w.worker(ctx, runnable)
	}

	w.watcher(ctx, runnable)
}

func (w *WorkerPool) Close() error {
	w.mu.Lock()
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil // after this, we can call Start again with new context
	}
	w.mu.Unlock()

	// We can ignore error here, as context.Background never done<-
	_ = w.sem.Acquire(context.Background(), int64(w.count))
	defer w.sem.Release(int64(w.count))
	return nil
}
