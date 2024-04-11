package utils

import "runtime"

func Parallelism(input ...int) int {
	// We can use runtime.NumOfCPU but that is hardcoded value
	// and cannot be changed, GOMAXPROCS is ENV which can be abused
	// to increase the number of goroutines running
	// (especially in IO bounded environments)
	defaults := runtime.GOMAXPROCS(0)

	if len(input) == 0 || input[0] == 0 {
		return defaults
	}

	return min(defaults, input[0])
}
