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
