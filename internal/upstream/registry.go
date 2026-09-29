package upstream

import (
	"fmt"
	"github.com/italo/go-api-gateway/internal/config"
	"time"
)

type Registry struct{ routes map[string][]*Upstream }

func NewRegistry(routes []config.RouteConfig) (*Registry, error) {
	r := &Registry{routes: make(map[string][]*Upstream, len(routes))}
	for _, route := range routes {
		for _, c := range route.Upstreams {
			u, e := New(c, route.CircuitBreaker, route.HealthCheck)
			if e != nil {
				return nil, fmt.Errorf("upstream %s: %w", c.ID, e)
			}
			r.routes[route.Name] = append(r.routes[route.Name], u)
		}
	}
	return r, nil
}
func (r *Registry) Route(name string) []*Upstream { return r.routes[name] }
func (r *Registry) All() map[string][]*Upstream   { return r.routes }
func (r *Registry) Eligible(name string, now time.Time, excluded string) []*Upstream {
	var out []*Upstream
	for _, u := range r.routes[name] {
		if u.ID != excluded && u.Eligible(now) {
			out = append(out, u)
		}
	}
	return out
}
