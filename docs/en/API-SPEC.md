# hiTech Full API & Protocol Specification

The hiTech platform exposes multiple interfaces—REST over HTTP, WebSocket channels, and gRPC services—to orchestrate autonomous deliveries between drones, parcel automats, and client applications. This document captures **every request and response shape** in detail so integrators can implement clients without needing to inspect source code.

---

## Table of Contents
1. [Conventions](#conventions)
2. [Authentication & Session Lifecycle](#authentication--session-lifecycle)
3. [Common Error Model](#common-error-model)
4. [HTTP API (Orchestrator, Admin, Mobile)](#http-api-orchestrator-admin-mobile)
   - [Users & Authentication](#users--authentication)
   - [Notification Devices](#notification-devices)
   - [Goods Catalog](#goods-catalog)
   - [Parcel Automats](#parcel-automats)
   - [Locker Cells](#locker-cells)
   - [Orders](#orders)
   - [Deliveries](#deliveries)
   - [Drones](#drones)
   - [QR Operations](#qr-operations)
   - [System Monitoring](#system-monitoring)
   - [Utility Endpoints](#utility-endpoints)
5. [Drone Service HTTP API](#drone-service-http-api)
6. [gRPC API](#grpc-api)
7. [WebSocket APIs](#websocket-apis)
   - [Drone ↔ Drone Service](#drone--drone-service-websocket)
   - [Admin Panel ↔ Drone Service](#admin-panel--drone-service-websocket)
   - [Video Proxy Channel](#video-proxy-channel)
8. [Data Model Reference](#data-model-reference)
9. [Status Enums](#status-enums)
10. [Change Management](#change-management)

---

## Conventions
- **Base URLs**
  - REST (orchestrator): `https://{env-host}/api/v1`
  - WebSocket (drone service): `wss://{drone-service-host}/ws/...`
  - gRPC: `grpc://{orchestrator-host}:{GO_GRPC_PORT}`
- **Content Types**: Requests with bodies must use `Content-Type: application/json`. Responses default to JSON unless noted.
- **Identifiers**: UUID v4 strings (e.g., `4e2a6a94-9c9d-4c54-a9fb-6b0f8dd695b5`) unless otherwise specified.
- **Timestamps**: ISO-8601 in UTC (e.g., `2025-11-11T10:15:30Z`) for responses; requests accept ISO-8601 or RFC3339 when allowed.
- **Numbers**: Use IEEE-754 doubles for decimal payloads (Go `float64`). Weight/size units are in **meters** for dimensions and **kilograms** for weight.
- **Localization**: All server messages returned by the orchestrator are English strings. SMS content depends on SMS Aero templates.
- **Rate Limiting**: Selected auth endpoints have per-IP rate limits; the API returns `429 Too Many Requests` with the standard error payload.

---

## Authentication & Session Lifecycle

### Token Model
- **Access Token**: JWT signed with `HS256`, default TTL 7 days. Required in the `Authorization: Bearer <access_token>` header for protected endpoints.
- **Refresh Token**: JWT signed with `HS256`, default TTL 14 days. Always sent in the `Authorization` header when calling `/auth/refresh`.
- **Roles**: `user`, `admin`. Role is embedded in the token and enforces server-side authorization.
- **Phone Verification**: Registration and login flows rely on SMS codes via SMS Aero.

### Sequence Diagrams

#### Registration (Email + Phone)
1. Client sends `POST /api/v1/auth/register` with profile data.
2. Server persists user with `phone_verified=false`, sends SMS code.
3. Client submits `POST /api/v1/auth/verify/phone` with code to receive tokens and QR payload.

#### Email Login
1. Client sends `POST /api/v1/auth/login`.
2. Server validates credentials, returns access & refresh tokens, QR metadata.

#### Phone Login
1. Client requests `POST /api/v1/auth/login/phone` with `phone`.
2. SMS is sent. Client reuses `POST /api/v1/auth/verify/phone` to exchange code for tokens.

#### Password Reset
1. Initiate with `POST /api/v1/auth/password/reset/request`.
2. Complete with `POST /api/v1/auth/password/reset`.

#### Token Refresh
1. Client calls `POST /api/v1/auth/refresh` with `Authorization: Bearer <refresh_token>`.
2. Server responds with new token pair.

### Required Headers
| Header | Value | Notes |
| --- | --- | --- |
| `Authorization` | `Bearer <access_token>` | Mandatory for protected REST endpoints. |
| `Content-Type` | `application/json` | For requests with JSON bodies. |
| `Accept` | `application/json` | Recommended. |

---

## Common Error Model
- All non-2xx responses (except empty 204) use:
  ```json
  {
    "error": "human-readable message"
  }
  ```
- Validation errors bubble up from Gin binding and include the offending field or constraint (e.g., `"error": "Key: 'CreateGoodRequest.Weight' Error:Field validation for 'Weight' failed on the 'gt' tag"`).
- Typical status codes:
  - `400 Bad Request`: Malformed input, validation error, or business rule violation.
  - `401 Unauthorized`: Missing/invalid token, invalid SMS/QR code.
  - `403 Forbidden`: Authenticated user lacks permission (e.g., registering a device for another user).
  - `404 Not Found`: Resource not present.
  - `409 Conflict`: Not currently emitted, but reserved for future duplicate handling.
  - `500 Internal Server Error`: Unhandled server-side exception.

---

## HTTP API (Orchestrator, Admin, Mobile)

**Base URL**: `https://{host}/api/v1`  
Protected endpoints require a valid access token unless otherwise stated.

### Users & Authentication

#### `POST /auth/register`
Registers a new user and triggers an SMS verification code.

- **Auth**: None.
- **Rate limit**: Enabled (per IP).
- **Request Body**

| Field | Type | Required | Constraints | Description |
| --- | --- | --- | --- | --- |
| `full_name` | string | Yes | 1..255 chars | Full name of the user. |
| `email` | string | Yes | valid email | Login email. Must be unique. |
| `phone` | string | Yes | digits, country code recommended | Mobile number for SMS login. |
| `password` | string | Yes | ≥ 6 chars | User password (stored hashed). |

- **Example Request**
```json
{
  "full_name": "Alice Robotics",
  "email": "alice@example.com",
  "phone": "79001234567",
  "password": "Sup3rSecure!"
}
```

- **Responses**
  - `201 Created`
    ```json
    {
      "message": "Registration successful. SMS code sent to 79001234567"
    }
    ```
  - `400 Bad Request` – validation failure.
  - `500 Internal Server Error` – persistence/SMS failure.

#### `POST /auth/verify/phone`
Confirms the SMS verification code and issues JWT + QR payload.

- **Auth**: None.
- **Request Body**

| Field | Type | Required | Constraints | Description |
| --- | --- | --- | --- | --- |
| `phone` | string | Yes | Must match registered phone | Phone number used during registration/login. |
| `code` | string | Yes | Exactly 4 digits | SMS code delivered to the user. |

- **Example Request**
```json
{
  "phone": "79001234567",
  "code": "5842"
}
```

- **Success 200 Response**
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

- **Error Codes**
  - `400`: Invalid payload or unverified phone.
  - `401`: Wrong or expired SMS code.
  - `500`: Token generation failure.

#### `POST /auth/login`
Email/password authentication.

- **Request Body**

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `email` | string | Yes | Registered email. |
| `password` | string | Yes | Plain password, min 6 chars. |

- **Responses**: Same schema as `/auth/verify/phone`.
- **Errors**: `400` (validation), `401` (invalid credentials), `500`.

#### `POST /auth/login/phone`
Sends a login SMS code to the supplied phone number.

- **Request Body**
  ```json
  { "phone": "79001234567" }
  ```
- **Success 200**
  ```json
  { "message": "Login code sent to your phone" }
  ```
- **Errors**: `400` invalid number or phone not verified, `404` user missing, `500` SMS failure.

#### `POST /auth/password/reset/request`
Triggers an SMS with a reset code.

- **Request Body**
  ```json
  { "phone": "79001234567" }
  ```
- **Success 200**: same as login phone message.
- **Errors**: `400`/`404`/`500`.

#### `POST /auth/password/reset`
Applies a new password after SMS verification.

- **Request Body**

| Field | Type | Required | Constraints |
| --- | --- | --- | --- |
| `phone` | string | Yes | Must match existing user. |
| `code` | string | Yes | 4-digit SMS code. |
| `new_password` | string | Yes | ≥ 6 chars. |

- **Success 200**
  ```json
  { "message": "Password reset successfully" }
  ```
- **Errors**: `400` invalid code, `500`.

#### `POST /auth/refresh`
Exchanges a refresh token for a new token pair.

- **Headers**: `Authorization: Bearer <refresh_token>`
- **Success 200**
  ```json
  {
    "access_token": "new-access",
    "refresh_token": "new-refresh",
    "expires_at": 1762851164
  }
  ```
- **Errors**: `401` invalid/expired refresh token, `500`.

#### `GET /auth/me`
Returns the authenticated user profile plus QR code.

- **Headers**: access token required.
- **Success 200**: identical to `POST /auth/verify/phone` payload.
- **Errors**: `401`, `404` if user record missing.

### Notification Devices

#### `POST /users/{id}/devices`
Registers an FCM/APNS device token for push notifications.

- **Auth**: Bearer; user can only register their own device.
- **Path Parameters**

| Name | Type | Description |
| --- | --- | --- |
| `id` | UUID | User ID that owns the device token. |

- **Request Body**

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `token` | string | Yes | Push token issued by FCM/APNS. |
| `platform` | string | Yes | Platform identifier (e.g., `android`, `ios`). |

- **Example Request**
```json
{
  "token": "fcm:dyk3Nz8...",
  "platform": "android"
}
```

- **Responses**
  - `200 OK`
    ```json
    { "message": "Device registered" }
    ```
    Returns `"Notifications are disabled"` when Firebase credentials are absent on the server.
  - `400` validation error.
  - `401` user not authenticated.
  - `403` attempting to register for another user.
  - `500` push service failure.

### Goods Catalog

Base path: `/goods` (protected).

#### `GET /goods`
- **Summary**: List all goods with current stock.
- **Response 200**
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
- **Errors**: `500`.

#### `POST /goods`
Creates **N** identical goods at once (quantity field).

- **Request Body**

| Field | Type | Required | Constraints | Description |
| --- | --- | --- | --- | --- |
| `name` | string | Yes | 1..255 chars | Product name. |
| `weight` | number | Yes | `> 0` | Weight in kilograms. |
| `height` | number | Yes | `> 0` | Height in meters. |
| `length` | number | Yes | `> 0` | Length in meters. |
| `width` | number | Yes | `> 0` | Width in meters. |
| `quantity` | integer | Yes | `> 0` | Number of units to create. |

- **Example Request**
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

- **Success 201**: Array of created goods (same structure as GET).
- **Errors**: `400`, `500`.

#### `GET /goods/{id}`
Returns a single good by UUID.

- **Errors**: `400` invalid UUID, `404` missing good.

#### `PATCH /goods/{id}`
Updates dimensions and name (quantity is immutable).

- **Request Body**: same schema as create but without `quantity`.
- **Success 200**: updated good.
- **Errors**: `400`, `500`.

#### `DELETE /goods/{id}`
Deletes the given good (no body on success).
- **Success 200**
  ```json
  { "message": "Good deleted successfully" }
  ```
- **Errors**: `400`, `500`.

### Parcel Automats

Protected administrative endpoints require authentication unless noted. Public endpoints (`/automats/qr-scan`, `/automats/confirm-pickup`) are unauthenticated to allow on-premise devices to interact.

#### `POST /automats`
Creates a parcel automat with cell dimensions.

- **Request Body**

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `city` | string | Yes | City name. |
| `address` | string | Yes | Street address. |
| `ip_address` | string | No | Optional static IP on LAN. |
| `coordinates` | string | No | Arbitrary coordinate string (e.g., `"55.751244,37.618423"`). |
| `aruco_id` | integer | Yes | Marker ID for drone landing recognition. |
| `number_of_cells` | integer | Yes | Must equal length of `cells`. |
| `cells` | array | Yes | Array of dimension objects. |

Cell element:

| Field | Type | Required | Constraints |
| --- | --- | --- | --- |
| `height` | number | Yes | `> 0` |
| `length` | number | Yes | `> 0` |
| `width` | number | Yes | `> 0` |

- **Example Request**
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

- **Success 201**
```json
{
  "id": "a941aa86-b0ab-4f75-a7b9-4dec56dde16f",
  "ip_address": "192.168.10.15",
  "city": "Ekaterinburg",
  "address": "Industrial Zone, Building 5",
  "number_of_cells": 2,
  "coordinates": "56.8380,60.5975",
  "aruco_id": 42,
  "is_working": true
}
```

#### `GET /automats`
Lists all automats (array of `ParcelAutomat`).

#### `GET /automats/working`
Filters automats with `is_working=true`.

#### `GET /automats/{id}`
Fetches one automat by UUID.

#### `PUT /automats/{id}`
Updates metadata:

| Field | Type | Required |
| --- | --- | --- |
| `city` | string | Yes |
| `address` | string | Yes |
| `ip_address` | string | No |
| `coordinates` | string | No |

#### `GET /automats/{id}/cells`
Returns an array of locker cells belonging to the automat.

#### `PATCH /automats/{id}/cells/{cellId}`
Updates 3D dimensions of a cell.

- **Request Body**: same as `UpdateCellRequest`.
- **Response 200**
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
Toggles availability.

- **Request Body**
  ```json
  { "is_working": true }
  ```
- **Response 200**: updated automat.

#### `DELETE /automats/{id}`
Deletes an automat.

- **Success 200**
  ```json
  { "message": "Parcel automat deleted successfully" }
  ```

#### `POST /automats/qr-scan` *(Public)*
Used by automat controllers to decode a customer QR.

- **Request Body**

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `qr_data` | string | Yes | Raw QR payload scanned from customer device. |
| `parcel_automat_id` | string | Yes | UUID of the automat performing the scan. |

- **Success 200**
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
- **Errors**: `400` invalid QR/automat ID.

#### `POST /automats/confirm-pickup` *(Public)*
Confirms that given cells have been emptied by a customer.

- **Request Body**
  ```json
  {
    "cell_ids": [
      "c7b84618-6242-49eb-87e4-604b0e54c7a5",
      "868f662d-2940-477a-9de7-8232d6ad91aa"
    ]
  }
  ```
- **Success 200**
  ```json
  {
    "success": true,
    "message": "Pickup confirmed successfully"
  }
  ```
- **Errors**: `400` invalid payload, `500`.

### Locker Cells

#### `GET /locker/cells/{id}`
- **Auth**: Bearer.
- **Response 200**
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
- **Errors**: `400`, `404`.

### Orders

#### `POST /orders`
Creates an order for the authenticated user.

- **Request Body**
  ```json
  { "good_id": "4a9b65f2-4f8a-4f3d-9d52-9aca43b0c8a1" }
  ```
- **Success 201**
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
- **Errors**: `400`, `401`, `500`.

#### `POST /orders/batch`
Creates multiple orders in one request (each for a different good).

- **Request Body**
  ```json
  {
    "good_ids": [
      "4a9b65f2-4f8a-4f3d-9d52-9aca43b0c8a1",
      "71e3ba1d-8e06-4127-996f-afdb238be6da"
    ]
  }
  ```
- **Success 201**: array of orders.
- **Errors**: same as single create.

#### `GET /orders/{id}`
Fetches order by UUID. Errors: `400`, `404`.

#### `POST /orders/{id}/return`
Cancels an existing order and rolls back allocated resources when the owner aborts delivery.

- **Auth**: Bearer.
- **Availability**: only while order status is `pending` or `in_progress`.
- **Side effects**:
  - The reserved locker cell switches back to `available`.
  - Item stock is replenished in inventory.
  - A high-priority `delivery.return` task is published so the assigned drone returns to base marker `131`.
- **Request Body**: _empty_.
- **Success 200**
  ```json
  { "success": true }
  ```
- **Errors**: `400` invalid UUID or order not returnable, `401`, `403` (order belongs to another user), `404` order not found, `500`.

#### `GET /orders/user/{userId}`
Returns orders for the specified user, including attached good metadata.

- **Response 200**
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

### Deliveries

#### `GET /deliveries/{id}`
Returns delivery metadata:
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
Updates delivery status.

- **Request Body**
  ```json
  { "status": "completed" }
  ```
- **Success 200**
  ```json
  { "success": true }
  ```
- **Errors**: `400`, `500`.

#### `POST /deliveries/confirm-loaded`
Confirms drone loading from a locker cell.

- **Request Body**
  ```json
  {
    "order_id": "2f611f70-0264-4f1f-a5f8-4a07b8942a60",
    "locker_cell_id": "c7b84618-6242-49eb-87e4-604b0e54c7a5",
    "timestamp": "2025-11-11T09:12:05Z"
  }
  ```
- **Success 200**
  ```json
  {
    "success": true,
    "message": "Goods loaded confirmed successfully"
  }
  ```
- **Errors**: `400` invalid UUID, `500`.

### Drones

#### `GET /drones`
List all drones (array of `Drone` entities).

#### `POST /drones`
- **Request Body**
  ```json
  {
    "model": "DJI Matrice 30",
    "ip_address": "192.168.20.11"
  }
  ```
- **Success 201**
  ```json
  {
    "id": "b84a0fda-60c2-4af7-8e81-dca9a2681a07",
    "model": "DJI Matrice 30",
    "ip_address": "192.168.20.11",
    "status": "idle"
  }
  ```

#### `GET /drones/{id}`
Returns drone metadata.

#### `PUT /drones/{id}`
Updates model/IP.

- **Request Body**
  ```json
  { "model": "DJI Matrice 30T", "ip_address": "192.168.20.20" }
  ```

#### `PATCH /drones/{id}/status`
Changes operational status.

- **Request Body**
  ```json
  { "status": "busy" }
  ```
- **Success 200**: updated drone record.

#### `GET /drones/{id}/status`
Returns live status information aggregated from the drone service.

- **Sample Response**
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
Removes a drone registration.

- **Success 200**
  ```json
  { "message": "Drone deleted successfully" }
  ```

### QR Operations

#### `POST /qr/validate`
- **Auth**: None (called by parcel automat services).
- **Request Body**
  ```json
  { "qr_data": "base64-encoded-payload" }
  ```
- **Success 200**
  ```json
  {
    "valid": true,
    "user_id": "0f28f2d9-1e61-4f8f-8d65-b8efcd2c3a8c",
    "email": "alice@example.com",
    "name": "Alice Robotics",
    "expires_at": 1762851164
  }
  ```
- **Errors**: `400` invalid payload, `401` expired/unknown QR.

#### `POST /qr/refresh`
- **Auth**: Bearer.
- **Success 200**
  ```json
  {
    "qr_code": "iVBORw0KGgoAAAANSUhEUgAA...",
    "expires_at": 1762851164
  }
  ```
- **Errors**: `401`, `500`.

### System Monitoring

#### `GET /monitoring/system-status`
Aggregates drones, parcel automats (with cell lists), and active deliveries.

- **Sample Response**
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

### Utility Endpoints
- `GET /metrics`: Exposes Prometheus metrics (no auth). Response is plaintext exposition format.
- `GET /swagger/*`: Swagger UI (served by `github.com/swaggo/gin-swagger`).

---

## Drone Service HTTP API

Base URL: `https://{drone-service-host}` (or behind orchestrator networking). All endpoints served by FastAPI in `backend/drone-service`.

### Health & Metrics
- `GET /health` → `{"status": "healthy", "service": "drone-service"}`
- `GET /status` → Returns the first drone state or fallback object if none connected.
- `GET /metrics` → Prometheus metrics (plaintext).

### Command Endpoint

#### `POST /api/drones/{drone_id}/command`
Sends a control command to a connected drone via WebSocket.

- **Auth**: Currently none (should be protected via network-level ACL).
- **Path Parameters**

| Name | Type | Description |
| --- | --- | --- |
| `drone_id` | string | UUID string assigned by orchestrator. |

- **Request Body**
  ```json
  { "command": "return_home" }
  ```
  Supported commands: currently only `"return_home"`.

- **Responses**
  - `200 OK`
    ```json
    { "success": true, "message": "Return home command sent" }
    ```
  - `404` if drone is not connected (`{"success": false, "message": "Drone not connected"}`).
  - `400` unknown command.

---

## gRPC API

### Endpoint
- Service: `pb.OrchestratorService`
- Default address: `:{GO_GRPC_PORT}` (50052 by default).
- Transport security: depends on deployment (staging/prod behind TLS-terminating proxy).

### RPCs

#### `RequestCellOpen`
Used by the parcel automat controller when a drone arrives and needs to open a cell.

- **Request Message**
  ```proto
  message CellOpenRequest {
    string order_id = 1;           // UUID string of the order
    string parcel_automat_id = 2;  // UUID string of the automat
  }
  ```
- **Response Message**
  ```proto
  message CellOpenResponse {
    bool success = 1;
    string message = 2;  // descriptive status
    string cell_id = 3;  // UUID string of opened cell when success=true
  }
  ```

- **Example `grpcurl` Call**
  ```bash
  grpcurl -plaintext \
    -d '{"order_id":"2f611f70-0264-4f1f-a5f8-4a07b8942a60","parcel_automat_id":"a941aa86-b0ab-4f75-a7b9-4dec56dde16f"}' \
    orchestrator:50052 pb.OrchestratorService/RequestCellOpen
  ```
- **Success Response**
  ```json
  {
    "success": true,
    "message": "Cell opened",
    "cellId": "c7b84618-6242-49eb-87e4-604b0e54c7a5"
  }
  ```
- **Failure Example**
  ```json
  {
    "success": false,
    "message": "No free cells",
    "cellId": ""
  }
  ```

---

## WebSocket APIs

### Drone ↔ Drone Service WebSocket

- **URL**: `wss://{drone-service-host}/ws/drone`
- **Purpose**: Bi-directional telemetry and command channel between the physical drone and drone-service backend.
- **Handshake**
  1. Client opens WS connection and immediately sends a registration message:
     ```json
     {
       "type": "register",
       "ip_address": "192.168.20.11"
     }
     ```
  2. Server validates IP by looking up the drone in the orchestrator DB and replies:
     ```json
     {
       "type": "registered",
       "drone_id": "b84a0fda-60c2-4af7-8e81-dca9a2681a07",
       "timestamp": "2025-11-11T09:14:20.123456"
     }
     ```
  3. Subsequent messages can be sent in either direction.

- **Inbound Message Types (from drone to server)**

| Type | Payload | Description |
| --- | --- | --- |
| `heartbeat` | `{ "battery_level": float, "position": {"latitude": float, "longitude": float, "altitude": float}, "status": "idle|in_flight|returning", "speed": float, "current_delivery_id": string?, "error_message": string? }` | Periodic health signal; persists battery and location. |
| `status_update` | Same fields as heartbeat. | Arbitrary state update (deprecated in favor of heartbeat). |
| `delivery_update` | `{ "delivery_id": string, "drone_status": "arrived_at_locker|returning|<other>" }` | Updates orchestrator about delivery progress. |
| `arrived_at_destination` | `{ "order_id": string, "parcel_automat_id": string }` | Signals arrival; orchestrator will attempt to open the cell via gRPC. |
| `cargo_dropped` | `{ "order_id": string, "locker_cell_id": string }` | Signals that the payload has been released. |

- **Outbound Message Types (from server to drone)**

| Type | Payload | Description |
| --- | --- | --- |
| `delivery_task` | Arbitrary task object (order route, coordinates) | Sent when the orchestrator assigns a delivery; exact shape depends on use case. |
| `command` | e.g., `{ "command": "return_home" }` | Real-time command forwarded from REST `/api/drones/{id}/command`. |
| `error` | `{ "message": string }` | Error notification (e.g., invalid registration). |

- **Disconnection Handling**: On disconnect the server unregisters the drone and clears connected state.

### Admin Panel ↔ Drone Service WebSocket

- **URL**: `wss://{drone-service-host}/ws/admin`
- **Purpose**: Admin dashboard receives live drone telemetry.
- **Message Flow**
  - Client may send `{"type": "ping"}`; server responds with `{"type": "pong", "timestamp": ...}`.
  - Every `settings.WEBSOCKET_BROADCAST_INTERVAL` seconds (configured via env) the server broadcasts:
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

### Video Proxy Channel

- **URL**: `wss://{drone-service-host}/ws/drone/{drone_id}/video`
- **Purpose**: Streams real-time video frames from drone to admin clients. Payload format depends on the external video proxy handler and is currently treated as opaque binary/JSON frames. Implementations should follow the same registration handshake as the drone channel.

---

## Data Model Reference

| Entity | Fields |
| --- | --- |
| `User` | `id`, `full_name`, `email?`, `phone_number?`, `phone_verified`, `created_at`, `role`, `qr_issued_at?`, `qr_expires_at?` |
| `Good` | `id`, `name`, `weight`, `height`, `length`, `width`, `quantity_available` |
| `Order` | `id`, `user_id`, `good_id`, `parcel_automat_id`, `locker_cell_id?`, `status`, `created_at` |
| `Delivery` | `id`, `orderID`, `droneID?`, `parcelAutomatID`, `status` |
| `ParcelAutomat` | `id`, `ip_address`, `city`, `address`, `number_of_cells`, `coordinates`, `aruco_id`, `is_working` |
| `LockerCell` | `id`, `post_id`, `height`, `length`, `width`, `status` |
| `Drone` | `id`, `model`, `ip_address`, `status` |
| `DroneStatus` (virtual) | `drone_id`, `status`, `battery_level`, `position{latitude,longitude,altitude}`, `current_delivery_id?`, `error_message?` |
| `SystemStatus` | `drones` (array of `Drone`), `automats` (array of automat+cells), `active_deliveries` (array of `Delivery`) |

---

## Status Enums

- **Delivery Status**: `pending`, `in_progress`, `in_transit` (internal), `completed`, `failed`.
- **Drone Status**: `idle`, `busy`, `in_flight`, `returning`, `error`. Custom statuses may appear based on telemetry.
- **Locker Cell Status**: `available`, `reserved`, `occupied`, `maintenance`.
- **Parcel Automat**: `is_working` boolean toggled via `/automats/{id}/status`.

---

## Change Management

- Versioning is encoded in the path (`/api/v1`). Backwards-incompatible changes must introduce `/api/v2`.
- Update this document alongside any new endpoints, status values, or payload adjustments.
- Refer to `backend/go-orchestrator/docs/swagger.yaml` for auto-generated Swagger; regenerate with `make swag` when controllers change.
- Coordinate with mobile, admin panel, and device firmware teams when adding/removing message types on WebSockets or altering gRPC contracts.

