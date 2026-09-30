package resilience

import (
	"context"
	"testing"
	"time"
)

func TestCircuitBreakerTransitions(t *testing.T) {
	now := time.Now()
	c := NewCircuitBreaker(2, 2, time.Second, 1)
	if !c.Allow(now) {
		t.Fatal("closed breaker denied")
	}
	c.Failure(now)
	c.Failure(now)
	if c.Allow(now) {
		t.Fatal("open breaker allowed")
	}
	later := now.Add(time.Second)
	if !c.Allow(later) {
		t.Fatal("half-open denied probe")
	}
	if c.Allow(later) {
		t.Fatal("half-open allowed extra probe")
	}
	c.Success(later)
	if !c.Allow(later) {
		t.Fatal("second probe denied")
	}
	c.Success(later)
	if c.State(later) != Closed {
		t.Fatal("breaker did not close")
	}
}
func TestWaitCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if Wait(ctx, time.Second) == nil {
		t.Fatal("wait did not cancel")
	}
}

func TestRetryHelpers(t *testing.T) {
	if !SafeMethod("GET") || SafeMethod("POST") || !RetryableStatus(503, []int{502, 503}) || RetryableStatus(500, []int{503}) {
		t.Fatal("retry classification failed")
	}
	if d := Backoff(10, time.Millisecond, 2*time.Millisecond); d < 2*time.Millisecond || d > 3*time.Millisecond {
		t.Fatalf("backoff out of range: %s", d)
	}
	if err := Wait(context.Background(), time.Nanosecond); err != nil {
		t.Fatal(err)
	}
}

func TestCircuitBreakerReopensAfterHalfOpenFailure(t *testing.T) {
	now := time.Now()
	c := NewCircuitBreaker(1, 1, time.Second, 1)
	c.Failure(now)
	if !c.Allow(now.Add(time.Second)) {
		t.Fatal("half-open probe denied")
	}
	c.Failure(now.Add(time.Second))
	if c.State(now.Add(time.Second)) != Open {
		t.Fatal("breaker did not reopen")
	}
}
