package utils

import (
	"math/rand"
	"time"
)

func RandomIntRange(min, max int) int {
	if max < min {
		return min
	}
	return rand.Intn(max-min+1) + min
}

func RandomIndex(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}

func RandomDurationMinutes(minMinutes, maxMinutes int) time.Duration {
	delta := RandomIntRange(minMinutes, maxMinutes)
	return time.Duration(delta) * time.Minute
}
