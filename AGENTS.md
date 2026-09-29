# Repository Guidelines

## Project Structure & Module Organization

go-api-gateway is a Go HTTP API gateway and reverse proxy. The implementation plan is in `go-api-gateway_IMPLEMENTATION_PLAN.md`. The intended layout keeps entry points in `cmd/go-api-gateway` and `cmd/mockservice`, implementation packages in `internal/`, administrative API definitions in `api/`, example YAML in `configs/`, and Docker artifacts in `deployments/`. Put integration tests in `tests/integration`; keep unit tests next to the package they exercise.

Keep HTTP routing, proxying, balancing, resilience, rate limiting, and observability in separate `internal/<area>` packages. Do not expose internal details as public Go APIs.

## Build, Test, and Development Commands

Once `go.mod` is present, run these commands from the repository root:

```bash
go run ./cmd/go-api-gateway serve --config configs/go-api-gateway.example.yaml
go test ./...
go test -race ./...
go test -bench=. ./...
go vet ./...
```

The first starts the gateway with an example configuration. The remaining commands run tests, detect races, benchmark, and check static issues. Use `golangci-lint run` and `govulncheck ./...` when available.

## Coding Style & Naming Conventions

Format Go code with `gofmt`; use tabs as Go tooling emits them. Follow standard Go naming: exported identifiers use `PascalCase`, unexported identifiers use `camelCase`, and package names are short lowercase words such as `ratelimit` or `observability`. Name files by responsibility (`round_robin.go`, `circuit_breaker.go`) and tests `*_test.go`.

Pass `context.Context` through request and background-work boundaries. Prefer atomics for independent counters and mutexes for coordinated state changes. Do not log request bodies, authorization headers, cookies, credentials, or internal upstream URLs in client-facing errors.

## Testing Guidelines

Use `testing` and `httptest`; use `testify` only when it improves readability. Cover routing precedence, rewriting, timeouts, retry safety, health transitions, rate limiting, and circuit-breaker states. Write deterministic concurrency tests and run affected packages with `-race`. Benchmark hot paths such as route matching and upstream selection.

## Commit & Pull Request Guidelines

No Git history exists yet, so use concise imperative commit subjects, for example `Add round-robin upstream selection`. Keep commits focused. Pull requests should explain the behavior change, test commands run, configuration or API changes, and operational effects. Link relevant issues and include sample output or screenshots for admin endpoints, metrics, or demo changes.

## Configuration & Security

Commit only sanitized example YAML. Validate configuration with `go-api-gateway validate --config <file>` before review. Keep administrative and profiling endpoints off public networks; pprof must remain disabled unless explicitly enabled.
