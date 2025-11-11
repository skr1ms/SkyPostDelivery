# Technology Stack Overview

This document lists the core technologies used across the hiTech platform, explaining why each was chosen and how it is employed.

## Backend Services

### Go Orchestrator
- **Language:** Go 1.21+
  - Chosen for concurrency, performance, and static typing in a mission-critical orchestrator.
- **Frameworks & Libraries:**
  - **Gin** (`github.com/gin-gonic/gin`): Lightweight HTTP framework for REST APIs. Provides middleware ecosystem, low latency.
  - **gRPC** (`google.golang.org/grpc`): High-performance RPC between orchestrator and drone-service for locker operations and status updates.
  - **Prometheus Client** (`github.com/prometheus/client_golang/prometheus`): Exposes metrics for observability (request counters, latency histograms).
  - **SQLC + pgx**: Type-safe data access to PostgreSQL. SQLC generates Go repositories from SQL queries; `pgx` handles connection pooling.
  - **RabbitMQ Client** (`github.com/rabbitmq/amqp091-go`): Publishes delivery tasks and confirmations to message queues.
  - **JWT** (`github.com/golang-jwt/jwt`): Implements authentication tokens for users/admins.
  - **RabbitMQ Workers**: Background delivery workers manage asynchronous workflows.

### Drone Service (Python)
- **Language:** Python 3.11
  - Python chosen for rapid development, async capabilities, and library ecosystem.
- **Frameworks & Libraries:**
  - **FastAPI**: Serves REST endpoints (health, metrics, admin views). Async-friendly.
  - **Pydantic**: Validates payloads, settings (`pydantic-settings` for env-based configuration).
  - **gRPC** (`grpcio`): Communicates with Go orchestrator (cell operations, state sync).
  - **WebSockets** (`websockets` / FastAPI WebSocket support): Real-time control channel to physical drones and admin dashboard.
  - **RabbitMQ** (`aio-pika` or custom wrapper): Consumes delivery/return tasks, publishes confirmations.
  - **Prometheus Client** (`prometheus_client`): Exposes FastAPI request metrics on `/metrics`.
  - **Asyncio**: Runs background workers, heartbeat loops, command listeners.

### Parsers (Hardware Integration)
- **Clover ROS Scripts (Python):**
  - **ROS** (Robot Operating System): Provides telemetry, navigation services, ArUco detection.
  - **Clover API**: Simplified commands for takeoff, navigation, landing.
  - **Custom Navigation Controller:** Implements mission logic, ArUco-based positioning, drop confirmation.
- **Parcel Automat Controller:**
  - Python service that interfaces with electronic locks, QR scanners, orchestrator API.
  - **Arduino firmware** (`arduino_controller.ino`): Controls actuators/sensors inside lockers.

## Frontend

### Admin Panel (Web)
- **React 18 + TypeScript:** Modern SPA stack for operators.
- **Vite:** Provides fast dev server and build pipeline.
- **Font Awesome:** Iconography.
- **Context API / Hooks:** Authentication and state management.
- **REST Fetch Clients:** Typed API wrappers hitting orchestrator endpoints.
- **Docker:** Containerised build and deploy.

### Mobile App (Flutter)
- **Flutter + Dart:** Cross-platform mobile app for Android/iOS/Web.
- **Clean Architecture Layers:** `data`, `domain`, `presentation` per feature.
- **HTTP Client:** Custom wrapper with interceptors for auth.
- **Provider / Riverpod (depending on feature):** State management.
- **Firebase (optional):** Push notifications (FCM) via service account.

## Data & Messaging
- **PostgreSQL:** Primary relational database for orchestrator (orders, deliveries, users, goods).
- **MinIO:** S3-compatible object storage for QR codes/media.
- **RabbitMQ:** Message broker orchestrating deliveries, priority queues, return commands.
- **Redis (optional, not in repo):** Could be added for caching/session storage if needed.

## Observability & Monitoring
- **Prometheus:** Metric collection (scrapes orchestrator, drone-service, exporters).
- **Grafana:** Dashboards for system metrics, drone status, RabbitMQ usage.
- **Loki + Promtail:** Centralised log aggregation via Docker autodiscovery.
- **Alertmanager (optional):** Receives Prometheus alerts (default config shipped but disabled).
- **cAdvisor & Node Exporter:** Container and host-level metrics.
- **RabbitMQ Exporter / Postgres Exporter:** Additional metrics for queues and DB.

## Infrastructure & Deployment
- **Docker Compose:** Orchestrates multi-service stacks for dev/prod.
- **Nginx:** Reverse proxy for TLS termination and routing.
- **Makefiles:** Build/run shortcuts (Go orchestrator).
- **Python Virtualenv/requirements.txt:** Reproducible environment for Python services.
- **Flutter tooling:** `flutter`, `dart` CLI for mobile builds.

## Security & Auth
- **JWT-based Auth:** Access/refresh tokens stored server-side for user sessions.
- **SMSAero API:** Sends verification codes (2FA) during registration.
- **MinIO with Access Keys:** Secures object storage endpoints.
- **TLS (via Nginx):** Recommended in production for API, admin panel, Grafana, etc.

## Testing & Quality
- **Go Tests:** `go test ./...` for orchestrator.
- **Pytest:** Drone-service and parsers include test suites (`tests/` directories).
- **Flutter Tests:** Widget/unit tests under `mobileapp/test/`.
- **Static Analysis:**
  - Go: `go vet`, `staticcheck` (optional).
  - Python: `mypy`, `ruff` (can be integrated).
  - Flutter: `dart analyze`.

## Summary
The stack combines high-performance Go microservices, Python-based hardware orchestration, and modern frontend/mobile clients, all supported by a Docker-based deployment and a comprehensive observability toolchain.
