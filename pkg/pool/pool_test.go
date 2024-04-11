package pool_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/CodeLieutenant/scylladbtest/pkg/pool"
)

func TestPool_NeverStart_CloseCalled(t *testing.T) {
	t.Parallel()

	p := pool.New(1)

	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPool_Start(t *testing.T) {
	t.Parallel()

	p := pool.New(1)
	var counter atomic.Uint32

	go p.Start(
		context.Background(),
		pool.RunFunc(func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				counter.Add(1)

				time.Sleep(10 * time.Millisecond)
			}
		}),
	)

	time.Sleep(20 * time.Millisecond)

	if err := p.Close(); err != nil {
		t.Fatal(err)
	}

	if counter.Load() < 2 {
		t.Fatalf("counter is not working %d", counter.Load())
	}
}

func TestPool_Respawn(t *testing.T) {
	t.Parallel()

	p := pool.New(1)
	var counter atomic.Uint32

	go p.Start(
		context.Background(),
		pool.RunFunc(func(context.Context) error {
			counter.Add(1)
			time.Sleep(10 * time.Millisecond)
			return nil
		}),
	)

	time.Sleep(20 * time.Millisecond)

	if err := p.Close(); err != nil {
		t.Fatal(err)
	}

	if counter.Load() < 2 {
		t.Fatalf("counter is not working %d", counter.Load())
	}
}
