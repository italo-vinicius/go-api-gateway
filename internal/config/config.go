package config

import "time"

const MaxRetryAttempts = 5

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Admin   AdminConfig   `yaml:"admin"`
	Logging LoggingConfig `yaml:"logging"`
	Routes  []RouteConfig `yaml:"routes"`
}

type ServerConfig struct {
	Address           string        `yaml:"address"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes"`
}

type AdminConfig struct {
	Address     string `yaml:"address"`
	EnablePprof bool   `yaml:"enable_pprof"`
}
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}
type RouteConfig struct {
	Name           string               `yaml:"name"`
	Match          MatchConfig          `yaml:"match"`
	Upstreams      []UpstreamConfig     `yaml:"upstreams"`
	Timeout        time.Duration        `yaml:"timeout"`
	Retries        RetryConfig          `yaml:"retries"`
	RateLimit      RateLimitConfig      `yaml:"rate_limit"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	HealthCheck    HealthCheckConfig    `yaml:"health_check"`
}
type MatchConfig struct {
	PathPrefix  string `yaml:"path_prefix"`
	StripPrefix bool   `yaml:"strip_prefix"`
}
type UpstreamConfig struct {
	ID     string `yaml:"id"`
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight"`
}
type RetryConfig struct {
	Attempts       int           `yaml:"attempts"`
	InitialBackoff time.Duration `yaml:"initial_backoff"`
	MaxBackoff     time.Duration `yaml:"max_backoff"`
	RetryStatuses  []int         `yaml:"retry_statuses"`
}
type RateLimitConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Key      string        `yaml:"key"`
	Header   string        `yaml:"header"`
	Requests int           `yaml:"requests"`
	Burst    int           `yaml:"burst"`
	Window   time.Duration `yaml:"window"`
	MaxKeys  int           `yaml:"max_keys"`
}
type CircuitBreakerConfig struct {
	FailureThreshold int           `yaml:"failure_threshold"`
	SuccessThreshold int           `yaml:"success_threshold"`
	OpenTimeout      time.Duration `yaml:"open_timeout"`
	HalfOpenMax      int           `yaml:"half_open_max"`
}
type HealthCheckConfig struct {
	Path               string        `yaml:"path"`
	Interval           time.Duration `yaml:"interval"`
	Timeout            time.Duration `yaml:"timeout"`
	UnhealthyThreshold int           `yaml:"unhealthy_threshold"`
	HealthyThreshold   int           `yaml:"healthy_threshold"`
}

func (c *Config) ApplyDefaults() {
	if c.Server.Address == "" {
		c.Server.Address = ":8080"
	}
	if c.Admin.Address == "" {
		c.Admin.Address = ":9090"
	}
	if c.Server.ReadHeaderTimeout == 0 {
		c.Server.ReadHeaderTimeout = 5 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 60 * time.Second
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = 10 * time.Second
	}
	if c.Server.MaxHeaderBytes == 0 {
		c.Server.MaxHeaderBytes = 1 << 20
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	for i := range c.Routes {
		r := &c.Routes[i]
		if r.Timeout == 0 {
			r.Timeout = 2 * time.Second
		}
		if r.Retries.Attempts == 0 {
			r.Retries.Attempts = 1
		}
		if r.Retries.InitialBackoff == 0 {
			r.Retries.InitialBackoff = 100 * time.Millisecond
		}
		if r.Retries.MaxBackoff == 0 {
			r.Retries.MaxBackoff = 500 * time.Millisecond
		}
		if len(r.Retries.RetryStatuses) == 0 {
			r.Retries.RetryStatuses = []int{502, 503, 504}
		}
		if r.RateLimit.Key == "" {
			r.RateLimit.Key = "ip"
		}
		if r.RateLimit.MaxKeys == 0 {
			r.RateLimit.MaxKeys = 10000
		}
		if r.CircuitBreaker.FailureThreshold == 0 {
			r.CircuitBreaker.FailureThreshold = 5
		}
		if r.CircuitBreaker.SuccessThreshold == 0 {
			r.CircuitBreaker.SuccessThreshold = 2
		}
		if r.CircuitBreaker.OpenTimeout == 0 {
			r.CircuitBreaker.OpenTimeout = 30 * time.Second
		}
		if r.CircuitBreaker.HalfOpenMax == 0 {
			r.CircuitBreaker.HalfOpenMax = 1
		}
		if r.HealthCheck.Path == "" {
			r.HealthCheck.Path = "/health"
		}
		if r.HealthCheck.Interval == 0 {
			r.HealthCheck.Interval = 5 * time.Second
		}
		if r.HealthCheck.Timeout == 0 {
			r.HealthCheck.Timeout = time.Second
		}
		if r.HealthCheck.UnhealthyThreshold == 0 {
			r.HealthCheck.UnhealthyThreshold = 3
		}
		if r.HealthCheck.HealthyThreshold == 0 {
			r.HealthCheck.HealthyThreshold = 2
		}
		for j := range r.Upstreams {
			if r.Upstreams[j].Weight == 0 {
				r.Upstreams[j].Weight = 1
			}
		}
	}
}
