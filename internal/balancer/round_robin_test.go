package balancer

import (
	"github.com/italo/go-api-gateway/internal/upstream"
	"sync"
	"testing"
)

func TestEmpty(t *testing.T) {
	if _, e := new(RoundRobin).Select(nil); e != ErrNoUpstream {
		t.Fatalf("%v", e)
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
