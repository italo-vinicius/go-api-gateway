package health

import (
	"context"
	"github.com/italo/go-api-gateway/internal/config"
	"github.com/italo/go-api-gateway/internal/upstream"
	"net/http"
	"sync"
	"time"
)

type Checker struct {
	client   *http.Client
	routes   []config.RouteConfig
	registry *upstream.Registry
	sem      chan struct{}
	wg       sync.WaitGroup
}

func New(routes []config.RouteConfig, reg *upstream.Registry) *Checker {
	return &Checker{client: &http.Client{Timeout: 5 * time.Second}, routes: routes, registry: reg, sem: make(chan struct{}, 16)}
}
func (c *Checker) Start(ctx context.Context) {
	for _, r := range c.routes {
		for _, u := range c.registry.Route(r.Name) {
			c.wg.Add(1)
			go c.loop(ctx, r.HealthCheck, u)
		}
	}
}
func (c *Checker) Wait() { c.wg.Wait() }
func (c *Checker) loop(ctx context.Context, h config.HealthCheckConfig, u *upstream.Upstream) {
	defer c.wg.Done()
	ticker := time.NewTicker(h.Interval)
	defer ticker.Stop()
	c.check(ctx, h, u)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.check(ctx, h, u)
		}
	}
}
func (c *Checker) check(parent context.Context, h config.HealthCheckConfig, u *upstream.Upstream) {
	select {
	case c.sem <- struct{}{}:
		defer func() { <-c.sem }()
	case <-parent.Done():
		return
	}
	ctx, cancel := context.WithTimeout(parent, h.Timeout)
	defer cancel()
	target := *u.URL
	target.Path = h.Path
	target.RawQuery = ""
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if e != nil {
		u.HealthResult(false, time.Now())
		return
	}
	resp, e := c.client.Do(req)
	if e != nil {
		u.HealthResult(false, time.Now())
		return
	}
	_ = resp.Body.Close()
	u.HealthResult(resp.StatusCode >= 200 && resp.StatusCode < 400, time.Now())
}
