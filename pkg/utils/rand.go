package utils

import (
	"math/rand"
	"time"
)

func RandomDuration(minHours, maxHours int) time.Duration {
	delta := rand.Intn(maxHours*60-minHours*60+1) + minHours*60
	return time.Duration(delta) * time.Minute
}
