# go-api-gateway

A Go HTTP gateway and reverse proxy focused on safe traffic control, resilience, and operational visibility.

## Quick start

```bash
docker compose -f deployments/docker-compose.yml up --build
curl http://localhost:8080/orders/42
curl http://localhost:9090/status
```

The public gateway and the administrative API intentionally listen on different ports. Local execution uses `go run ./cmd/go-api-gateway serve --config configs/go-api-gateway.example.yaml`.

## What it does

- YAML-defined prefix routes, with longest-prefix precedence and optional stripping.
- Concurrent round-robin among healthy upstreams.
- Per-route total timeouts, safe-method retries with capped exponential jitter, and per-upstream circuit breakers.
- Active health checks, local token-bucket rate limiting by client IP or header, structured JSON errors, request IDs, and Prometheus metrics.
- `/health/live`, `/health/ready`, `/status`, `/metrics`, and optional admin-only pprof endpoints.

## Architecture

`request ID → route match → rate limit → healthy/circuit-eligible upstream → attempt/retry → metrics and log`

The implementation uses the standard `net/http` stack to make HTTP cancellation, transports, and forwarding explicit. Health checks detect background availability; circuit breakers react to live request failures. Retries are intentionally restricted to `GET`, `HEAD`, and `OPTIONS` to avoid duplicating side effects. The timeout covers every retry and backoff.

## Operations and trade-offs

Rate-limit state is in-process, so a distributed limit needs shared storage such as Redis. There is no database or service discovery by design. TLS is expected to terminate at an edge proxy. This is a learning-quality gateway, not a replacement for production platforms such as Envoy, Kong, or Traefik.

Run `go test ./...`, `go test -race ./...`, and `go test -bench=. -benchmem ./...`. The mock services expose `/health`, `/info`, and a demo-only `POST /admin/failure-mode` endpoint. `bash scripts/demo.sh` shows routing, limits, status, and metrics after Compose is running.
