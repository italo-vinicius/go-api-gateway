package main

import (
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type state struct {
	name            string
	latency, jitter time.Duration
	failureRate     float64
	failureStatus   int
	healthy         atomic.Bool
}

func main() {
	s := &state{name: env("MOCK_NAME", "mock"), latency: duration("MOCK_LATENCY", 0), jitter: duration("MOCK_JITTER", 0), failureRate: floatEnv("MOCK_FAILURE_RATE", 0), failureStatus: intEnv("MOCK_FAILURE_STATUS", http.StatusServiceUnavailable)}
	s.healthy.Store(true)
	m := http.NewServeMux()
	m.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if !s.healthy.Load() {
			http.Error(w, "unhealthy", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(200)
	})
	m.HandleFunc("/info", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"name": s.name, "healthy": s.healthy.Load()})
	})
	m.HandleFunc("/admin/failure-mode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		var x struct {
			Healthy     *bool    `json:"healthy"`
			FailureRate *float64 `json:"failure_rate"`
			Latency     string   `json:"latency"`
		}
		if json.NewDecoder(r.Body).Decode(&x) != nil {
			http.Error(w, "invalid JSON", 400)
			return
		}
		if x.Healthy != nil {
			s.healthy.Store(*x.Healthy)
		}
		if x.FailureRate != nil {
			s.failureRate = *x.FailureRate
		}
		if x.Latency != "" {
			d, e := time.ParseDuration(x.Latency)
			if e != nil {
				http.Error(w, "invalid latency", 400)
				return
			}
			s.latency = d
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	m.HandleFunc("/", s.response)
	addr := env("MOCK_PORT", ":3000")
	slog.Info("mock service listening", "name", s.name, "address", addr)
	if e := http.ListenAndServe(addr, m); e != nil {
		slog.Error("server stopped", "error", e)
		os.Exit(1)
	}
}
func (s *state) response(w http.ResponseWriter, r *http.Request) {
	delay := s.latency
	if s.jitter > 0 {
		delay += time.Duration(rand.Int64N(int64(s.jitter) + 1))
	}
	if delay > 0 {
		time.Sleep(delay)
	}
	if !s.healthy.Load() || rand.Float64() < s.failureRate {
		http.Error(w, "simulated failure", s.failureStatus)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Mock-Instance", s.name)
	_ = json.NewEncoder(w).Encode(map[string]any{"instance": s.name, "method": r.Method, "path": r.URL.Path, "query": r.URL.Query()})
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func intEnv(k string, d int) int {
	v, e := strconv.Atoi(os.Getenv(k))
	if e == nil {
		return v
	}
	return d
}
func floatEnv(k string, d float64) float64 {
	v, e := strconv.ParseFloat(os.Getenv(k), 64)
	if e == nil {
		return v
	}
	return d
}
func duration(k string, d time.Duration) time.Duration {
	v, e := time.ParseDuration(os.Getenv(k))
	if e == nil {
		return v
	}
	return d
}
