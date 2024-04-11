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
	newTimeTicket := int64(0)
	now := int64(0)
	ticker := l.pool.Get().(*time.Ticker)

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		now = time.Now().UnixNano()
		timeTicket := l.next.Load()

		if timeTicket == 0 || now-timeTicket > int64(l.singleRequest) {
			newTimeTicket = now
		} else {
			newTimeTicket = timeTicket + int64(l.singleRequest)
		}

		if l.next.CompareAndSwap(timeTicket, newTimeTicket) {
			break
		}
	}

	if sleepDuration := time.Duration(newTimeTicket - now); sleepDuration > 0 {
		ticker.Reset(sleepDuration)
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-ticker.C:
			return nil
		}
	}

	return nil
}
