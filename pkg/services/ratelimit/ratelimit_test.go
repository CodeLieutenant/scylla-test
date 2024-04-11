package ratelimit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/CodeLieutenant/scylladbtest/pkg/services/ratelimit"
)

func TestRateLimit(t *testing.T) {
	t.Parallel()
	limiter := ratelimit.NewLeakyBucket(1, 1*time.Second)

	if err := limiter.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	if err := limiter.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}

	if time.Since(start) < 1*time.Second {
		t.Fatal("rate limit not working")
	}
}

func TestCancel(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.NewLeakyBucket(1, 10*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	t.Cleanup(cancel)

	if err := limiter.Ready(ctx); err != nil {
		t.Fatal(err)
	}

	start := time.Now()

	err := limiter.Ready(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if time.Since(start) > 2*time.Second {
		t.Fatal("cancel is not working")
	}
}
