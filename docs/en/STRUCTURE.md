# hiTech Monorepo Structure

> The repository is organised as a monorepo containing backend services, drone-side parsers, the admin web panel, mobile app, deployment assets, and documentation. This document explains the purpose of each top-level directory and highlights important subcomponents.

## Top-Level Overview

| Path | Description |
| --- | --- |
| `admin_panel/` | React + Vite admin UI used by operators to monitor deliveries, manage goods, automats, drones, and users. |
| `backend/` | Server-side services: Go orchestrator, Python drone-service, and hardware parsers for drones/parcel automats. |
| `deployment/` | Docker Compose bundles and Nginx configs for dev/prod environments. |
| `docs/` | Documentation (EN/RU) plus architecture diagrams. |
| `mobileapp/` | Flutter mobile client for couriers/customers. |
| `monitoring/` | Prometheus, Grafana, Loki, Promtail configuration. |
| `proto/` | Shared gRPC definitions. |
| `env.example` | Template for `.env`/`.env.dev`. |

## Backend Services (`backend/`)

### `go-orchestrator/`
- **Language:** Go (Gin, gRPC).
- **Purpose:** Central API handling authentication, orders, goods management, deliveries, lockers, and integration with drone-service and parcel automat controllers.
- **Key paths:**
  - `cmd/app/main.go` – entry point for HTTP & gRPC servers.
  - `internal/controller/http` – Gin HTTP handlers organised by version (`v1`) and resource.
  - `internal/controller/grpc` – gRPC server for inter-service communication (cell opening, status updates).
  - `internal/usecase` – business logic orchestrating repos, queue publishing, etc.
  - `internal/usecase/repo` – repositories implemented using `sqlc` against PostgreSQL.
  - `pkg/pb` – generated gRPC clients.
  - `migrations/` – DB schema migrations.
  - `config/` – configuration loading (env, YAML).

### `drone-service/`
- **Language:** Python (FastAPI + gRPC clients + WebSockets).
- **Purpose:** Middleware between orchestrator and physical drones. Manages delivery tasks, maintains drone state, handles WebSocket connections, and exposes metrics.
- **Key paths:**
  - `app/api` – FastAPI routers, schemas, metrics endpoint.
  - `app/presentation` – WebSocket handlers for drones and admin panel.
  - `app/core` – use cases, entities, drone manager, state repository.
  - `app/infrastructure` – adapters (RabbitMQ client, Postgres state storage).
  - `app/services/workers` – background consumers (delivery, return tasks).
  - `services/delivery_worker.py` (deprecated; use `app/services/workers/delivery.py`).
  - `config/` – settings loader (`pydantic-settings`).

### `parsers/`
Hardware-facing scripts packaged per subsystem.
- `drone/` – ROS-integrated control logic, navigation, WebSocket client, delivery workflow.
  - `app/navigation` – Clover API wrapper, ArUco navigation.
  - `app/services` – WebSocket service, delivery orchestration.
  - `scripts/` – standalone ROS scripts (`takeoff.py`, `land.py`, etc.) for `rosrun` usage.
  - `test_scripts/` – diagnostic flight scripts, e.g., `simple_flight_test.py`.
  - `config/` – ArUco map, drone settings.
  - `HOW_TO_RUN.md` (deleted) replaced by `docs/` guides.
- `parcel_automat/` – Python controller for smart lockers interacting with orchestrator (QR scanning, cell control). Includes Arduino firmware (`arduino_controller.ino`).

## Admin Panel (`admin_panel/`)
- **Stack:** React 18, TypeScript, Vite.
- **Structure:**
  - `src/pages/` – high-level screens (Dashboard, Drones, Goods, Monitoring, etc.).
  - `src/components/` – reusable UI components (Button, ConfirmModal, Layout).
  - `src/api/` – API client wrappers for orchestrator endpoints.
  - `src/context/AuthContext.tsx` – authentication state/shared hooks.
  - `src/config/api_config.ts` – runtime URLs (read from env `VITE_*`).
  - `dist/` – production bundle.

## Mobile App (`mobileapp/`)
- **Framework:** Flutter/Dart.
- **Structure:**
  - `lib/core/` – DI container, networking (`HttpClient`), theme, push notification service.
  - `lib/features/` – modularised features (`auth`, `goods`, `orders`, `qr`, `notifications`, etc.).
    - Each feature contains `data`, `domain`, `presentation` layers following clean architecture.
  - `assets/`, `android/`, `ios/`, `web/` – platform resources.
  - `test/` – widget tests.

## Documentation (`docs/`)
- `en/` & `ru/` – language-specific guides (`API-SPEC.md`, `DEPLOYMENT.md`, `OBSERVABILITY.md`, translations kept in sync).
- `arch.png` – system architecture diagram.
- New documents should follow the EN/RU parity requirement.

## Deployment (`deployment/`)
- `docker/` – Compose files:
  - `docker-compose.dev.yml` – local full stack (hot reload, all monitoring components).
  - `docker-compose.prod.yml` – production baseline (behind Nginx).
  - `docker-compose.npm.yml` – variant designed for Nginx Proxy Manager environments.
- `nginx/` – reverse proxy configs for dev/prod (routing admin, API, drone-service, minio, grafana).

## Monitoring (`monitoring/`)
- `prometheus/` – Prometheus config (`prometheus.yml`) + alert rules (`alerts.yml`).
- `grafana/` – prebuilt dashboards and provisioning.
- `loki/`, `promtail/` – log aggregation configs.
- See `docs/en/OBSERVABILITY.md` for detailed setup instructions.

## Shared Assets
- `proto/orchestrator.proto` – gRPC definitions for orchestrator<->drone-service communication. Regenerated into Go (`pkg/pb`) and Python modules.
- `env.example` – environment template. Copy to `.env` (prod) or `.env.dev` (dev) and fill secrets (JWT keys, Postgres credentials, RabbitMQ, Grafana, etc.).

## Development Tips
- **Go services:** Run `make run` inside `backend/go-orchestrator`. Use `sqlc` for repo layers.
- **Python services:** Use Poetry or venv (requirements pinned). Uvicorn entry points defined in `main.py`.
- **ROS scripts:** Ensure ROS workspace sourced; use `rosrun clover <script>.py`.
- **Testing:**
  - Go: `go test ./...`.
  - Python: `pytest` in respective service folders.
  - Flutter: `flutter test`.
- **Docker:** `docker compose -f deployment/docker/docker-compose.dev.yml up` brings up full stack.

---
For deeper subsystem guides, consult `docs/en/DEPLOYMENT.md`, `docs/en/OBSERVABILITY.md`, and feature-specific READMEs inside each component.
