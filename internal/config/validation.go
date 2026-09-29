package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func (c Config) Validate() error {
	var errs []error
	if c.Server.Address == c.Admin.Address {
		errs = append(errs, errors.New("server and admin addresses must differ"))
	}
	names := map[string]bool{}
	ids := map[string]bool{}
	for _, r := range c.Routes {
		p := "route " + r.Name
		if r.Name == "" {
			errs = append(errs, errors.New("route name is required"))
		} else if names[r.Name] {
			errs = append(errs, fmt.Errorf("duplicate route name %q", r.Name))
		}
		names[r.Name] = true
		if !strings.HasPrefix(r.Match.PathPrefix, "/") {
			errs = append(errs, fmt.Errorf("%s: path_prefix must start with /", p))
		}
		if len(r.Upstreams) == 0 {
			errs = append(errs, fmt.Errorf("%s: at least one upstream is required", p))
		}
		if r.Timeout <= 0 {
			errs = append(errs, fmt.Errorf("%s: timeout must be positive", p))
		}
		if r.Retries.Attempts < 1 || r.Retries.Attempts > MaxRetryAttempts {
			errs = append(errs, fmt.Errorf("%s: retry attempts must be between 1 and %d", p, MaxRetryAttempts))
		}
		if r.Retries.InitialBackoff <= 0 || r.Retries.MaxBackoff <= 0 {
			errs = append(errs, fmt.Errorf("%s: retry backoffs must be positive", p))
		}
		if r.CircuitBreaker.FailureThreshold <= 0 || r.CircuitBreaker.SuccessThreshold <= 0 || r.CircuitBreaker.OpenTimeout <= 0 {
			errs = append(errs, fmt.Errorf("%s: circuit breaker values must be positive", p))
		}
		if r.HealthCheck.Interval <= 0 || r.HealthCheck.Timeout <= 0 || r.HealthCheck.UnhealthyThreshold <= 0 || r.HealthCheck.HealthyThreshold <= 0 {
			errs = append(errs, fmt.Errorf("%s: health check values must be positive", p))
		}
		if r.RateLimit.Enabled && (r.RateLimit.Requests <= 0 || r.RateLimit.Burst <= 0 || r.RateLimit.Window <= 0 || r.RateLimit.MaxKeys <= 0) {
			errs = append(errs, fmt.Errorf("%s: rate limit values must be positive", p))
		}
		if r.RateLimit.Key != "ip" && r.RateLimit.Key != "header" {
			errs = append(errs, fmt.Errorf("%s: rate_limit key must be ip or header", p))
		}
		if r.RateLimit.Key == "header" && r.RateLimit.Header == "" {
			errs = append(errs, fmt.Errorf("%s: rate_limit header is required", p))
		}
		for _, u := range r.Upstreams {
			if u.ID == "" {
				errs = append(errs, fmt.Errorf("%s: upstream id is required", p))
			} else if ids[u.ID] {
				errs = append(errs, fmt.Errorf("duplicate upstream id %q", u.ID))
			}
			ids[u.ID] = true
			parsed, e := url.Parse(u.URL)
			if e != nil || parsed.Scheme != "http" && parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
				errs = append(errs, fmt.Errorf("%s: upstream %q has invalid URL", p, u.ID))
			}
		}
	}
	return errors.Join(errs...)
}
