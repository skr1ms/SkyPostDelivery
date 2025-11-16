# hiTech Deployment Guide

## Scope
- Covers deploying the drone delivery platform (Go orchestrator, Python drone service, admin panel, storage, messaging, and observability stack).
- Applies to local development, staging, and production environments using Docker Compose.
- Assumes Linux hosts with Docker Engine ≥ 24 and Docker Compose plugin ≥ 2.20.

## Prerequisites
- Docker Engine and Docker Compose CLI installed and running.
- Git access to clone `hiTech`.
- Access credentials for:
  - PostgreSQL (initial superuser and target database).
  - MinIO root user/password.
  - RabbitMQ admin user/password.
  - SMS Aero API (email and key) for SMS delivery.
  - Firebase service account JSON file if push notifications are required.
- DNS entries (or hosts file mappings) for public-facing services when running in production with Nginx.

## Repository Layout (deployment-relevant)
- `env.example` — base environment template; copy to `.env` (prod) or `.env.dev`.
- `deployment/docker/docker-compose.dev.yml` — full local stack with exposed ports.
- `deployment/docker/docker-compose.prod.yml` — production stack expected behind an edge proxy.
- `deployment/docker/docker-compose.npm.yml` — alternative prod stack for environments using Nginx Proxy Manager (ports already exposed).
- `deployment/nginx/dev.conf` and `deployment/nginx/prod.conf` — reverse proxy configurations for admin panel, API, drone service, MinIO, Grafana, and RabbitMQ.
- `monitoring/` — Prometheus, Grafana, Loki, and Promtail configuration bundles.

## Environment Configuration
1. Copy `env.example` to the correct file:
   - Local development: `cp env.example .env.dev`
   - Production: `cp env.example .env`
2. Populate secrets and URLs:
   - `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `QR_HMAC_SECRET`
   - `POSTGRES_*` credentials and `DATABASE_URL` (note: Compose expects service hostnames like `postgres`).
   - `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`, `MINIO_PUBLIC_URL`
   - RabbitMQ credentials (`RABBITMQ_USER`, `RABBITMQ_PASSWORD`)
   - `SMSAERO_EMAIL`, `SMSAERO_API_KEY`
   - `ADMIN_PANEL_URL`, `VITE_API_URL`, `VITE_MINIO_URL`, `VITE_GRAFANA_URL`
   - `DRONE_SERVICE_HTTP_URL`, `GRPC_DRONE_SERVICE_URL`, `ORCHESTRATOR_GRPC_URL`
   - `FIREBASE_CREDENTIALS_FILE` — absolute path inside the container; Compose mounts the host file at `/app/secrets/firebase-service-account.json`. Update the host path in the Compose file if you store the JSON elsewhere.
3. Optional monitoring:
   - `PROMETHEUS_PORT`, `LOKI_PORT`, `GRAFANA_PORT`, and Grafana admin credentials.

### Secrets and External Services
- **Firebase**: Place the service account JSON on the host, adjust the bind mount in the Compose file (`go-orchestrator.volumes`) if the path differs from `/home/takuya/googleFirebase`.
- **SMS Aero**: Required for user registration/login flows; leave empty to disable SMS (application logs a warning).
- **Admin bootstrap**: Environment variables prefixed with `ADMIN_` and `SECOND_ADMIN_` seed admin accounts on startup; ensure strong passwords and valid phone numbers.

## Local Development Environment (`docker-compose.dev.yml`)
1. Ensure Docker is running and the `.env.dev` file is present.
2. Start the stack:
   ```bash
   docker compose --env-file .env.dev -f deployment/docker/docker-compose.dev.yml up --build -d
   ```
3. Verify core services:
   - Go orchestrator HTTP API: `http://localhost:${HTTP_PORT}` (default 8080)
   - Go orchestrator metrics: `http://localhost:9091/metrics`
   - Drone service HTTP/API: `http://localhost:${DRONE_SERVICE_HTTP_PORT}` (default 8081)
   - Drone service metrics: `http://localhost:9092/metrics`
   - Admin panel: `http://localhost`
   - PostgreSQL: `localhost:5433`
   - MinIO console: `http://localhost:9001`
   - RabbitMQ UI: `http://localhost:15672`
   - Grafana: `http://localhost:3000` (default creds `admin`/`admin`, configurable via env)
   - Prometheus: `http://localhost:9090`
   - Loki API: `http://localhost:3100`

