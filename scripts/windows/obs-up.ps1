Write-Host "Starting observability stack (otel-collector, prometheus, grafana)..."
docker compose up -d otel-collector prometheus grafana
Write-Host "Grafana: http://localhost:3000 | Prometheus: http://localhost:9091 | OTLP: http://localhost:4318"

