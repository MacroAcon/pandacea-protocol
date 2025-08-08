## Observability (Windows-first)

This doc explains how to enable end-to-end tracing, metrics, and dashboards.

### Prerequisites
- Docker Desktop for Windows
- PowerShell

### Environment
Create or update `.env` with:

```
LOG_LEVEL=INFO
PANDACEA_OTEL=1
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

Optional:
- `DEPLOYMENT_ENV=development` (used as resource attribute)

### Start the stack

PowerShell:

```
make obs-up
```

This starts:
- OpenTelemetry Collector (OTLP at 4317/4318)
- Prometheus (mapped to localhost:9091)
- Grafana (localhost:3000)

Then run the agent and SDK/tests as usual. With `PANDACEA_OTEL=1`, traces and metrics are exported.

### Stop the stack

```
make obs-down
```

### Health endpoints
- Liveness: `/healthz`
- Readiness: `/readyz`
- Metrics: `/metrics`

### Notes
- Logs are JSON and include `trace_id` and `span_id` when tracing is enabled.
- Sensitive values such as private keys are not logged; ensure `LOG_LEVEL` is appropriate in production.

