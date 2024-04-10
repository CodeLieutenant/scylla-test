package utils

import "runtime"

func Parallelism(input ...int) int {
	max := runtime.GOMAXPROCS(0)

	if len(input) == 0 {
		return max
	}

	if input[0] > max {
		return max
	}

	return input[0]
}
