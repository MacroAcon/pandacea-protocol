Write-Host "Stopping observability stack..."
docker compose stop otel-collector prometheus grafana
Write-Host "Stopped."

