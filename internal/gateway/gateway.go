package gateway

import (
	"context"
	"github.com/google/uuid"
	"github.com/italo/go-api-gateway/internal/balancer"
	"github.com/italo/go-api-gateway/internal/config"
	"github.com/italo/go-api-gateway/internal/observability"
	"github.com/italo/go-api-gateway/internal/ratelimit"
	"github.com/italo/go-api-gateway/internal/resilience"
	"github.com/italo/go-api-gateway/internal/upstream"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Gateway struct {
	router   *Router
	registry *upstream.Registry
	balancer balancer.RoundRobin
	client   *http.Client
	logger   *slog.Logger
	metrics  *observability.Metrics
	limits   map[string]*ratelimit.Limiter
	mu       sync.Mutex
}

func New(cfg *config.Config, logger *slog.Logger, metrics *observability.Metrics) (*Gateway, error) {
	reg, e := upstream.NewRegistry(cfg.Routes)
	if e != nil {
		return nil, e
	}
	limits := map[string]*ratelimit.Limiter{}
	for _, r := range cfg.Routes {
		if r.RateLimit.Enabled {
			limits[r.Name] = ratelimit.New(r.RateLimit.Requests, r.RateLimit.Burst, r.RateLimit.Window, r.RateLimit.MaxKeys)
		}
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxIdleConns = 200
	tr.MaxIdleConnsPerHost = 50
	tr.IdleConnTimeout = 90 * time.Second
	return &Gateway{router: NewRouter(cfg.Routes), registry: reg, client: &http.Client{Transport: tr}, logger: logger, metrics: metrics, limits: limits}, nil
}
func (g *Gateway) Registry() *upstream.Registry { return g.registry }
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	id := r.Header.Get("X-Request-ID")
	if id == "" {
		id = uuid.NewString()
	}
	route, target, ok := g.router.Match(r.URL.Path)
	if !ok {
		WriteError(w, 404, "ROUTE_NOT_FOUND", "No route matches this request", id)
		return
	}
	g.metrics.Active.WithLabelValues(route.Name).Inc()
	defer g.metrics.Active.WithLabelValues(route.Name).Dec()
	defer func() { g.metrics.Duration.WithLabelValues(route.Name, r.Method).Observe(time.Since(start).Seconds()) }()
	if l := g.limits[route.Name]; l != nil {
		key := clientKey(r, route.RateLimit)
		allowed, retry := l.Allow(key, time.Now())
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(max(1, int(retry.Seconds()))))
			g.metrics.RateRejected.WithLabelValues(route.Name).Inc()
			g.finish(route.Name, r.Method, 429)
			WriteError(w, 429, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded", id)
			return
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), route.Timeout)
	defer cancel()
	attempts := route.Retries.Attempts
	safe := resilience.SafeMethod(r.Method)
	var lastErr error
	var previous string
	for attempt := 0; attempt < attempts; attempt++ {
		items := g.registry.Eligible(route.Name, time.Now(), previous)
		u, e := g.balancer.Select(items)
		if e != nil {
			g.finish(route.Name, r.Method, 503)
			WriteError(w, 503, "UPSTREAM_UNAVAILABLE", "No healthy upstream is currently available", id)
			return
		}
		previous = u.ID
		resp, e := g.call(ctx, r, u, target, id)
		if e == nil && !resilience.RetryableStatus(resp.StatusCode, route.Retries.RetryStatuses) {
			defer func() { _ = resp.Body.Close() }()
			copyResponse(w, resp, id)
			u.Result(resp.StatusCode < 500, time.Now())
			g.metrics.UpstreamRequests.WithLabelValues(route.Name, u.ID, strconv.Itoa(resp.StatusCode)).Inc()
			g.finish(route.Name, r.Method, resp.StatusCode)
			return
		}
		if resp != nil {
			_ = resp.Body.Close()
			g.metrics.UpstreamRequests.WithLabelValues(route.Name, u.ID, strconv.Itoa(resp.StatusCode)).Inc()
			u.Result(false, time.Now())
		} else {
			u.Result(false, time.Now())
		}
		lastErr = e
		if !safe || attempt+1 >= attempts {
			break
		}
		g.metrics.Retries.WithLabelValues(route.Name, "upstream_failure").Inc()
		if e = resilience.Wait(ctx, resilience.Backoff(attempt, route.Retries.InitialBackoff, route.Retries.MaxBackoff)); e != nil {
			lastErr = e
			break
		}
	}
	if ctx.Err() == context.DeadlineExceeded {
		g.finish(route.Name, r.Method, 504)
		WriteError(w, 504, "UPSTREAM_TIMEOUT", "Upstream request timed out", id)
	} else {
		_ = lastErr
		g.finish(route.Name, r.Method, 503)
		WriteError(w, 503, "UPSTREAM_UNAVAILABLE", "Upstream request failed", id)
	}
}
func (g *Gateway) finish(route, method string, status int) {
	g.metrics.Requests.WithLabelValues(route, method, strconv.Itoa(status)).Inc()
}
func (g *Gateway) call(ctx context.Context, in *http.Request, u *upstream.Upstream, path, id string) (*http.Response, error) {
	done := u.Begin()
	defer done()
	target := *u.URL
	target.Path = joinPath(u.URL.Path, path)
	target.RawQuery = in.URL.RawQuery
	out, e := http.NewRequestWithContext(ctx, in.Method, target.String(), in.Body)
	if e != nil {
		return nil, e
	}
	out.Header = in.Header.Clone()
	out.Header.Set("X-Request-ID", id)
	out.Header.Del("Connection")
	out.Header.Del("Keep-Alive")
	out.Header.Del("Proxy-Authenticate")
	out.Header.Del("Proxy-Authorization")
	out.Header.Del("TE")
	out.Header.Del("Trailer")
	out.Header.Del("Transfer-Encoding")
	out.Header.Del("Upgrade")
	host, _, _ := net.SplitHostPort(in.RemoteAddr)
	if host == "" {
		host = in.RemoteAddr
	}
	if prior := out.Header.Get("X-Forwarded-For"); prior != "" {
		host = prior + ", " + host
	}
	out.Header.Set("X-Forwarded-For", host)
	out.Header.Set("X-Forwarded-Host", in.Host)
	if in.TLS != nil {
		out.Header.Set("X-Forwarded-Proto", "https")
	} else {
		out.Header.Set("X-Forwarded-Proto", "http")
	}
	out.Host = target.Host
	return g.client.Do(out)
}
func joinPath(base, p string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(p, "/")
}
func copyResponse(w http.ResponseWriter, resp *http.Response, id string) {
	for k, v := range resp.Header {
		if !isHop(k) {
			w.Header()[k] = append([]string(nil), v...)
		}
	}
	w.Header().Set("X-Request-ID", id)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
func isHop(k string) bool {
	switch strings.ToLower(k) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	}
	return false
}
func clientKey(r *http.Request, c config.RateLimitConfig) string {
	if c.Key == "header" {
		return r.Header.Get(c.Header)
	}
	host, _, e := net.SplitHostPort(r.RemoteAddr)
	if e == nil {
		return host
	}
	return r.RemoteAddr
}
func (g *Gateway) CleanupLimits(now time.Time) {
	for _, l := range g.limits {
		l.Cleanup(now)
	}
}
func (g *Gateway) CloseIdleConnections() { g.client.CloseIdleConnections() }
func (g *Gateway) HealthRoute(_ context.Context) bool {
	for _, v := range g.registry.All() {
		for _, u := range v {
			if u.Healthy() {
				return true
			}
		}
	}
	return false
}

var _ = url.URL{}
