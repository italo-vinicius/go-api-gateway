#!/usr/bin/env bash
set -euo pipefail

base=http://localhost:8080
admin=http://localhost:9090
echo 'Alternating orders instances:'
for _ in 1 2 3 4; do curl -fsS "$base/orders/42"; done
echo 'Gateway status:'
curl -fsS "$admin/status"
echo 'Rate-limit sample (expect 429 after burst):'
for _ in $(seq 1 25); do curl -s -o /dev/null -w '%{http_code}\n' "$base/orders/42"; done
echo 'Metrics:'
curl -fsS "$admin/metrics" | grep '^go_api_gateway_' || true