### Dev Stack Composition
| Service | Image / Build | Purpose | Notes |
| --- | --- | --- | --- |
| `postgres` | `postgres:17-alpine` | Primary relational database | Data stored in `postgres_data_dev` volume |
| `minio` | `minio/minio:latest` | S3-compatible object store for QR codes & delivery records | Console on port 9001, API on 9000 |
| `rabbitmq` | `rabbitmq:4.2.0-rc.1-management` | Messaging backbone for drone workflow | Prometheus & management plugins enabled |
| `go-orchestrator` | Built from `backend/go-orchestrator` | REST + gRPC orchestrator | Exposes 8080 (HTTP) & `${GO_GRPC_PORT}` |
| `drone-service` | Built from `backend/drone-service` | Drone control, WebSocket, gRPC worker | Depends on RabbitMQ, PostgreSQL |
| `admin-panel` | Built from `admin_panel` | Operator UI (Vite/React) | Served via Nginx (port 80) |
| `prometheus`, `loki`, `promtail`, `grafana`, `cadvisor`, `postgres-exporter`, `node-exporter` | Prebuilt observability stack | Metrics, logs, dashboards | Promtail tails Docker logs |

### Helpful Dev Commands
- Tail logs: `docker compose -f deployment/docker/docker-compose.dev.yml logs -f go-orchestrator`
- Rebuild orchestrator after code changes: `docker compose -f deployment/docker/docker-compose.dev.yml up --build go-orchestrator -d`
- Stop stack: `docker compose -f deployment/docker/docker-compose.dev.yml down`
- Remove volumes (DANGER: deletes data): `docker compose -f deployment/docker/docker-compose.dev.yml down -v`

## Production Deployment (`docker-compose.prod.yml`)
1. Prepare the host:
   - Harden the OS (firewall, automatic updates).
   - Install Docker and Docker Compose.
   - Configure persistent storage (local volumes or bind mounts with backups).
2. Create `.env` from `env.example` and fill secrets with production values.
3. Adjust `deployment/nginx/prod.conf`:
   - Update `server_name` entries to match your domains.
   - Review upstream names if you rename services.
   - Consider enabling TLS by mounting certificates and switching `listen 80` to `listen 443 ssl`.
4. Launch:
   ```bash
   docker compose --env-file .env -f deployment/docker/docker-compose.prod.yml up --build -d
   ```
5. Validate readiness:
   - `docker compose -f deployment/docker/docker-compose.prod.yml ps`
   - Access the admin panel domain and confirm the UI loads through Nginx.
   - Check orchestrator and drone-service logs for `HTTP server started` messages.
   - Check Prometheus (`curl http://<prometheus-host>:9090/-/ready`) and Grafana login.
6. Configure monitoring endpoints to trust your domain and enforce authentication as needed.

### Reverse Proxy Considerations
- Nginx container handles routing for:
  - `/api/` → Go orchestrator
  - `/` → Admin panel SPA
  - `py.*` hostnames for drone service (including WebSocket upgrade)
  - Dedicated hosts for MinIO, Grafana, RabbitMQ
- If deploying behind an external load balancer or proxy:
  - Disable or remove the bundled Nginx service.
  - Expose orchestrator, admin panel, drone service, and MinIO ports as required.
  - Ensure `VITE_API_URL`, `VITE_MINIO_URL`, `VITE_GRAFANA_URL` match external URLs.

### Alternative Production Stack (`docker-compose.npm.yml`)
- Designed for installations that already use Nginx Proxy Manager.
- Exposes service ports directly (no internal Nginx).
- Requires external proxy to terminate TLS and route traffic.
- Start with:
  ```bash
  docker compose --env-file .env -f deployment/docker/docker-compose.npm.yml up --build -d
  ```

