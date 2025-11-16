# Drone Parser

## Overview (EN)

**Drone Parser** - это служба на Raspberry Pi дрона, которая:
- Подключается к drone-service через WebSocket
- Управляет полетами через ROS/Clover API
- Выполняет автономные миссии доставки
- Контролирует сброс груза через сервопривод
- Передает телеметрию и видео в реальном времени

### Stack
- **Python 3.8+**
- **ROS Noetic** (робототехническая платформа)
- **Clover** (пакет от COEX для управления дронами)
- **WebSockets** (связь с бэкендом)
- **pigpio** (управление GPIO для сервопривода)
- **OpenCV** (обработка видео с камеры)

### Key Features
- ✈️ **Автономные полеты** по ArUco маркерам
- 📦 **Автоматический сброс груза** через сервопривод
- 📡 **Real-time телеметрия** (батарея, позиция, статус)
- 📹 **Видеопоток** с камеры дрона
- 🔄 **Автоматический возврат** на базу после доставки
- 🛡️ **Безопасные режимы** (timeout, emergency return)
- 🔧 **Systemd служба** для автозапуска

---

## 📋 Быстрый старт

### 1. Установка на Raspberry Pi

```bash
# Подключитесь к дрону
ssh pi@192.168.11.1

# Клонируйте репозиторий
cd /home/pi
git clone <repo-url> hiTech
cd hiTech/backend/parsers/drone

# Установите зависимости
pip3 install -r requirements.txt
sudo apt-get install -y pigpio python3-pigpio

# Настройте конфигурацию
cp .env.example .env
nano .env  # Укажите DRONE_SERVICE_HOST и DRONE_IP

# Установите как службу
cd run_as_service
sudo ./install-service.sh
```

### 2. Запуск службы

```bash
# Запуск
sudo systemctl start drone-parser

# Проверка статуса
sudo systemctl status drone-parser

# Просмотр логов
sudo journalctl -u drone-parser -f
```

### 3. Ручное тестирование (без службы)

```bash
# Запуск main.py напрямую
cd /home/pi/hiTech/backend/parsers/drone
python3 main.py
```

---

## 📁 Структура проекта

```
drone/
├── main.py                      # Главный процесс (ROS bridge + WebSocket)
├── requirements.txt             # Python зависимости
├── .env.example                 # Пример конфигурации
├── README.md                    # Этот файл
├── DEPLOYMENT.md                # Полная инструкция по развертыванию
├── app/
│   ├── dependencies.py          # DI контейнер
│   ├── hardware/
│   │   ├── ros_bridge.py       # Мост ROS топиков
│   │   └── flight_manager.py   # Менеджер скриптов полета
│   ├── models/
│   │   └── schemas.py          # Pydantic модели
│   └── services/
│       ├── websocket_service.py # WebSocket клиент
│       └── delivery_service.py  # Логика доставки (legacy)
├── config/
│   └── config.py               # Настройки из .env
├── root_scripts/               # Скрипты для /root на дроне
│   ├── delivery_flight.py      # 🚀 Полет к постамату + сброс груза
│   ├── flight_back.py          # 🏠 Возврат на базу
│   ├── SERVO_SETUP.md          # Инструкция по сервоприводу
│   └── README.md               # Описание скриптов
└── run_as_service/             # Systemd служба
    ├── install-service.sh      # Скрипт установки
    ├── uninstall-service.sh    # Скрипт удаления
    └── drone-parser.service    # Systemd unit файл
```

---

## 🔧 Конфигурация

### Переменные окружения (.env)

```bash
# IP адреса
DRONE_IP=192.168.11.1              # IP дрона (Raspberry Pi)
DRONE_SERVICE_HOST=192.168.1.100  # IP сервера drone-service
DRONE_SERVICE_PORT=8001            # Порт WebSocket

# Камера
CAMERA_INDEX=0                     # /dev/video0
VIDEO_FPS=5                        # Частота кадров (5 FPS)

# Полет
CRUISE_ALTITUDE=1.5               # Высота полета (метры)
CRUISE_SPEED=0.5                  # Скорость полета (м/с)
LANDING_ALTITUDE=0.5              # Высота перед посадкой

# ROS
USE_CLOVER_API=true               # Использовать Clover API

# Логирование
LOG_LEVEL=INFO                    # DEBUG, INFO, WARNING, ERROR
```

---

## 🚁 Как это работает

### 1. Постоянная служба (main.py)

```python
# main.py запускается как systemd служба и работает постоянно
┌────────────────────────────────────┐
│  main.py (Drone Application)       │
├────────────────────────────────────┤
│  - ROS bridge (камера, батарея)   │
│  - WebSocket клиент к backend     │
│  - Слушает ROS топики:            │
│    • /drone/delivery/arrived       │
│    • /drone/delivery/drop_ready    │
│    • /drone/delivery/home_arrived  │
│  - Передает в backend через WS    │
└────────────────────────────────────┘
```

### 2. Получение задачи доставки

```
Backend → WebSocket → main.py → flight_manager.launch_delivery_flight()
                                       ↓
                              python3 /root/delivery_flight.py 135 131
```

### 3. Выполнение полета (delivery_flight.py)

```python
# delivery_flight.py - отдельный процесс
1. Взлет на 1.5м
2. Полет к ArUco маркеру 135 (постамат)
3. Посадка на постамат
4. Публикация → /drone/delivery/arrived
   ↓
   main.py ловит → отправляет в backend
   ↓
   Backend открывает ячейку → отправляет drop_cargo команду
   ↓
   main.py получает → публикует /drone/delivery/drop_confirm
5. delivery_flight.py получает подтверждение через ROS топик
6. Открытие сервопривода (0° → 180°)
7. Ожидание 10 секунд
8. Запуск python3 /root/flight_back.py
9. ЗАВЕРШЕНИЕ delivery_flight.py
```

