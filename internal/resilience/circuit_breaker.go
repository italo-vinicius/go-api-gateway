package resilience

import (
	"sync"
	"time"
)

type State string

const (
	Closed   State = "closed"
	Open     State = "open"
	HalfOpen State = "half_open"
)

type CircuitBreaker struct {
	mu                                              sync.Mutex
	state                                           State
	failures, successes, probes                     int
	failureThreshold, successThreshold, halfOpenMax int
	openTimeout                                     time.Duration
	openedAt                                        time.Time
}

func NewCircuitBreaker(failure, success int, timeout time.Duration, probes int) *CircuitBreaker {
	return &CircuitBreaker{state: Closed, failureThreshold: failure, successThreshold: success, openTimeout: timeout, halfOpenMax: probes}
}
func (c *CircuitBreaker) Allow(now time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state == Open {
		if now.Sub(c.openedAt) < c.openTimeout {
			return false
		}
		c.state = HalfOpen
		c.successes = 0
		c.probes = 0
	}
	if c.state == HalfOpen {
		if c.probes >= c.halfOpenMax {
			return false
		}
		c.probes++
	}
	return true
}
func (c *CircuitBreaker) Success(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch c.state {
	case HalfOpen:
		c.probes--
		c.successes++
		if c.successes >= c.successThreshold {
			c.state = Closed
			c.failures = 0
			c.successes = 0
		}
	case Closed:
		c.failures = 0
	}
}
func (c *CircuitBreaker) Failure(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state == HalfOpen {
		if c.probes > 0 {
			c.probes--
		}
		c.open(now)
		return
	}
	if c.state == Closed {
		c.failures++
		if c.failures >= c.failureThreshold {
			c.open(now)
		}
	}
}
func (c *CircuitBreaker) open(now time.Time) {
	c.state = Open
	c.openedAt = now
	c.failures = 0
	c.successes = 0
	c.probes = 0
}
func (c *CircuitBreaker) State(now time.Time) State {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state == Open && now.Sub(c.openedAt) >= c.openTimeout {
		return HalfOpen
	}
	return c.state
}
