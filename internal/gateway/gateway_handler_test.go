package gateway

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/italo/go-api-gateway/internal/config"
	"github.com/italo/go-api-gateway/internal/observability"
	"github.com/prometheus/client_golang/prometheus"
)

func testGateway(t *testing.T, target string, rate config.RateLimitConfig, timeout time.Duration, attempts int) *Gateway {
	t.Helper()
	cfg := &config.Config{Routes: []config.RouteConfig{{
		Name: "api", Match: config.MatchConfig{PathPrefix: "/api"}, Upstreams: []config.UpstreamConfig{{ID: "one", URL: target}}, Timeout: timeout,
		Retries: config.RetryConfig{Attempts: attempts, InitialBackoff: time.Millisecond, MaxBackoff: time.Millisecond, RetryStatuses: []int{http.StatusServiceUnavailable}}, RateLimit: rate,
		CircuitBreaker: config.CircuitBreakerConfig{FailureThreshold: 5, SuccessThreshold: 1, OpenTimeout: time.Second, HalfOpenMax: 1},
		HealthCheck:    config.HealthCheckConfig{Path: "/health", Interval: time.Second, Timeout: time.Second, HealthyThreshold: 1, UnhealthyThreshold: 1},
	}}}
	cfg.ApplyDefaults()
	g, err := New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), observability.NewMetrics(prometheus.NewRegistry()))
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestGatewayErrorsForUnknownRoute(t *testing.T) {
	g := testGateway(t, "http://127.0.0.1:1", config.RateLimitConfig{}, time.Second, 1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/missing", nil)
	g.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), "ROUTE_NOT_FOUND") {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}

func TestGatewayForwardsAndSanitizesHeaders(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/items" || r.URL.Query().Get("q") != "x" || r.Header.Get("X-Forwarded-For") == "" || r.Header.Get("Connection") != "" {
			http.Error(w, "bad forwarding", http.StatusBadRequest)
			return
		}
		w.Header().Set("Connection", "close")
		w.Header().Set("X-Upstream", "ok")
		_, _ = w.Write([]byte("ok"))
	}))
	defer up.Close()
	g := testGateway(t, up.URL, config.RateLimitConfig{}, time.Second, 1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/items?q=x", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("Connection", "keep-alive")
	g.ServeHTTP(w, r)
	if w.Code != http.StatusOK || w.Body.String() != "ok" || w.Header().Get("X-Upstream") != "ok" || w.Header().Get("Connection") != "" || w.Header().Get("X-Request-ID") == "" {
		t.Fatalf("%d %#v %q", w.Code, w.Header(), w.Body.String())
	}
}

func TestGatewayRateLimits(t *testing.T) {
	g := testGateway(t, "http://127.0.0.1:1", config.RateLimitConfig{Enabled: true, Key: "ip", Requests: 1, Burst: 1, Window: time.Hour, MaxKeys: 10}, time.Second, 1)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api", nil)
		r.RemoteAddr = "127.0.0.1:1"
		g.ServeHTTP(w, r)
		if i == 0 && w.Code == http.StatusTooManyRequests {
			t.Fatal("first request limited")
		}
		if i == 1 && (w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") == "") {
			t.Fatalf("%d", w.Code)
		}
	}
}

func TestGatewayRetriesSafeRequests(t *testing.T) {
	var calls int
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			http.Error(w, "retry", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("recovered"))
	}))
	defer up.Close()
	g := testGateway(t, up.URL, config.RateLimitConfig{}, time.Second, 2)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api", nil))
	if w.Code != http.StatusOK || calls != 2 {
		t.Fatalf("status=%d calls=%d", w.Code, calls)
	}
}

func TestGatewayTimeout(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { time.Sleep(50 * time.Millisecond) }))
	defer up.Close()
	g := testGateway(t, up.URL, config.RateLimitConfig{}, time.Millisecond, 1)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api", nil))
	if w.Code != http.StatusGatewayTimeout || !strings.Contains(w.Body.String(), "UPSTREAM_TIMEOUT") {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}

func TestGatewayHelpers(t *testing.T) {
	if joinPath("/base/", "/path") != "/base/path" || !isHop("Connection") {
		t.Fatal("helper failure")
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:8"
	if clientKey(r, config.RateLimitConfig{Key: "ip"}) != "10.0.0.1" {
		t.Fatal("ip key")
	}
	r.Header.Set("X-Key", "x")
	if clientKey(r, config.RateLimitConfig{Key: "header", Header: "X-Key"}) != "x" {
		t.Fatal("header key")
	}
}
