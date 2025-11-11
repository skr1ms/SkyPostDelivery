# Drone Service

## Overview (EN)
- Language: Python 3.11+ (FastAPI, asyncio, gRPC, Pydantic).
- Role: mediates between orchestrator and physical drones; manages WebSocket sessions, delivery tasks, telemetry, and status updates.
- Integrations: RabbitMQ (`deliveries`, `delivery.return`), PostgreSQL state store, MinIO, orchestrator gRPC.
- Metrics: Prometheus endpoint on `:9092/metrics`.

### Useful Commands
- `poetry install` or `pip install -r requirements.txt` — set up environment.
- `uvicorn main:app --reload` — run FastAPI + background workers locally.
- `pytest tests/` — run test suite.

### Key Directories
- `app/api/` — HTTP routes, schemas, metrics handler.
- `app/presentation/` — WebSocket handlers (drone/admin).
- `app/core/` — use cases, entities, drone manager.
- `app/infrastructure/` — adapters (RabbitMQ, Postgres state repository).
- `app/services/workers/` — background consumers for delivery and return queues.

### Documentation
- EN: [`docs/en/ARCHITECTURE.md`](../../docs/en/ARCHITECTURE.md), [`docs/en/WORKFLOW.md`](../../docs/en/WORKFLOW.md), [`docs/en/OBSERVABILITY.md`](../../docs/en/OBSERVABILITY.md)
- RU: [`docs/ru/АРХИТЕКТУРА.md`](../../docs/ru/АРХИТЕКТУРА.md), [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md), [`docs/ru/МОНИТОРИНГ И ЛОГИРОВАНИЕ.md`](../../docs/ru/МОНИТОРИНГ%20И%20ЛОГИРОВАНИЕ.md)

---

## Обзор (RU)
- Язык: Python 3.11+ (FastAPI, asyncio, gRPC, Pydantic).
- Роль: связующее звено между оркестратором и физическими дронами; управляет WebSocket-подключениями, задачами доставки, телеметрией и статусами.
- Интеграции: RabbitMQ (`deliveries`, `delivery.return`), PostgreSQL, MinIO, gRPC с оркестратором.
- Метрики: Prometheus на `:9092/metrics`.

### Полезные команды
- `pip install -r requirements.txt` или `poetry install` — установка зависимостей.
- `uvicorn main:app --reload` — запуск FastAPI и фоновых воркеров локально.
- `pytest tests/` — тесты.

### Основные директории
- `app/api/` — HTTP-роуты, схемы, метрики.
- `app/presentation/` — WebSocket-обработчики (дрон/админ-панель).
- `app/core/` — бизнес-логика, менеджер дронов.
- `app/infrastructure/` — адаптеры RabbitMQ/Postgres.
- `app/services/workers/` — фоновые потребители очередей.

### Документация
- EN: [`docs/en/ARCHITECTURE.md`](../../docs/en/ARCHITECTURE.md), [`docs/en/WORKFLOW.md`](../../docs/en/WORKFLOW.md), [`docs/en/OBSERVABILITY.md`](../../docs/en/OBSERVABILITY.md)
- RU: [`docs/ru/АРХИТЕКТУРА.md`](../../docs/ru/АРХИТЕКТУРА.md), [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md), [`docs/ru/МОНИТОРИНГ И ЛОГИРОВАНИЕ.md`](../../docs/ru/МОНИТОРИНГ%20И%20ЛОГИРОВАНИЕ.md)
