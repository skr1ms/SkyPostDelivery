# Parcel Automat Parser

## Overview (EN)
- Stack: Python 3 (FastAPI/HTTP clients, asyncio), Arduino firmware (`arduino_controller.ino`).
- Runs on the parcel automat controller, managing locker cells, QR scanning, and communication with the orchestrator.
- Responsibilities:
  - Validate QR codes and request cell opening via orchestrator APIs/gRPC.
  - Control locker electronics through Arduino (open/close, status sensors).
  - Report cell status updates back to the system.
  - Expose health endpoints for monitoring.

### Useful Commands
- `pip install -r requirements.txt` — install dependencies.
- `python main.py` — start the controller service.
- `pytest tests/` — run automated tests.
- Arduino firmware: open `arduino_controller.ino` in the Arduino IDE, flash to the MCU.

### Key Directories
- `app/` — FastAPI application, services, repositories.
- `config/` — settings and environment configuration.
- `tests/` — unit/integration tests.
- `run_as_service/` — systemd service scripts for deployment.

### Documentation
- EN: [`docs/en/ARCHITECTURE.md`](../../docs/en/ARCHITECTURE.md)
- RU: [`docs/ru/АРХИТЕКТУРА.md`](../../docs/ru/АРХИТЕКТУРА.md), [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md)

---

## Обзор (RU)
- Стек: Python 3 (FastAPI/HTTP клиенты, asyncio), прошивка Arduino (`arduino_controller.ino`).
- Запускается на контроллере постамата, управляет ячейками, сканером QR и взаимодействием с оркестратором.
- Обязанности:
  - Проверка QR-кодов и запрос открывания ячеек через API/gRPC оркестратора.
  - Управление электроникой (замки, датчики) через Arduino.
  - Отправка статусов ячеек в систему.
  - Предоставление health-эндпоинтов для мониторинга.

### Полезные команды
- `pip install -r requirements.txt` — установка зависимостей.
- `python main.py` — запуск сервиса контроллера.
- `pytest tests/` — автоматические тесты.
- Прошивка Arduino: открыть `arduino_controller.ino` в Arduino IDE и прошить МК.

### Основные каталоги
- `app/` — FastAPI-приложение, сервисы, репозитории.
- `config/` — конфигурация.
- `tests/` — тесты.
- `run_as_service/` — скрипты systemd для деплоя.

### Документация
- EN: [`docs/en/ARCHITECTURE.md`](../../docs/en/ARCHITECTURE.md)
- RU: [`docs/ru/АРХИТЕКТУРА.md`](../../docs/ru/АРХИТЕКТУРА.md), [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md)
