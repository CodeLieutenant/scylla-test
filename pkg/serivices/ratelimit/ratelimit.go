package ratelimit

import (
	"context"
	"io"
	"time"
)

var _ Limiter = &LeakyBucketLimiter{}

type (
	Limiter interface {
		io.Closer
		Ready(ctx context.Context) <-chan time.Time
	}
	LeakyBucketLimiter struct {
		limit int
		dur   time.Duration
		chs   []chan time.Time
	}
)

func NewLeakyBucket(limit, parallel int, dur time.Duration) Limiter {
	return &LeakyBucketLimiter{
		limit: limit,
		dur:   dur,
		chs:   make([]chan time.Time, parallel),
	}
}

func (l *LeakyBucketLimiter) Ready(ctx context.Context) <-chan time.Time {
	ch := make(chan time.Time)
	l.chs = append(l.chs, ch)

	// Bad IMPL
	go func() {
		defer close(ch)
		ticker := time.NewTicker(l.dur)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ch <- time.Now()
			}
		}
	}()

	return ch
}

func (l *LeakyBucketLimiter) Close() error {
	return nil
}
