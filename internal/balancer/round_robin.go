package balancer

import (
	"errors"
	"github.com/italo/go-api-gateway/internal/upstream"
	"sync/atomic"
)

var ErrNoUpstream = errors.New("no eligible upstream")

type RoundRobin struct{ next atomic.Uint64 }

func (r *RoundRobin) Select(items []*upstream.Upstream) (*upstream.Upstream, error) {
	if len(items) == 0 {
		return nil, ErrNoUpstream
	}
	n := r.next.Add(1) - 1
	return items[n%uint64(len(items))], nil
}
