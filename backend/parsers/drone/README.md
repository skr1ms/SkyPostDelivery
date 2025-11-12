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
- `./start_drone.sh` — **main startup script** (sources ROS, runs WebSocket bridge).
- `python3 test_scripts/simple_flight_test.py` — manual flight test between markers 131→135.
- `python3 tools/delivery_flight.py` — standalone delivery scenario (requires ROS).
- `pytest` in `tests/` — unit/integration tests (requires ROS mocks).

### How to Run
1. Install ROS dependencies (see `requirements.txt`).
2. Copy `.env.example` to `.env` and configure drone settings.
3. Source ROS: `source /opt/ros/noetic/setup.bash && source ~/catkin_ws/devel/setup.bash`.
4. Run: `./start_drone.sh`

The drone will connect to `drone-service` via WebSocket, subscribe to ROS topics (`/main_camera/image_raw`, `/mavros/battery`, etc.), and wait for delivery commands.

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
- `./start_drone.sh` — **основной скрипт запуска** (инициализирует ROS, запускает WebSocket-мост).
- `python3 test_scripts/simple_flight_test.py` — ручной тестовый полёт между маркерами 131→135.
- `python3 tools/delivery_flight.py` — автономный сценарий доставки (требует ROS).
- `pytest` в каталоге `tests/` — тесты (нужны ROS-заглушки).

### Как запустить
1. Установить ROS-зависимости (см. `requirements.txt`).
2. Скопировать `.env.example` в `.env` и настроить параметры дрона.
3. Активировать ROS: `source /opt/ros/noetic/setup.bash && source ~/catkin_ws/devel/setup.bash`.
4. Запустить: `./start_drone.sh`

Дрон подключится к `drone-service` по WebSocket, подпишется на ROS-топики (`/main_camera/image_raw`, `/mavros/battery` и т.д.) и будет ожидать команды на доставку.

### Основные компоненты
- `app/navigation/` — Clover API, менеджер карты ArUco, контроллер навигации.
- `app/services/` — WebSocketService, DeliveryService (управление миссией).
- `config/` — конфиг дрона, карта ArUco.
- `scripts/` — скрипты `rosrun` для операций (взлёт, посадка, навигация).
- `test_scripts/` — диагностические/учебные скрипты.

### Документация
- EN: [`docs/en/WORKFLOW.md`](../../docs/en/WORKFLOW.md), [`docs/en/ARCHITECTURE.md`](../../docs/en/ARCHITECTURE.md)
- RU: [`docs/ru/СЦЕНАРИЙ РАБОТЫ.md`](../../docs/ru/СЦЕНАРИЙ%20РАБОТЫ.md), [`docs/ru/АРХИТЕКТУРА.md`](../../docs/ru/АРХИТЕКТУРА.md)
