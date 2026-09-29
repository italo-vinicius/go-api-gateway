package resilience

import (
	"context"
	"math/rand/v2"
	"net/http"
	"time"
)

func SafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
func RetryableStatus(status int, allowed []int) bool {
	for _, s := range allowed {
		if status == s {
			return true
		}
	}
	return false
}
func Backoff(attempt int, initial, maximum time.Duration) time.Duration {
	d := initial
	for i := 0; i < attempt && d < maximum; i++ {
		d *= 2
	}
	if d > maximum {
		d = maximum
	}
	if d <= 0 {
		return 0
	}
	return d + time.Duration(rand.Int64N(int64(d)/2+1))
}
func Wait(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
