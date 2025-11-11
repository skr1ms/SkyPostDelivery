# Полная спецификация API и протоколов hiTech

Платформа hiTech предоставляет несколько интерфейсов — REST поверх HTTP, каналы WebSocket и gRPC-сервисы — для оркестрации автономной доставки между дронами, постаматами и клиентскими приложениями. Этот документ подробно описывает **каждую структуру запроса и ответа**, чтобы интеграторы могли реализовать клиентов без изучения исходного кода.

---

## Содержание
1. [Условные обозначения](#условные-обозначения)
2. [Аутентификация и жизненный цикл сессии](#аутентификация-и-жизненный-цикл-сессии)
3. [Общая модель ошибок](#общая-модель-ошибок)
4. [HTTP API (оркестратор, админ, мобильное приложение)](#http-api-оркестратор-админ-мобильное-приложение)
   - [Пользователи и аутентификация](#пользователи-и-аутентификация)
   - [Устройства для уведомлений](#устройства-для-уведомлений)
   - [Каталог товаров](#каталог-товаров)
   - [Постаматы](#постаматы)
   - [Ячейки](#ячейки)
   - [Заказы](#заказы)
   - [Доставки](#доставки)
   - [Дроны](#дроны)
   - [Операции с QR](#операции-с-qr)
   - [Мониторинг системы](#мониторинг-системы)
   - [Вспомогательные endpoints](#вспомогательные-endpoints)
5. [HTTP API сервиса дронов](#http-api-сервиса-дронов)
6. [gRPC API](#grpc-api)
7. [WebSocket API](#websocket-api)
   - [Дрон ↔ сервис дронов](#дрон--сервис-дронов)
   - [Админ-панель ↔ сервис дронов](#админ-панель--сервис-дронов)
   - [Канал видеопрокси](#канал-видеопрокси)
8. [Справочник моделей данных](#справочник-моделей-данных)
9. [Статусные перечисления](#статусные-перечисления)
10. [Управление изменениями](#управление-изменениями)

---

## Условные обозначения
- **Базовые URL**
  - REST (оркестратор): `https://{env-host}/api/v1`
  - WebSocket (сервис дронов): `wss://{drone-service-host}/ws/...`
  - gRPC: `grpc://{orchestrator-host}:{GO_GRPC_PORT}`
- **Типы контента**: запросы с телом должны использовать `Content-Type: application/json`. Ответы по умолчанию в JSON, если не указано иное.
- **Идентификаторы**: строки UUID v4 (пример: `4e2a6a94-9c9d-4c54-a9fb-6b0f8dd695b5`), если не сказано иначе.
- **Отметки времени**: ISO-8601 в UTC (например, `2025-11-11T10:15:30Z`) в ответах; запросы принимают ISO-8601 или RFC3339.
- **Числа**: десятичные значения передаются в формате IEEE-754 double (`float64` в Go). Размеры — в **метрах**, масса — в **килограммах**.
- **Локализация**: все сообщения сервера (оркестратор) возвращаются на английском. Содержимое SMS зависит от шаблонов SMS Aero.
- **Ограничение частоты**: отдельные auth-endpoints имеют лимиты по IP. При превышении возвращается `429 Too Many Requests` со стандартным телом ошибки.

---

## Аутентификация и жизненный цикл сессии

### Модель токенов
- **Access Token**: JWT, подписанный `HS256`, стандартный TTL — 7 дней. Нужен в заголовке `Authorization: Bearer <access_token>` для защищённых endpoints.
- **Refresh Token**: JWT, подписанный `HS256`, стандартный TTL — 14 дней. Передаётся в `Authorization` при вызове `/auth/refresh`.
- **Роли**: `user`, `admin`. Роль зашита в токен и применяется для авторизации на сервере.
- **Подтверждение телефона**: регистрация и логин используют SMS-коды через SMS Aero.

### Последовательности

#### Регистрация (email + телефон)
1. Клиент отправляет `POST /api/v1/auth/register` с данными профиля.
2. Сервер создаёт пользователя с `phone_verified=false`, отправляет SMS-код.
3. Клиент отправляет `POST /api/v1/auth/verify/phone` с кодом и получает токены + полезную нагрузку для QR.

#### Логин по email
1. Клиент отправляет `POST /api/v1/auth/login`.
2. Сервер проверяет учётные данные и возвращает пару токенов и данные по QR.

#### Логин по телефону
1. Клиент отправляет `POST /api/v1/auth/login/phone` с полем `phone`.
2. Отправляется SMS. Клиент повторно вызывает `POST /api/v1/auth/verify/phone`, чтобы обменять код на токены.

#### Сброс пароля
1. Инициализация: `POST /api/v1/auth/password/reset/request`.
2. Подтверждение: `POST /api/v1/auth/password/reset`.

#### Обновление токена
1. Клиент вызывает `POST /api/v1/auth/refresh` с `Authorization: Bearer <refresh_token>`.
2. Сервер возвращает новую пару токенов.

### Требуемые заголовки
| Заголовок | Значение | Комментарий |
| --- | --- | --- |
| `Authorization` | `Bearer <access_token>` | Обязателен для защищённых REST endpoints. |
| `Content-Type` | `application/json` | Для запросов с JSON-телом. |
| `Accept` | `application/json` | Рекомендуется. |

---

## Общая модель ошибок
- Все ответы, отличные от 2xx (кроме пустого 204), имеют вид:
  ```json
  {
    "error": "описание проблемы"
  }
  ```
- Ошибки валидации приходят из Gin и содержат поле/ограничение (например, `"error": "Key: 'CreateGoodRequest.Weight' Error:Field validation for 'Weight' failed on the 'gt' tag"`).
- Типовые коды состояния:
  - `400 Bad Request`: некорректные данные, нарушение бизнес-правил.
  - `401 Unauthorized`: отсутствует/некорректный токен, неверный SMS/QR-код.
  - `403 Forbidden`: у аутентифицированного пользователя нет прав.
  - `404 Not Found`: ресурс не найден.
  - `409 Conflict`: зарезервирован под будущие случаи дублирования.
  - `500 Internal Server Error`: необработанное исключение на сервере.

---

## HTTP API (оркестратор, админ, мобильное приложение)

**Базовый URL**: `https://{host}/api/v1`  
Все защищённые endpoints требуют действующего access token, если не указано иное.

### Пользователи и аутентификация

#### `POST /auth/register`
Регистрирует нового пользователя и инициирует отправку SMS-кода.

- **Авторизация**: не требуется.
- **Лимиты**: есть (по IP).
- **Тело запроса**

| Поле | Тип | Обязательность | Ограничения | Описание |
| --- | --- | --- | --- | --- |
| `full_name` | string | Да | 1..255 символов | ФИО пользователя. |
| `email` | string | Да | корректный email | Уникальная почта для логина. |
| `phone` | string | Да | цифры, желательно с кодом страны | Номер телефона для SMS.
| `password` | string | Да | ≥ 6 символов | Пароль (хранится в хеше). |

- **Пример запроса**
```json
{
  "full_name": "Alice Robotics",
  "email": "alice@example.com",
  "phone": "79001234567",
  "password": "Sup3rSecure!"
}
```
- **Ответы**
  - `201 Created`
    ```json
    {
      "message": "Registration successful. SMS code sent to 79001234567"
    }
    ```
  - `400 Bad Request` — ошибка валидации.
  - `500 Internal Server Error` — ошибка при сохранении или отправке SMS.

#### `POST /auth/verify/phone`
Подтверждает SMS-код и выдает JWT + данные для QR.

- **Авторизация**: нет.
- **Тело запроса**

| Поле | Тип | Обязательность | Ограничения | Описание |
| --- | --- | --- | --- | --- |
| `phone` | string | Да | совпадает с номером регистрации | Телефон пользователя. |
| `code` | string | Да | ровно 4 цифры | Полученный по SMS код. |

- **Пример запроса**
```json
{
  "phone": "79001234567",
  "code": "5842"
}
```
- **Успех 200**
```json
{
  "user": {
    "id": "0f28f2d9-1e61-4f8f-8d65-b8efcd2c3a8c",
    "full_name": "Alice Robotics",
    "email": "alice@example.com",
    "phone_number": "79001234567",
    "phone_verified": true,
    "created_at": "2025-11-11T08:12:44Z",
    "role": "user",
    "qr_issued_at": "2025-11-11T08:12:44Z",
    "qr_expires_at": "2025-11-18T08:12:44Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1762851164,
  "qr_code": "iVBORw0KGgoAAAANSUhEUgAA..."
}
```
- **Ошибки**
  - `400`: некорректные данные/телефон не подтверждён.
  - `401`: неверный или просроченный код.
  - `500`: ошибка генерации токена.

#### `POST /auth/login`
Аутентификация по email/паролю.

- **Тело запроса**

| Поле | Тип | Обязательность | Описание |
| --- | --- | --- | --- |
| `email` | string | Да | Зарегистрированный email. |
| `password` | string | Да | Обычный пароль (≥ 6 символов). |

- **Ответы**: аналогично `/auth/verify/phone`.
- **Ошибки**: `400` (валидация), `401` (неверные данные), `500`.

#### `POST /auth/login/phone`
Отправляет SMS-код для входа.

- **Тело запроса**
  ```json
  { "phone": "79001234567" }
  ```
- **Успех 200**
  ```json
  { "message": "Login code sent to your phone" }
  ```
- **Ошибки**: `400` (неверный номер), `404` (нет пользователя), `500` (ошибка SMS).

#### `POST /auth/password/reset/request`
Отправляет SMS с кодом для сброса пароля.

- **Тело запроса**
  ```json
  { "phone": "79001234567" }
  ```
- **Ответ**: `200` — тот же формат, что и при логине по телефону.
- **Ошибки**: `400`/`404`/`500`.

#### `POST /auth/password/reset`
Применяет новый пароль после подтверждения кода.

- **Тело запроса**

| Поле | Тип | Обязательность | Ограничения |
| --- | --- | --- | --- |
| `phone` | string | Да | существующий пользователь |
| `code` | string | Да | 4-значный код |
| `new_password` | string | Да | ≥ 6 символов |

- **Успех 200**
  ```json
  { "message": "Password reset successfully" }
  ```
- **Ошибки**: `400` (неверный код), `500`.

#### `POST /auth/refresh`
Обмен refresh-токена на новую пару токенов.

- **Заголовки**: `Authorization: Bearer <refresh_token>`
- **Успех 200**
  ```json
  {
    "access_token": "new-access",
    "refresh_token": "new-refresh",
    "expires_at": 1762851164
  }
  ```
- **Ошибки**: `401` (недействительный refresh), `500`.

#### `GET /auth/me`
Возвращает профиль пользователя и данные по QR.

- **Заголовки**: требуется access token.
- **Успех 200**: как `/auth/verify/phone`.
- **Ошибки**: `401`, `404` (если пользователь не найден).

### Устройства для уведомлений

#### `POST /users/{id}/devices`
Регистрирует FCM/APNS токен для push-уведомлений.

- **Авторизация**: Bearer; пользователь может регистрировать только своё устройство.
- **Параметры пути**

| Имя | Тип | Описание |
| --- | --- | --- |
| `id` | UUID | ID пользователя. |

- **Тело запроса**

| Поле | Тип | Обязательность | Описание |
| --- | --- | --- | --- |
| `token` | string | Да | Токен FCM/APNS. |
| `platform` | string | Да | `android`, `ios` и т.п. |

- **Пример**
```json
{
  "token": "fcm:dyk3Nz8...",
  "platform": "android"
}
```
- **Ответы**
  - `200 OK`
    ```json
    { "message": "Device registered" }
    ```
    Если нет Firebase-конфигурации, возвращается сообщение `"Notifications are disabled"`.
  - Ошибки: `400` (валидация), `401`, `403`, `500`.

### Каталог товаров

Базовый путь: `/goods` (защищённый).

#### `GET /goods`
- **Назначение**: список товаров с остатками.
- **Ответ 200**
```json
[
  {
    "id": "4a9b65f2-4f8a-4f3d-9d52-9aca43b0c8a1",
    "name": "Medical Kit",
    "weight": 1.3,
    "height": 0.18,
    "length": 0.30,
    "width": 0.22,
    "quantity_available": 15
  }
]
```
- **Ошибки**: `500`.

#### `POST /goods`
Создаёт **N** одинаковых товаров (через поле `quantity`).

- **Тело запроса**

| Поле | Тип | Обязательность | Ограничения | Описание |
| --- | --- | --- | --- | --- |
| `name` | string | Да | 1..255 символов | Название продукта. |
| `weight` | number | Да | `> 0` | Масса в килограммах. |
| `height` | number | Да | `> 0` | Высота в метрах. |
| `length` | number | Да | `> 0` | Длина в метрах. |
| `width` | number | Да | `> 0` | Ширина в метрах. |
| `quantity` | integer | Да | `> 0` | Кол-во экземпляров. |

- **Пример**
```json
{
  "name": "First Aid Kit",
  "weight": 1.3,
  "height": 0.18,
  "length": 0.30,
  "width": 0.22,
  "quantity": 5
}
```
- **Успех 201**: массив созданных товаров.
- **Ошибки**: `400`, `500`.

#### `GET /goods/{id}`
Возвращает товар по UUID.

- **Ошибки**: `400`, `404`.

#### `PATCH /goods/{id}`
Обновляет размеры и название (quantity неизменяем).

- **Тело запроса**: то же, что при создании, без `quantity`.
- **Успех 200**: обновлённый товар.
- **Ошибки**: `400`, `500`.

#### `DELETE /goods/{id}`
Удаляет товар.

- **Успех 200**
  ```json
  { "message": "Good deleted successfully" }
  ```
- **Ошибки**: `400`, `500`.

### Постаматы

Защищённые административные endpoints требуют авторизации, если не указано иное. Публичные (`/automats/qr-scan`, `/automats/confirm-pickup`) доступны без токена для взаимодействия с локальными устройствами.

#### `POST /automats`
Создаёт постамат с калибровкой ячеек.

- **Тело запроса**

| Поле | Тип | Обязательность | Описание |
| --- | --- | --- | --- |
| `city` | string | Да | Город. |
| `address` | string | Да | Адрес. |
| `ip_address` | string | Нет | Опциональный статический IP. |
| `coordinates` | string | Нет | Произвольные координаты (`"55.751244,37.618423"`). |
| `aruco_id` | integer | Да | ID ArUco-маркера для посадки. |
| `number_of_cells` | integer | Да | Должен совпадать с длиной `cells`. |
| `cells` | array | Да | Массив объектов размеров. |

Элемент `cells`:

| Поле | Тип | Обязательность | Ограничения |
| --- | --- | --- | --- |
| `height` | number | Да | `> 0` |
| `length` | number | Да | `> 0` |
| `width` | number | Да | `> 0` |

- **Пример**
```json
{
  "city": "Ekaterinburg",
  "address": "Industrial Zone, Building 5",
  "ip_address": "192.168.10.15",
  "coordinates": "56.8380,60.5975",
  "aruco_id": 42,
  "number_of_cells": 2,
  "cells": [
    { "height": 0.4, "length": 0.5, "width": 0.5 },
    { "height": 0.3, "length": 0.4, "width": 0.3 }
  ]
}
```
- **Успех 201**
```json
{
  "id": "a941aa86-b0ab-4f75-a7b9-4dec56dde16f",
  "ip_address": "192.168.10.15",
  "city": "Ekaterinбург",
  "address": "Industrial Zone, Building 5",
  "number_of_cells": 2,
  "coordinates": "56.8380,60.5975",
  "aruco_id": 42,
  "is_working": true
}
```

#### `GET /automats`
Список всех постаматов.

#### `GET /automats/working`
Фильтр `is_working=true`.

#### `GET /automats/{id}`
Возвращает постамат по UUID.

#### `PUT /automats/{id}`
Обновляет метаданные:

| Поле | Тип | Обязательность |
| --- | --- | --- |
| `city` | string | Да |
| `address` | string | Да |
| `ip_address` | string | Нет |
| `coordinates` | string | Нет |

#### `GET /automats/{id}/cells`
Возвращает массив ячеек постамата.

#### `PATCH /automats/{id}/cells/{cellId}`
Обновляет размеры ячейки.

- **Тело запроса**: как `UpdateCellRequest`.
- **Ответ 200**
  ```json
  {
    "id": "c7b84618-6242-49eb-87e4-604b0e54c7a5",
    "post_id": "a941aa86-b0ab-4f75-a7b9-4dec56dde16f",
    "height": 0.5,
    "length": 0.6,
    "width": 0.45,
    "status": "available"
  }
  ```

#### `PATCH /automats/{id}/status`
Переключает доступность постамата.

- **Тело запроса**
  ```json
  { "is_working": true }
  ```
- **Ответ 200**: обновлённый постамат.

#### `DELETE /automats/{id}`
Удаляет постамат.

- **Успех 200**
  ```json
  { "message": "Parcel automat deleted successfully" }
  ```

#### `POST /automats/qr-scan` *(публичный)*
Используется контроллером постамата для обработки QR-кода клиента.

- **Тело запроса**

| Поле | Тип | Обязательность | Описание |
| --- | --- | --- | --- |
| `qr_data` | string | Да | Сырые данные QR. |
| `parcel_automat_id` | string | Да | UUID постамата. |

- **Успех 200**
```json
{
  "success": true,
  "message": "QR code processed successfully",
  "cell_ids": [
    "c7b84618-6242-49eb-87e4-604b0e54c7a5",
    "868f662d-2940-477a-9de7-8232d6ad91aa"
  ]
}
```
- **Ошибки**: `400` (неверный QR/ID).

#### `POST /automats/confirm-pickup` *(публичный)*
Подтверждает, что указанные ячейки освобождены клиентом.

- **Тело запроса**
  ```json
  {
    "cell_ids": [
      "c7b84618-6242-49eb-87e4-604b0e54c7a5",
      "868f662d-2940-477a-9de7-8232d6ad91aa"
    ]
  }
  ```
- **Успех 200**
  ```json
  {
    "success": true,
    "message": "Pickup confirmed successfully"
  }
  ```
- **Ошибки**: `400`, `500`.

### Ячейки

#### `GET /locker/cells/{id}`
- **Авторизация**: Bearer.
- **Ответ 200**
```json
{
  "id": "c7b84618-6242-49eb-87e4-604b0e54c7a5",
  "post_id": "a941aa86-b0ab-4f75-a7b9-4dec56dde16f",
  "height": 0.5,
  "length": 0.6,
  "width": 0.45,
  "status": "occupied"
}
```
- **Ошибки**: `400`, `404`.

### Заказы

#### `POST /orders`
Создаёт заказ для аутентифицированного пользователя.

- **Тело запроса**
  ```json
  { "good_id": "4a9b65f2-4f8a-4f3d-9d52-9aca43b0c8a1" }
  ```
- **Успех 201**
  ```json
  {
    "id": "2f611f70-0264-4f1f-a5f8-4a07b8942a60",
    "user_id": "0f28f2d9-1e61-4f8f-8d65-b8efcd2c3a8c",
    "good_id": "4a9b65f2-4f8a-4f3d-9d52-9aca43b0c8a1",
    "parcel_automat_id": "a941aa86-b0ab-4f75-a7b9-4dec56dde16f",
    "locker_cell_id": "c7b84618-6242-49eb-87e4-604b0e54c7a5",
    "status": "pending",
    "created_at": "2025-11-11T08:30:00Z"
  }
  ```
- **Ошибки**: `400`, `401`, `500`.

#### `POST /orders/batch`
Создаёт несколько заказов за один запрос (по разным товарам).

- **Тело запроса**
  ```json
  {
    "good_ids": [
      "4a9b65f2-4f8a-4f3d-9d52-9aca43b0c8a1",
      "71e3ba1d-8e06-4127-996f-afdb238be6da"
    ]
  }
  ```
- **Успех 201**: массив заказов.
- **Ошибки**: аналогично одиночному созданию.

#### `GET /orders/{id}`
Возвращает заказ по UUID; ошибки: `400`, `404`.

#### `POST /orders/{id}/return`
Отменяет заказ и освобождает ресурсы, если владелец прерывает доставку.

- **Auth**: Bearer.
- **Доступно**, пока статус заказа `pending` или `in_progress`.
- **Эффекты**:
  - забронированная ячейка возвращается в состояние `available`;
  - товар снова доступен на складе (увеличивается количество);
  - в RabbitMQ публикуется приоритетная задача `delivery.return`, дрон разворачивается на базовую ArUco-метку `131`.
- **Тело запроса**: отсутствует.
- **Успех 200**
  ```json
  { "success": true }
  ```
- **Ошибки**: `400` (невалидный UUID или заказ уже нельзя отменить), `401`, `403` (попытка отменить чужой заказ), `404` заказ не найден, `500`.

#### `GET /orders/user/{userId}`
Возвращает заказы пользователя с данными товара.

- **Ответ 200**
```json
[
  {
    "id": "2f611f70-0264-4f1f-a5f8-4a07b8942a60",
    "user_id": "0f28f2d9-1e61-4f8f-8d65-b8efcd2c3a8c",
    "good_id": "4a9b65f2-4f8a-4f3d-9d52-9aca43b0c8a1",
    "parcel_automat_id": "a941aa86-b0ab-4f75-a7b9-4dec56dde16f",
    "status": "pending",
    "created_at": "2025-11-11T08:30:00Z",
    "good": {
      "id": "4a9b65f2-4f8a-4f3d-9d52-9aca43b0c8a1",
      "name": "Medical Kit",
      "weight": 1.3,
      "height": 0.18,
      "length": 0.30,
      "width": 0.22,
      "quantity_available": 10
    }
  }
]
```

### Доставки

#### `GET /deliveries/{id}`
Возвращает метаданные доставки:
```json
{
  "id": "8f264378-6a59-4e6f-9ada-1955f9cce0b1",
  "orderID": "2f611f70-0264-4f1f-a5f8-4a07b8942a60",
  "droneID": "b84a0fda-60c2-4af7-8e81-dca9a2681a07",
  "parcelAutomatID": "a941aa86-b0ab-4f75-a7b9-4dec56dde16f",
  "status": "in_transit"
}
```

#### `PUT /deliveries/{id}/status`
Обновляет статус доставки.

- **Тело запроса**
  ```json
  { "status": "completed" }
  ```
- **Успех 200**
  ```json
  { "success": true }
  ```
- **Ошибки**: `400`, `500`.

#### `POST /deliveries/confirm-loaded`
Подтверждает загрузку дрона из ячейки.

- **Тело запроса**
  ```json
  {
    "order_id": "2f611f70-0264-4f1f-a5f8-4a07b8942a60",
    "locker_cell_id": "c7b84618-6242-49eb-87e4-604b0e54c7a5",
    "timestamp": "2025-11-11T09:12:05Z"
  }
  ```
- **Успех 200**
  ```json
  {
    "success": true,
    "message": "Goods loaded confirmed successfully"
  }
  ```
- **Ошибки**: `400`, `500`.

### Дроны

#### `GET /drones`
Список всех дронов.

#### `POST /drones`
- **Тело запроса**
  ```json
  {
    "model": "DJI Matrice 30",
    "ip_address": "192.168.20.11"
  }
  ```
- **Успех 201**
  ```json
  {
    "id": "b84a0fda-60c2-4af7-8e81-dca9a2681a07",
    "model": "DJI Matrice 30",
    "ip_address": "192.168.20.11",
    "status": "idle"
  }
  ```

#### `GET /drones/{id}`
Возвращает данные дрона.

#### `PUT /drones/{id}`
Обновляет модель/IP.

- **Тело запроса**
  ```json
  { "model": "DJI Matrice 30T", "ip_address": "192.168.20.20" }
  ```

#### `PATCH /drones/{id}/status`
Изменяет статус дрона.

- **Тело запроса**
  ```json
  { "status": "busy" }
  ```
- **Успех 200**: обновлённая запись.

#### `GET /drones/{id}/status`
Возвращает актуальный статус дрона, агрегированный сервисом дронов.

- **Пример ответа**
```json
{
  "drone_id": "b84a0fda-60c2-4af7-8e81-dca9a2681a07",
  "status": "in_flight",
  "battery_level": 62.5,
  "position": {
    "latitude": 56.83801,
    "longitude": 60.59747,
    "altitude": 120.4
  },
  "speed": 8.1,
  "current_delivery_id": "8f264378-6a59-4e6f-9ada-1955f9cce0b1",
  "error_message": null
}
```

#### `DELETE /drones/{id}`
Удаляет дрона.

- **Успех 200**
  ```json
  { "message": "Drone deleted successfully" }
  ```

### Операции с QR

#### `POST /qr/validate`
- **Авторизация**: нет (используется сервисами постаматов).
- **Тело запроса**
  ```json
  { "qr_data": "base64-encoded-payload" }
  ```
- **Успех 200**
  ```json
  {
    "valid": true,
    "user_id": "0f28f2d9-1e61-4f8f-8d65-b8efcd2c3a8c",
    "email": "alice@example.com",
    "name": "Alice Robotics",
    "expires_at": 1762851164
  }
  ```
- **Ошибки**: `400`, `401`.

#### `POST /qr/refresh`
- **Авторизация**: Bearer.
- **Успех 200**
  ```json
  {
    "qr_code": "iVBORw0KGgoAAAANSUhEUgAA...",
    "expires_at": 1762851164
  }
  ```
- **Ошибки**: `401`, `500`.

### Мониторинг системы

#### `GET /monitoring/system-status`
Агрегирует точки данных по дронам, постаматам и активным доставкам.

- **Пример ответа**
```json
{
  "drones": [
    {
      "id": "b84a0fda-60c2-4af7-8e81-dca9a2681a07",
      "model": "DJI Matrice 30",
      "ip_address": "192.168.20.11",
      "status": "in_flight"
    }
  ],
  "automats": [
    {
      "id": "a941aa86-b0ab-4f75-a7b9-4dec56dde16f",
      "ip_address": "192.168.10.15",
      "city": "Ekaterinburg",
      "address": "Industrial Zone, Building 5",
      "number_of_cells": 2,
      "coordinates": "56.8380,60.5975",
      "aruco_id": 42,
      "is_working": true,
      "cells": [
        {
          "id": "c7b84618-6242-49eb-87e4-604b0e54c7a5",
          "post_id": "a941aa86-b0ab-4f75-a7b9-4dec56dde16f",
          "height": 0.5,
          "length": 0.6,
          "width": 0.45,
          "status": "occupied"
        }
      ]
    }
  ],
  "active_deliveries": [
    {
      "id": "8f264378-6a59-4e6f-9ada-1955f9cce0b1",
      "orderID": "2f611f70-0264-4f1f-a5f8-4a07b8942a60",
      "droneID": "b84a0fda-60c2-4af7-8e81-dca9a2681a07",
      "parcelAutomatID": "a941aa86-b0ab-4f75-a7b9-4dec56dde16f",
      "status": "in_transit"
    }
  ]
}
```

### Вспомогательные endpoints
- `GET /metrics`: отдает метрики Prometheus (текстовый формат).
- `GET /swagger/*`: Swagger UI (`github.com/swaggo/gin-swagger`).

---

## HTTP API сервиса дронов

Базовый URL: `https://{drone-service-host}` (либо внутри сети оркестратора). Все endpoints реализованы на FastAPI в `backend/drone-service`.

### Здоровье и метрики
- `GET /health` → `{"status": "healthy", "service": "drone-service"}`
- `GET /status` → состояние первого подключённого дрона или заглушка.
- `GET /metrics` → метрики Prometheus.

### Endpoint команд

#### `POST /api/drones/{drone_id}/command`
Отправляет команду подключённому дрону через WebSocket.

- **Авторизация**: пока нет (предполагается защита на уровне сети).
- **Параметры пути**

| Имя | Тип | Описание |
| --- | --- | --- |
| `drone_id` | string | UUID, назначенный оркестратором. |

- **Тело запроса**
  ```json
  { "command": "return_home" }
  ```
  Поддерживается только `"return_home"`.

- **Ответы**
  - `200 OK`
    ```json
    { "success": true, "message": "Return home command sent" }
    ```
  - `404`
    ```json
    { "success": false, "message": "Drone not connected" }
    ```
  - `400` — неизвестная команда.

---

## gRPC API

### Endpoint
- Сервис: `pb.OrchestratorService`
- Адрес по умолчанию: `:{GO_GRPC_PORT}` (по умолчанию 50052).
- Транспортная безопасность: зависит от окружения (в staging/prod за TLS-прокси).

### RPC

#### `RequestCellOpen`
Вызывается контроллером постамата, когда дрон прибыл и необходимо открыть ячейку.

- **Запрос**
  ```proto
  message CellOpenRequest {
    string order_id = 1;           // UUID заказа
    string parcel_automat_id = 2;  // UUID постамата
  }
  ```
- **Ответ**
  ```proto
  message CellOpenResponse {
    bool success = 1;
    string message = 2;  // статус
    string cell_id = 3;  // UUID открытой ячейки при success=true
  }
  ```

- **Пример `grpcurl`**
  ```bash
  grpcurl -plaintext \
    -d '{"order_id":"2f611f70-0264-4f1f-a5f8-4a07b8942a60","parcel_automat_id":"a941aa86-b0ab-4f75-a7b9-4dec56dde16f"}' \
    orchestrator:50052 pb.OrchestratorService/RequestCellOpen
  ```
- **Успех**
  ```json
  {
    "success": true,
    "message": "Cell opened",
    "cellId": "c7b84618-6242-49eb-87e4-604b0e54c7a5"
  }
  ```
- **Ошибка**
  ```json
  {
    "success": false,
    "message": "No free cells",
    "cellId": ""
  }
  ```

---

## WebSocket API

### Дрон ↔ сервис дронов

- **URL**: `wss://{drone-service-host}/ws/drone`
- **Назначение**: двусторонний канал телеметрии и команд между физическим дроном и backend-ом.
- **Handshake**
  1. Клиент открывает WS и сразу отправляет сообщение регистрации:
     ```json
     {
       "type": "register",
       "ip_address": "192.168.20.11"
     }
     ```
  2. Сервер проверяет IP по базе оркестратора и отвечает:
     ```json
     {
       "type": "registered",
       "drone_id": "b84a0fda-60c2-4af7-8e81-dca9a2681a07",
       "timestamp": "2025-11-11T09:14:20.123456"
     }
     ```
  3. Далее сообщения могут идти в обе стороны.

- **Входящие типы (от дрона)**

| Тип | Payload | Описание |
| --- | --- | --- |
| `heartbeat` | `{ "battery_level": float, "position": {...}, "status": "idle|in_flight|returning", "speed": float, "current_delivery_id": string?, "error_message": string? }` | Периодический сигнал здоровья. |
| `status_update` | те же поля, что и heartbeat | Произвольное обновление состояния (устаревшее, вместо него предпочтительнее heartbeat). |
| `delivery_update` | `{ "delivery_id": string, "drone_status": "arrived_at_locker|returning|..." }` | Прогресс доставки. |
| `arrived_at_destination` | `{ "order_id": string, "parcel_automat_id": string }` | Сигнал о прибытии; оркестратор откроет ячейку через gRPC. |
| `cargo_dropped` | `{ "order_id": string, "locker_cell_id": string }` | Груз сброшен. |

- **Исходящие (от сервера)**

| Тип | Payload | Описание |
| --- | --- | --- |
| `delivery_task` | объект задачи (маршрут, координаты) | Отправляется при назначении доставки. |
| `command` | `{ "command": "return_home" }` | Реaltime-команда, проксируемая из REST. |
| `error` | `{ "message": string }` | Уведомление об ошибке (например, неверная регистрация). |

- **Обрыв соединения**: при отключении сервер снимает регистрацию дрона.

### Админ-панель ↔ сервис дронов

- **URL**: `wss://{drone-service-host}/ws/admin`
- **Назначение**: админ-дэшборд получает живые данные.
- **Поведение**
  - Клиент может послать `{"type": "ping"}`; сервер отвечает `{"type": "pong", "timestamp": ...}`.
  - Каждые `settings.WEBSOCKET_BROADCAST_INTERVAL` секунд рассылается:
    ```json
    {
      "type": "drones_status",
      "timestamp": "2025-11-11T09:15:00.000000",
      "drones": [
        {
          "drone_id": "b84a0fda-60c2-4af7-8e81-dca9a2681a07",
          "status": "in_flight",
          "battery_level": 62.5,
          "position": { "latitude": 56.83801, "longitude": 60.59747, "altitude": 120.4 },
          "speed": 8.1,
          "current_delivery_id": "8f264378-6a59-4e6f-9ada-1955f9cce0b1",
          "error_message": null,
          "last_updated": "2025-11-11T09:14:58.000000"
        }
      ]
    }
    ```

### Канал видеопрокси

- **URL**: `wss://{drone-service-host}/ws/drone/{drone_id}/video`
- **Назначение**: поток видеокадров от дрона к админ-клиентам. Формат полезной нагрузки зависит от обработчика видеопрокси и рассматривается как бинарный/JSON.

---

## Справочник моделей данных

| Сущность | Поля |
| --- | --- |
| `User` | `id`, `full_name`, `email?`, `phone_number?`, `phone_verified`, `created_at`, `role`, `qr_issued_at?`, `qr_expires_at?` |
| `Good` | `id`, `name`, `weight`, `height`, `length`, `width`, `quantity_available` |
| `Order` | `id`, `user_id`, `good_id`, `parcel_automat_id`, `locker_cell_id?`, `status`, `created_at` |
| `Delivery` | `id`, `orderID`, `droneID`, `parcelAutomatID`, `status` |
| `ParcelAutomat` | `id`, `ip_address`, `city`, `address`, `number_of_cells`, `coordinates`, `aruco_id`, `is_working` |
| `LockerCell` | `id`, `post_id`, `height`, `length`, `width`, `status` |
| `Drone` | `id`, `model`, `ip_address`, `status` |
| `DroneStatus` (виртуальная) | `drone_id`, `status`, `battery_level`, `position`, `current_delivery_id?`, `error_message?` |
| `SystemStatus` | `drones`, `automats`, `active_deliveries` |

---

## Статусные перечисления

- **Delivery Status**: `pending`, `in_progress`, `in_transit`, `completed`, `failed`.
- **Drone Status**: `idle`, `busy`, `in_flight`, `returning`, `error` (возможны кастомные значения из телеметрии).
- **Locker Cell Status**: `available`, `reserved`, `occupied`, `maintenance`.
- **Parcel Automat**: булево поле `is_working`.

---

## Управление изменениями

- Версионирование зашито в путь (`/api/v1`). Для несовместимых изменений — `/api/v2`.
- Обновляйте документ при добавлении endpoints, статусов или изменении payload.
- Swagger генерируется из `backend/go-orchestrator/docs/swagger.yaml` (`make swag`).
- Согласовывайте изменения WebSocket/gRPC с командами мобильного приложения, админ-панели и прошивки устройств.
