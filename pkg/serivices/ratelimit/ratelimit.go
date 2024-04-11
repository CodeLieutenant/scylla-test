package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

var _ Limiter = &LeakyBucketLimiter{}

type (
	Limiter interface {
		Ready(ctx context.Context) error
	}
	LeakyBucketLimiter struct {
		pool          sync.Pool
		next          atomic.Int64
		singleRequest time.Duration
	}
)

func NewLeakyBucket(limit int64, rate time.Duration) Limiter {
	singleRequest := int64(rate) / limit

	return &LeakyBucketLimiter{
		singleRequest: time.Duration(singleRequest),
		pool: sync.Pool{
			New: func() any {
				t := time.NewTicker(1 * time.Second)
				t.Stop()
				return t
			},
		},
	}
}

func (l *LeakyBucketLimiter) Ready(ctx context.Context) error {
	newTimeOfNextPermissionIssue := int64(0)
	now := int64(0)
	ticker := l.pool.Get().(*time.Ticker)

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		// This implementation needs clock skew(max wait time)

		now = time.Now().UnixNano()
		timeOfNextPermissionIssue := l.next.Load()

		if timeOfNextPermissionIssue == 0 || now-timeOfNextPermissionIssue > int64(l.singleRequest) {
			newTimeOfNextPermissionIssue = now
		} else {
			newTimeOfNextPermissionIssue = timeOfNextPermissionIssue + int64(l.singleRequest)
		}

		if l.next.CompareAndSwap(timeOfNextPermissionIssue, newTimeOfNextPermissionIssue) {
			break
		}
	}

	if sleepDuration := newTimeOfNextPermissionIssue - now; sleepDuration > 0 {
		ticker.Reset(time.Duration(sleepDuration))
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-ticker.C:
			return nil
		}
	}

	return nil
}
