# go-api-gateway

Um gateway HTTP e proxy reverso em Go, focado em controle de tráfego seguro, resiliência e visibilidade operacional.

## Início rápido

Pré-requisitos: Docker com Docker Compose para a demonstração em containers, ou Go 1.25+ instalado e disponível no `PATH` para execução local. Confirme com `go version`.

```bash
docker compose -f deployments/docker-compose.yml up --build
curl http://localhost:8080/orders/42
curl http://localhost:9090/status
```

O gateway público e a API administrativa usam portas separadas intencionalmente. Para executar localmente, use `go run ./cmd/go-api-gateway serve --config configs/go-api-gateway.example.yaml` após instalar o Go. Com Docker, aguarde os containers iniciarem antes de chamar os endpoints com `curl`.

## Funcionalidades

- Rotas por prefixo definidas em YAML, com precedência pelo prefixo mais longo e remoção opcional do prefixo.
- Round-robin concorrente entre upstreams saudáveis.
- Timeouts totais por rota, retries para métodos seguros com backoff exponencial limitado e jitter, além de circuit breakers por upstream.
- Health checks ativos, rate limiting local com token bucket por IP ou header, erros JSON estruturados, request IDs e métricas Prometheus.
- Endpoints `/health/live`, `/health/ready`, `/status`, `/metrics` e pprof opcional, restrito à porta administrativa.

## Arquitetura

`request ID → correspondência de rota → rate limit → upstream saudável e elegível pelo circuito → tentativa/retry → métricas e logs`

A implementação usa a biblioteca padrão `net/http` para tornar explícitos o cancelamento HTTP, os transports e os headers de encaminhamento. Health checks observam disponibilidade em segundo plano; circuit breakers reagem a falhas de tráfego real. Retries são deliberadamente limitados a `GET`, `HEAD` e `OPTIONS`, evitando a repetição de efeitos colaterais. O timeout abrange todas as tentativas e backoffs.

## Operação e trade-offs

O estado do rate limit fica no processo; para um limite distribuído, seria necessário armazenamento compartilhado, como Redis. Por decisão de escopo, não há banco de dados nem descoberta de serviços. TLS deve ser terminado por um proxy de borda. Este projeto prioriza estudo e portfólio, não substitui soluções de produção como Envoy, Kong ou Traefik.

Execute `go test ./...`, `go test -race ./...` e `go test -bench=. -benchmem ./...`. Os serviços simulados expõem `/health`, `/info` e o endpoint de demonstração `POST /admin/failure-mode`. Depois de iniciar o Compose, `bash scripts/demo.sh` demonstra roteamento, limites, status e métricas.
