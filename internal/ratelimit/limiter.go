package ratelimit

import (
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
}
type Limiter struct {
	mu                   sync.Mutex
	buckets              map[string]bucket
	requests, burst, max int
	window               time.Duration
}

func New(requests, burst int, window time.Duration, max int) *Limiter {
	return &Limiter{buckets: map[string]bucket{}, requests: requests, burst: burst, window: window, max: max}
}
func (l *Limiter) Allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	if !ok {
		if len(l.buckets) >= l.max {
			return false, l.window
		}
		b = bucket{tokens: float64(l.burst), last: now}
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = min(float64(l.burst), b.tokens+elapsed*float64(l.requests)/l.window.Seconds())
	b.last = now
	if b.tokens < 1 {
		l.buckets[key] = b
		return false, time.Duration((1-b.tokens)*l.window.Seconds()/float64(l.requests)) * time.Second
	}
	b.tokens--
	l.buckets[key] = b
	return true, 0
}
func (l *Limiter) Cleanup(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, b := range l.buckets {
		if now.Sub(b.last) > l.window*2 {
			delete(l.buckets, k)
		}
	}
}
