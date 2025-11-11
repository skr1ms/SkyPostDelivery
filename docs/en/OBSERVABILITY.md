# hiTech Observability Guide

## Overview
- **Logging:** Application containers emit structured logs to stdout. Promtail harvests the logs and pushes them to Loki for centralized querying.
- **Metrics:** Each service exposes Prometheus-compatible metrics. Prometheus scrapes the targets and Grafana visualises dashboards.
- **Alerting:** Prometheus ships with placeholder alert rules (`monitoring/prometheus/alerts.yml`). Extend them and wire to Alertmanager when production alerting is required.

## Logging Pipeline

### Flow
1. Services (Go orchestrator, Python drone-service, admin panel, Postgres, MinIO) write logs to stdout/stderr.
2. `promtail` (see `monitoring/promtail/promtail-config.yml`) uses Docker socket discovery to tail containers whose `com.docker.compose.service` label matches the configured jobs.
3. Promtail enriches log entries with labels:
   - `container_name`
   - `service` (Compose service name)
   - stack-wide metadata (`environment`, `compose_project`) provided automatically by Docker Compose.
4. Logs are pushed to Loki at `http://loki:3100` (internal network).
5. Grafana queries Loki via Explore or dashboards.

### Service Coverage
| Service | Promtail job | Notes |
| --- | --- | --- |
| Go orchestrator | `go-orchestrator` | Uses Go `log` package. Adjust verbosity with `LOG_LEVEL` (see `.env`). |
| Drone service | `drone-service` | Python `logging` module initialises INFO level by default. |
| Admin panel | `admin-panel` | Vite/Node logs available for troubleshooting frontend builds. |
| Postgres | `postgres` | Captures backend SQL logs. Consider toggling `POSTGRES_LOG_STATEMENT` in environment if volume is high. |
| MinIO | `minio` | Useful for storage audit.

To include additional containers, copy an existing job in `promtail-config.yml` and replace `com.docker.compose.service` with the target service label.

### Loki Storage & Retention
- Storage: local filesystem at `/loki/{chunks,rules}` (see `monitoring/loki/loki-config.yml`).
- Retention: `retention_enabled` is `false` by default. Enable and tune `compactor.retention_enabled` and `limits_config` before going to production.
- Access: Grafana (port `GRAFANA_PORT`, default `3000`) → *Explore* tab → data source `Loki`.

### Query Cheatsheet
| Goal | Loki LogQL example |
| --- | --- |
| Errors from orchestrator | `{service="go-orchestrator"} |= "ERROR"` |
| Drone WebSocket events | `{service="drone-service"} |= "websocket"` |
| Admin panel build output | `{service="admin-panel"}` |

## Metrics Pipeline

### Targets & Ports
Prometheus configuration lives in `monitoring/prometheus/prometheus.yml`. Default scrape interval is 15 seconds.

| Job | Target | Exposed Port | Purpose |
| --- | --- | --- | --- |
| prometheus | `localhost:9090` | 9090 | Self-monitoring |
| go-orchestrator | `go-orchestrator:9091` | 9091 | HTTP & gRPC middleware counters/histograms |
| drone-service | `drone-service:9092` | 9092 | FastAPI request metrics, WebSocket counters |
| rabbitmq | `rabbitmq:15692` | 15692 | Requires `RABBITMQ_PROMETHEUS_PLUGIN=1` (enabled in Compose) |
| cadvisor | `cadvisor:8080` | 8080 | Container CPU/memory usage |
| node-exporter | `node-exporter:9100` | 9100 | Host-level metrics |
| postgres | `postgres-exporter:9187` | 9187 | Requires `POSTGRES_EXPORTER_*` env values |

### Application Instrumentation
- **Go orchestrator:** `internal/controller/http/middleware/prometheus.go` and `internal/controller/grpc/middleware/prometheus.go` expose request counters (`hitech_http_requests_total`) and latency histograms (`hitech_http_request_duration_seconds`). Mount `prometheus/promhttp` handler at `/metrics` (port 9091 defined in Compose).
- **Drone service:** `app/api/prometheus.py` registers `fastapi_requests_total` and `fastapi_request_duration_seconds`. Metrics handler served on `/metrics` (port 9092).
- **Custom metrics:** Import `prometheus` client libraries (`github.com/prometheus/client_golang/prometheus` or `prometheus_client` in Python) and register counters/histograms before application start. They will be picked up automatically by Prometheus.

### Grafana Dashboards
- Location: `monitoring/grafana/dashboards/` (JSON exported dashboards). Import them via Grafana UI (`Dashboards → Import`).
- Credentials: `GRAFANA_USER` / `GRAFANA_PASSWORD` from environment (see `env.example` lines 60-61). Change defaults immediately for non-local deployments.
- Recommended data sources:
  - `Prometheus` → URL `http://prometheus:9090`.
  - `Loki` → URL `http://loki:3100`.

### Alerts & Notifications
- Alert rules: `monitoring/prometheus/alerts.yml` contains starter rules. Extend them with service-specific thresholds (e.g., high error rate, queue backlog).
- Hook to Alertmanager by adding an `alertmanager` target in `prometheus.yml` and pointing Grafana notifications to your channel (email, Slack, etc.).

## Operational Tips
- **Log Volume Control:** adjust application log levels through env vars (`LOG_LEVEL`, `DRONE_LOG_LEVEL`). For Go, wrap `log.Println` with conditional logging if needed.
- **Disk Capacity:** Loki stores chunks locally; monitor `/loki` volume usage via cAdvisor or host metrics.
- **Backups:** Metrics are time-series and usually not backed up. If compliance requires, snapshot Prometheus TSDB directory `/prometheus`.
- **Security:** Both Prometheus and Loki are internal-only in default Compose files. When exposing externally, protect with reverse proxy + auth.
- **Local Access URLs (default dev compose):**
  - Prometheus: `http://localhost:${PROMETHEUS_PORT:-9090}`
  - Grafana: `http://localhost:${GRAFANA_PORT:-3000}`
  - Loki (API): `http://localhost:${LOKI_PORT:-3100}`

## Extending the Stack
1. **Add a new metrics target:**
   - Expose `/metrics` on the service container.
   - Append a `job_name` stanza to `monitoring/prometheus/prometheus.yml`.
   - Reload Prometheus (`docker compose restart prometheus`).
2. **Add a new log source:**
   - Label the container with `com.docker.compose.service=<name>`.
   - Add a matching job to `promtail-config.yml`.
   - Restart Promtail (`docker compose restart promtail`).
3. **Dashboards as Code:** Commit exported JSON dashboards under `monitoring/grafana/dashboards` to keep them versioned.

---
For quick start, run `docker compose -f deployment/docker/docker-compose.dev.yml up promtail loki prometheus grafana` and log into Grafana to verify both data sources are healthy.
