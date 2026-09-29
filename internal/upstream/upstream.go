package upstream

import (
	"github.com/italo/go-api-gateway/internal/config"
	"github.com/italo/go-api-gateway/internal/resilience"
	"net/url"
	"sync/atomic"
	"time"
)

type Upstream struct {
	ID        string
	URL       *url.URL
	healthy   atomic.Bool
	active    atomic.Int64
	failures  atomic.Int64
	successes atomic.Int64
	lastCheck atomic.Int64
	breaker   *resilience.CircuitBreaker
	cfg       config.HealthCheckConfig
}

func New(c config.UpstreamConfig, cb config.CircuitBreakerConfig, hc config.HealthCheckConfig) (*Upstream, error) {
	u, e := url.Parse(c.URL)
	if e != nil {
		return nil, e
	}
	x := &Upstream{ID: c.ID, URL: u, breaker: resilience.NewCircuitBreaker(cb.FailureThreshold, cb.SuccessThreshold, cb.OpenTimeout, cb.HalfOpenMax), cfg: hc}
	x.healthy.Store(true)
	return x, nil
}
func (u *Upstream) Eligible(now time.Time) bool                 { return u.healthy.Load() && u.breaker.Allow(now) }
func (u *Upstream) Healthy() bool                               { return u.healthy.Load() }
func (u *Upstream) Active() int64                               { return u.active.Load() }
func (u *Upstream) BreakerState(now time.Time) resilience.State { return u.breaker.State(now) }
func (u *Upstream) Begin() func()                               { u.active.Add(1); return func() { u.active.Add(-1) } }
func (u *Upstream) Result(ok bool, now time.Time) {
	if ok {
		u.breaker.Success(now)
	} else {
		u.breaker.Failure(now)
	}
}
func (u *Upstream) HealthResult(ok bool, now time.Time) {
	u.lastCheck.Store(now.UnixNano())
	if ok {
		u.failures.Store(0)
		if u.successes.Add(1) >= int64(u.cfg.HealthyThreshold) {
			u.healthy.Store(true)
		}
	} else {
		u.successes.Store(0)
		if u.failures.Add(1) >= int64(u.cfg.UnhealthyThreshold) {
			u.healthy.Store(false)
		}
	}
}
