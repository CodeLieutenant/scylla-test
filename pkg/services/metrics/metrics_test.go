package metrics_test

import (
	"testing"
	"time"

	"github.com/CodeLieutenant/scylladbtest/pkg/services/metrics"
)

func TestMetrics(t *testing.T) {
	t.Parallel()

	m := metrics.New(2)

	m.Do(2 * time.Second)
	m.Do(2 * time.Second)

	metrics := m.Collect()

	if metrics.Count != 2 {
		t.Fatalf("expected 2, got %v", metrics.Count)
	}

	if metrics.AverageTime != 2*time.Second {
		t.Fatalf("expected 1s, got %v", metrics.AverageTime)
	}

	if metrics.P99 != 2*time.Second {
		t.Fatalf("expected 1s, got %v", metrics.P99)
	}
}

func TestMetricsInsertAfterFull(t *testing.T) {
	t.Parallel()

	m := metrics.New(2)

	m.Do(2 * time.Second)
	m.Do(2 * time.Second)
	m.Do(5 * time.Second)
	m.Do(5 * time.Second)

	metrics := m.Collect()

	if metrics.Count != 4 {
		t.Fatalf("expected 4, got %v", metrics.Count)
	}

	if metrics.AverageTime != (5 * time.Second / 2) {
		t.Fatalf("expected 5s, got %v", metrics.AverageTime)
	}

	if metrics.P99 != 5*time.Second {
		t.Fatalf("expected 2s, got %v", metrics.P99)
	}
}

func TestMetricsString(t *testing.T) {
	t.Parallel()

	m := metrics.New(2)

	m.Do(2 * time.Second)
	m.Do(2 * time.Second)

	metrics := m.Collect()

	expected := "\rMetrics:\r\nAverage Time: 2s\r\nCount: 2\r\nExecution Time: 4s\r\nP99: 2s\r\nP95: 2s\r\nP90: 2s\r\nP75: 2s\r\nP50: 2s"

	if metrics.String() != expected {
		t.Fatalf("expected %s, got %v", expected, metrics.String())
	}
}
