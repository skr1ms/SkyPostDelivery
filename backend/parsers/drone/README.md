# Drone Parser

## Overview (EN)
- Stack: Python 3, ROS (Clover), asyncio, WebSockets.
- Runs on the drone (Raspberry Pi) and executes physical delivery missions.
- Responsibilities:
  - Connect to drone-service via WebSocket (`/ws/drone`).
  - Execute tasks (takeoff, navigate to ArUco marker, landing, payload drop, return to base).
  - Report telemetry and state updates back to server.
  - Provide standalone `rosrun` scripts for manual operations (takeoff, land, navigate, etc.).

### Useful Commands
- `rosrun clover simple_flight_test.py` — sample flight between markers 131→135.
- `python3 tools/delivery_flight.py` — run full delivery scenario (requires ROS environment).
- `pytest` in `tests/` — unit/integration tests (requires ROS mocks).

### Key Components
- `app/navigation/` — Clover API wrapper, ArUco map manager, navigation controller.
- `app/services/` — WebSocketService, DeliveryService orchestrating mission workflow.
- `config/` — drone configuration, ArUco map.
- `scripts/` — CLI/ROS scripts for takeoff, landing, marker navigation.
- `test_scripts/` — diagnostic scripts for manual verification.

### Documentation
- EN: [`docs/en/WORKFLOW.md`](../../docs/en/WORKFLOW.md), [`docs/en/ARCHITECTURE.md`](../../docs/en/ARCHITECTURE.md)
- RU: [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md), [`docs/ru/АРХИТЕКТУРА.md`](../../docs/ru/АРХИТЕКТУРА.md)

---

## Обзор (RU)
- Стек: Python 3, ROS (Clover), asyncio, WebSocket.
- Запускается на борту дрона (Raspberry Pi) и выполняет реальную миссию доставки.
- Обязанности:
  - Подключение к drone-service по WebSocket (`/ws/drone`).
  - Выполнение команд (взлет, перелёт к ArUco, посадка, сброс груза, возврат на базу).
  - Отправка телеметрии и статусов обратно на сервер.
  - Набор скриптов `rosrun` для ручного управления (взлёт, посадка, навигация и т.д.).

### Полезные команды
- `rosrun clover simple_flight_test.py` — тестовый полёт между маркерами 131→135.
- `python3 tools/delivery_flight.py` — полный сценарий доставки (требует ROS окружения).
- `pytest` в каталоге `tests/` — тесты (нужны ROS-заглушки).

### Основные компоненты
- `app/navigation/` — Clover API, менеджер карты ArUco, контроллер навигации.
- `app/services/` — WebSocketService, DeliveryService (управление миссией).
- `config/` — конфиг дрона, карта ArUco.
- `scripts/` — скрипты `rosrun` для операций (взлёт, посадка, навигация).
- `test_scripts/` — диагностические/учебные скрипты.

### Документация
- EN: [`docs/en/WORKFLOW.md`](../../docs/en/WORKFLOW.md), [`docs/en/ARCHITECTURE.md`](../../docs/en/ARCHITECTURE.md)
- RU: [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md), [`docs/ru/АРХИТЕКТУРА.md`](../../docs/ru/АРХИТЕКТУРА.md)
