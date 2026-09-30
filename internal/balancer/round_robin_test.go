package balancer

import (
	"github.com/italo/go-api-gateway/internal/upstream"
	"sync"
	"testing"
	"time"

	"github.com/italo/go-api-gateway/internal/config"
)

func TestEmpty(t *testing.T) {
	if _, e := new(RoundRobin).Select(nil); e != ErrNoUpstream {
		t.Fatalf("%v", e)
	}
}

func TestRoundRobinAlternates(t *testing.T) {
	cb := config.CircuitBreakerConfig{FailureThreshold: 2, SuccessThreshold: 1, OpenTimeout: time.Second, HalfOpenMax: 1}
	hc := config.HealthCheckConfig{HealthyThreshold: 1, UnhealthyThreshold: 1}
	a, err := upstream.New(config.UpstreamConfig{ID: "a", URL: "http://a.example"}, cb, hc)
	if err != nil {
		t.Fatal(err)
	}
	b, err := upstream.New(config.UpstreamConfig{ID: "b", URL: "http://b.example"}, cb, hc)
	if err != nil {
		t.Fatal(err)
	}
	r := new(RoundRobin)
	first, _ := r.Select([]*upstream.Upstream{a, b})
	second, _ := r.Select([]*upstream.Upstream{a, b})
	if first.ID == second.ID {
		t.Fatalf("did not alternate: %s", first.ID)
	}
}
func TestConcurrentEmpty(t *testing.T) {
	var wg sync.WaitGroup
	r := new(RoundRobin)
	for range 100 {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = r.Select([]*upstream.Upstream{}) }()
	}
	wg.Wait()
}
