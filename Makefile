GO ?= go

build:
	$(GO) build ./cmd/...
run:
	$(GO) run ./cmd/go-api-gateway serve --config configs/go-api-gateway.example.yaml
test:
	$(GO) test ./...
test-race:
	$(GO) test -race ./...
lint:
	golangci-lint run
benchmark:
	$(GO) test -bench=. -benchmem ./...
compose-up:
	docker compose -f deployments/docker-compose.yml up --build
compose-down:
	docker compose -f deployments/docker-compose.yml down
demo:
	bash scripts/demo.sh
