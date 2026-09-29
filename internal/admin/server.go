package admin

import (
	"encoding/json"
	"github.com/italo/go-api-gateway/internal/gateway"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"net/http/pprof"
	"time"
)

func Handler(g *gateway.Gateway, reg *prometheus.Registry, pprofEnabled bool, start time.Time) http.Handler {
	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	m.HandleFunc("/health/live", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	m.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		if !g.HealthRoute(r.Context()) {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	m.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		type us struct {
			ID      string `json:"id"`
			Healthy bool   `json:"healthy"`
			Active  int64  `json:"active_requests"`
			Circuit string `json:"circuit_state"`
		}
		routes := []map[string]any{}
		for name, items := range g.Registry().All() {
			list := []us{}
			for _, u := range items {
				list = append(list, us{u.ID, u.Healthy(), u.Active(), string(u.BreakerState(time.Now()))})
			}
			routes = append(routes, map[string]any{"name": name, "upstreams": list})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"version": "0.1.0", "uptime_seconds": int(time.Since(start).Seconds()), "routes": routes})
	})
	if pprofEnabled {
		m.HandleFunc("/debug/pprof/", pprof.Index)
		m.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		m.HandleFunc("/debug/pprof/profile", pprof.Profile)
		m.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		m.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	return m
}
