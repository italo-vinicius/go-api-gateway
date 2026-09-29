package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterIsolatedAndRefills(t *testing.T) {
	now := time.Now()
	l := New(1, 1, time.Second, 10)
	if ok, _ := l.Allow("a", now); !ok {
		t.Fatal("first request denied")
	}
	if ok, _ := l.Allow("a", now); ok {
		t.Fatal("second request allowed")
	}
	if ok, _ := l.Allow("b", now); !ok {
		t.Fatal("different key denied")
	}
	if ok, _ := l.Allow("a", now.Add(time.Second)); !ok {
		t.Fatal("token did not refill")
	}
}
func BenchmarkLimiter(b *testing.B) {
	l := New(1000000, 1000000, time.Second, 100)
	for b.Loop() {
		l.Allow("client", time.Now())
	}
}
