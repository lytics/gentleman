package retrier

import (
	"math"
	"time"
)

// ConstantBackoff generates a simple back-off strategy of retrying 'n' times, and waiting 'amount' time after each one.
func ConstantBackoff(n int, amount time.Duration) []time.Duration {
	ret := make([]time.Duration, n)
	for i := range ret {
		ret[i] = amount
	}
	return ret
}

// ExponentialBackoff generates a simple back-off strategy of retrying 'n' times, and doubling the amount of
// time waited after each one.
func ExponentialBackoff(n int, initialAmount time.Duration) []time.Duration {
	ret := make([]time.Duration, n)
	next := initialAmount
	for i := range ret {
		ret[i] = next
		next *= 2
	}
	return ret
}

// LogarithmicBackoff generates a simple back-off strategy of retrying 'n' times, following log(x+5)*time.Second as growing function.
// The first 10 intervals will be:
// 1s 2s 3s 4s 6s 8s 10s 12s 14s 16s 18s 20s 22s 24s 26s 28s 30s 33s 36s 39s 42s 45s 48s 51s 54s
func LogarithmicBackoff(n int, initialAmount time.Duration) []time.Duration {
	ret := make([]time.Duration, n)
	next := initialAmount
	for i := range ret {
		ret[i] = next
		next += time.Duration(math.Log(float64(i)+5.0)) * time.Second
	}
	return ret
}