## Observability & Operations
- **Prometheus**: Scrapes orchestrator (`/metrics`), drone service, exporters, and writes to `prometheus_data` volume.
- **Grafana**: Pre-provisioned dashboards located in `monitoring/grafana/dashboards`; adjust datasources via provisioning files.
- **Loki + Promtail**: Collect container logs; add labels via Compose service labels (`logging=promtail`).
- **cAdvisor & Node Exporter**: Provide host resource metrics.
- **Alerting**: `monitoring/prometheus/alerts.yml` contains starter alert rules; integrate with Alertmanager if required.
- **Health Checks**:
  - Orchestrator migrations run on startup; failure stops the container.
  - Drone service has `/health` endpoint.
  - Postgres and MinIO include Compose health checks; RabbitMQ uses `rabbitmq-diagnostics`.

## Data Management
- **Database migrations**: Automatically executed by the orchestrator on startup via `pkg/migrator`. If a migration fails, fix SQL files and restart the container.
- **Admin seeding**: On boot, orchestrator creates/updates admin users defined in env and generates QR codes; clear env vars to skip.
- **Backups**:
  - Postgres: `docker compose exec postgres pg_dump -U $POSTGRES_USER $POSTGRES_DB > backup.sql`
  - MinIO: Use `mc mirror` or schedule object storage replication.
  - Grafana dashboards: stored in `grafana_data` volume; export if using external persistence.
- **Restores**:
  - Postgres: `psql -U $POSTGRES_USER -d $POSTGRES_DB < backup.sql`
  - MinIO: `mc mirror` back into bucket.

## Upgrades & Rollbacks
1. Pull latest code: `git pull origin main`.
2. Review release notes/changelogs for migration changes.
3. Rebuild affected images:
   ```bash
   docker compose --env-file .env -f deployment/docker/docker-compose.prod.yml up --build go-orchestrator drone-service admin-panel -d
   ```
4. Monitor logs (`docker compose ... logs -f`) for migrations or dependency errors.
5. Rollback by checking out previous commit and re-running the compose build, or by restoring backups.

## Troubleshooting
- **Container crash loops**: Inspect logs (`docker compose ... logs <service>`). For orchestrator failures, look for migration errors or missing env vars.
- **JWT/auth issues**: Ensure secrets and clock synchronization (NTP) on hosts.
- **MinIO SSL redirects**: `MINIO_CONSOLE_SECURE_TLS_REDIRECT` is disabled by default; set to `"on"` when serving behind HTTPS.
- **RabbitMQ refuses connections**: Confirm credentials in env and ensure firewall allows `5672` and `15672`.
- **Firebase push disabled**: If `FIREBASE_CREDENTIALS_FILE` is empty, push sender falls back to no-op; populate and mount the JSON to enable notifications.
- **Admin panel CORS errors**: Update `ADMIN_PANEL_URL` to the externally reachable URL so orchestrator allows it.

## Appendix: Default Ports & URLs
| Component | Port | Source |
| --- | --- | --- |
| Go orchestrator API | 8080 | `HTTP_PORT` |
| Go orchestrator gRPC | 50052 (default) | `GO_GRPC_PORT` |
| Go orchestrator metrics | 9091 | Compose static mapping |
| Drone service HTTP/WebSocket | 8081 | `DRONE_SERVICE_HTTP_PORT` |
| Drone service gRPC | 50051 | `DRONE_SERVICE_GRPC_PORT` |
| Drone service metrics | 9092 | Compose static mapping |
| PostgreSQL | 5432 (5433 on dev host) | `POSTGRES_PORT`/Compose |
| MinIO API / Console | 9000 / 9001 | Env |
| RabbitMQ AMQP / UI | 5672 / 15672 | Env |
| Grafana | 3000 | `GRAFANA_PORT` |
| Prometheus | 9090 | `PROMETHEUS_PORT` |
| Loki | 3100 | `LOKI_PORT` |
| Admin panel | 80 | Exposed by Nginx container |

Keep this document under version control and update it alongside infrastructure changes to maintain alignment with the deployed stack.

