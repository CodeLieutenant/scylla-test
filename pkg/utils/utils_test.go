package utils_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/CodeLieutenant/scylladbtest/pkg/utils"
)

func TestParallelism(t *testing.T) {
	t.Parallel()
	p := runtime.GOMAXPROCS(0)

	data := []struct {
		value    int
		expected int
	}{
		{value: 0, expected: p},
		{value: 1, expected: 1},
		{value: 100, expected: p},
	}

	for _, d := range data {
		t.Run(fmt.Sprintf("With Value %d", d.value), func(t *testing.T) {
			t.Parallel()
			value := utils.Parallelism(d.value)

			if d.expected != value {
				t.Fatalf("Expected %d, got %d", d.expected, value)
			}
		})
	}
}

func TestParallelism_Defaults(t *testing.T) {
	t.Parallel()
	value := utils.Parallelism()

	if runtime.GOMAXPROCS(0) != value {
		t.Fatalf("Expected %d, got %d", runtime.GOMAXPROCS(0), value)
	}
}
