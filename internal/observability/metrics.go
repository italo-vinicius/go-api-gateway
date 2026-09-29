package observability

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	Requests         *prometheus.CounterVec
	Duration         *prometheus.HistogramVec
	UpstreamRequests *prometheus.CounterVec
	Retries          *prometheus.CounterVec
	RateRejected     *prometheus.CounterVec
	Active           *prometheus.GaugeVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{Requests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "go_api_gateway_requests_total", Help: "Gateway requests."}, []string{"route", "method", "status"}), Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "go_api_gateway_request_duration_seconds", Help: "Gateway request duration."}, []string{"route", "method"}), UpstreamRequests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "go_api_gateway_upstream_requests_total", Help: "Upstream requests."}, []string{"route", "upstream", "status"}), Retries: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "go_api_gateway_retries_total", Help: "Gateway retries."}, []string{"route", "reason"}), RateRejected: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "go_api_gateway_rate_limit_rejections_total", Help: "Rate-limit rejections."}, []string{"route"}), Active: prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "go_api_gateway_active_requests", Help: "Active requests."}, []string{"route"})}
	reg.MustRegister(m.Requests, m.Duration, m.UpstreamRequests, m.Retries, m.RateRejected, m.Active)
	return m
}
