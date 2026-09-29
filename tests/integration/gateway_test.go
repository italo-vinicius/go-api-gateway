package integration

import (
	"github.com/italo/go-api-gateway/internal/config"
	"github.com/italo/go-api-gateway/internal/gateway"
	"github.com/italo/go-api-gateway/internal/observability"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGatewayForwardsRequest(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/item" || r.URL.Query().Get("x") != "1" || r.Header.Get("X-Request-ID") == "" {
			http.Error(w, "bad request", 400)
			return
		}
		b, _ := io.ReadAll(r.Body)
		w.Write(append([]byte(r.Method+":"), b...))
	}))
	defer up.Close()
	cfg := &config.Config{Routes: []config.RouteConfig{{Name: "api", Match: config.MatchConfig{PathPrefix: "/api"}, Upstreams: []config.UpstreamConfig{{ID: "one", URL: up.URL}}, Timeout: time.Second, Retries: config.RetryConfig{Attempts: 1, InitialBackoff: time.Millisecond, MaxBackoff: time.Millisecond}, CircuitBreaker: config.CircuitBreakerConfig{FailureThreshold: 2, SuccessThreshold: 1, OpenTimeout: time.Second, HalfOpenMax: 1}, HealthCheck: config.HealthCheckConfig{Path: "/", Interval: time.Second, Timeout: time.Second, HealthyThreshold: 1, UnhealthyThreshold: 1}}}}
	cfg.ApplyDefaults()
	reg := prometheus.NewRegistry()
	g, e := gateway.New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), observability.NewMetrics(reg))
	if e != nil {
		t.Fatal(e)
	}
	r := httptest.NewRequest(http.MethodPost, "http://gateway/api/item?x=1", strings.NewReader("body"))
	r.RemoteAddr = "127.0.0.1:1000"
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "POST:body" {
		t.Fatalf("%d %q", w.Code, w.Body.String())
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request id")
	}
}