### 4. Возврат домой (flight_back.py)

```python
# flight_back.py - новый процесс
1. Взлет (если нужно)
2. Полет к ArUco маркеру 131 (база)
3. Посадка на базу
4. Публикация → /drone/delivery/home_arrived
   ↓
   main.py ловит → отправляет статус IDLE в backend
5. ЗАВЕРШЕНИЕ flight_back.py
```

---

## 🔌 ROS Топики

### Публикуемые (Outgoing)

| Топик | Тип | Описание |
|-------|-----|----------|
| `/drone/delivery/arrived` | `PoseStamped` | Дрон прилетел к постамату |
| `/drone/delivery/drop_ready` | `PoseStamped` | Готов к сбросу груза |
| `/drone/delivery/home_arrived` | `PoseStamped` | Дрон вернулся на базу |

### Слушаемые (Incoming)

| Топик | Тип | Описание |
|-------|-----|----------|
| `/drone/delivery/drop_confirm` | `Bool` | Подтверждение сброса от backend |
| `/mavros/battery` | `BatteryState` | Состояние батареи |
| `/mavros/local_position/pose` | `PoseStamped` | Текущая позиция |
| `/main_camera/image_raw` | `Image` | Видеопоток с камеры |

### ROS Сервисы (Clover API)

| Сервис | Тип | Описание |
|--------|-----|----------|
| `/navigate` | `Navigate` | Полет к точке |
| `/get_telemetry` | `GetTelemetry` | Получить телеметрию |
| `/land` | `Trigger` | Посадка |

---

## 🎮 Управление сервоприводом

### Подключение

```
Сервопривод → Raspberry Pi
────────────────────────────
Brown (GND)  → Pin 6  (GND)
Red (5V)     → Pin 2  (5V)
Orange (PWM) → Pin 33 (GPIO 13)
```

### Использование в коде

```python
import pigpio

pi = pigpio.pi()
pi.set_mode(13, pigpio.OUTPUT)

# Закрыть дверь (0°)
pi.set_servo_pulsewidth(13, 500)

# Открыть дверь (180°)
pi.set_servo_pulsewidth(13, 2500)

# Выключить PWM
pi.set_servo_pulsewidth(13, 0)
pi.stop()
```

См. подробности в [`root_scripts/SERVO_SETUP.md`](root_scripts/SERVO_SETUP.md)

---

## 🐛 Отладка

### Просмотр логов

```bash
# Логи службы
sudo journalctl -u drone-parser -f

# Логи ROS
tail -f ~/.ros/log/latest/rosout.log

# Системные логи
dmesg | tail
```

### Проверка ROS

```bash
# Список топиков
rostopic list | grep drone

# Эхо топика
rostopic echo /drone/delivery/arrived

# Информация о топике
rostopic info /mavros/battery
```

### Проверка WebSocket

```bash
# Тест подключения
nc -zv <DRONE_SERVICE_HOST> 8001

# Проверка конфига
cat .env | grep DRONE_SERVICE
```

### Ручной запуск скрипта

```bash
# ВНИМАНИЕ: Дрон взлетит! Используйте только для теста!
cd /root
python3 delivery_flight.py 135 131
```

---

## 📚 Дополнительная документация

- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Полная инструкция по развертыванию на Raspberry Pi
- **[root_scripts/README.md](root_scripts/README.md)** - Описание скриптов полета
- **[root_scripts/SERVO_SETUP.md](root_scripts/SERVO_SETUP.md)** - Настройка сервопривода
- **[Clover Documentation](https://clover.coex.tech/ru/)** - Официальная документация Clover

---

## 🔐 Безопасность

### Режимы отказа

1. **Timeout на подтверждение** (10 сек)
   - Если backend не отвечает → аварийный возврат на базу
   
2. **Потеря связи WebSocket**
   - Автоматическое переподключение каждые 5 секунд
   
3. **Ошибка сервопривода**
   - Симуляция сброса груза (без реального действия)
   
4. **Ошибка ROS/Clover**
   - Логирование + fallback в безопасный режим

### Emergency Stop

```bash
# Остановка службы
sudo systemctl stop drone-parser

# Посадка дрона (если в воздухе)
rosservice call /land

# Экстренное выключение моторов
rosservice call /mavros/cmd/arming "value: false"
```

---

## 🧪 Тестирование

### Unit Tests

```bash
cd /home/pi/hiTech/backend/parsers/drone
pytest tests/ -v
```

### Integration Tests

```bash
# Тест WebSocket подключения
python3 tests/test_websocket_connection.py

# Тест ROS топиков
python3 tests/test_ros_topics.py

# Тест сервопривода
python3 tests/test_servo.py
```

### Симуляция в Gazebo

```bash
# Запуск симуляции Clover
roslaunch clover_simulation simulator.launch

# В новом терминале
cd /home/pi/hiTech/backend/parsers/drone
python3 main.py
```

---

## ⚠️ Важно

- **НЕ ЗАПУСКАЙТЕ** скрипты полета без проверки окружения
- **ВСЕГДА ИМЕЙТЕ** ручное управление как backup
- **ТЕСТИРУЙТЕ** новый код в симуляции перед реальными полетами
- **СЛЕДИТЕ** за зарядом батареи (минимум 30% для полета)
- **ПРОВЕРЯЙТЕ** ArUco карту перед полетом

---

## 📄 Лицензия

[MIT License](../../LICENSE)
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
