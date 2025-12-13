# API Specification

## Overview

SkyPost Delivery provides multiple API interfaces for different components:
- **REST API** (Go-Orchestrator): Main HTTP API for clients and admin panel
- **gRPC API** (Go-Orchestrator): Inter-service communication
- **WebSocket API** (Drone-Service): Real-time drone and admin communication
- **HTTP API** (Locker-Agent): Parcel automat cell control

**Base URL (Production)**: `https://api.skypost.delivery`  
**Base URL (Development)**: `http://localhost:8080`

**API Version**: v1  
**Protocol**: HTTP/1.1, WebSocket (RFC 6455), gRPC (HTTP/2)  
**Content-Type**: `application/json`  
**Character Encoding**: UTF-8

## Authentication

### JWT Bearer Token

Most endpoints require authentication using JWT (JSON Web Tokens).

**Header Format**:
```http
Authorization: Bearer <access_token>
```

**Token Structure**:
```json
{
  "sub": "user_uuid",
  "email": "user@example.com",
  "full_name": "John Doe",
  "role": "client|admin",
  "exp": 1705320000,
  "iat": 1705233600
}
```

**Token Expiration**:
- Access Token: 24 hours
- Refresh Token: 7 days

**Token Refresh Flow**:
1. Access token expires after 24 hours
2. Client sends refresh token to `/api/v1/auth/refresh`
3. Server validates refresh token
4. Server issues new access and refresh token pair

### Rate Limiting

**Authentication Endpoints**: 10 requests per minute per IP  
**Order Creation**: 20 requests per minute per user  
**QR Validation**: 30 requests per minute per IP  
**General Endpoints**: 100 requests per minute per IP

**Rate Limit Headers**:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705233660
```

**429 Too Many Requests Response**:
```json
{
  "error": "Rate limit exceeded. Try again in 60 seconds."
}
```

## Common Response Formats

### Success Response

```json
{
  "success": true,
  "data": { ... }
}
```

### Error Response

```json
{
  "error": "Error message description"
}
```

### HTTP Status Codes

| Code | Description | Usage |
|------|-------------|-------|
| 200 | OK | Successful GET, PUT, PATCH request |
| 201 | Created | Successful POST request creating resource |
| 204 | No Content | Successful DELETE request |
| 400 | Bad Request | Invalid request format or parameters |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Authenticated but insufficient permissions |
| 404 | Not Found | Resource does not exist |
| 409 | Conflict | Resource conflict (duplicate entry) |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server-side error |
| 503 | Service Unavailable | Service temporarily down |

## REST API Endpoints (Go-Orchestrator)

Base Path: `/api/v1`

### Authentication & Users

#### POST /api/v1/auth/register

Register a new user account and send SMS verification code.

**Request Body**:
```json
{
  "full_name": "John Doe",
  "email": "john.doe@example.com",
  "phone": "+79991234567",
  "password": "SecurePass123!"
}
```

**Validation Rules**:
- `full_name`: Required, 2-255 characters
- `email`: Optional, valid email format
- `phone`: Required, E.164 format (+country code)
- `password`: Required, minimum 8 characters, at least one uppercase, one lowercase, one digit

**Response** (201 Created):
```json
{
  "message": "Registration successful. SMS code sent to +79991234567"
}
```

**Errors**:
- 400: Validation error (invalid format)
- 409: Email or phone already registered
- 500: SMS service failure

**Rate Limit**: 10 requests/minute per IP

---

#### POST /api/v1/auth/verify/phone

Verify phone number with SMS code and activate account.

**Request Body**:
```json
{
  "phone": "+79991234567",
  "code": "123456"
}
```

**Validation Rules**:
- `phone`: Required, E.164 format
- `code`: Required, 6 digits

**Response** (200 OK):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "full_name": "John Doe",
    "email": "john.doe@example.com",
    "phone_number": "+79991234567",
    "phone_verified": true,
    "role": "client",
    "created_at": "2024-01-15T10:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "access_expires_at": 1705320000,
  "refresh_expires_at": 1705838400,
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANS..."
}
```

**Response Fields**:
- `user`: User object with account details
- `access_token`: JWT token for API authentication (24h)
- `refresh_token`: Token for refreshing access token (7d)
- `access_expires_at`: Unix timestamp when access token expires
- `refresh_expires_at`: Unix timestamp when refresh token expires
- `qr_code`: Base64-encoded PNG QR code for parcel pickup

**Errors**:
- 400: Invalid phone or code format
- 401: Invalid or expired verification code
- 404: Phone number not registered
- 500: Database or token generation error

**Rate Limit**: 10 requests/minute per IP

---

#### POST /api/v1/auth/login

Authenticate user with email and password.

**Request Body**:
```json
{
  "email": "john.doe@example.com",
  "password": "SecurePass123!"
}
```

**Response** (200 OK):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "full_name": "John Doe",
    "email": "john.doe@example.com",
    "phone_number": "+79991234567",
    "phone_verified": true,
    "role": "client",
    "created_at": "2024-01-15T10:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "access_expires_at": 1705320000,
  "refresh_expires_at": 1705838400,
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANS..."
}
```

**Errors**:
- 400: Invalid request format
- 401: Invalid email or password
- 500: Server error

**Rate Limit**: 10 requests/minute per IP

---

#### POST /api/v1/auth/login/phone

Request SMS code for phone-based login.

**Request Body**:
```json
{
  "phone": "+79991234567"
}
```

**Response** (200 OK):
```json
{
  "message": "Login code sent to your phone"
}
```

**Errors**:
- 400: Phone not verified (need to complete registration first)
- 404: User with phone not found
- 500: SMS service failure

**Flow**:
1. User submits phone number
2. Server generates 6-digit code, stores with 5-minute expiration
3. SMS sent via SMSAero API
4. User verifies with `/auth/verify/phone` endpoint

**Rate Limit**: 10 requests/minute per IP

---

#### POST /api/v1/auth/password/reset/request

Request password reset via SMS code.

**Request Body**:
```json
{
  "phone": "+79991234567"
}
```

**Response** (200 OK):
```json
{
  "message": "Password reset code sent to your phone"
}
```

**Errors**:
- 400: Phone not verified
- 404: User not found
- 500: SMS service failure

**Rate Limit**: 10 requests/minute per IP

---

#### POST /api/v1/auth/password/reset

Reset password with SMS verification code.

**Request Body**:
```json
{
  "phone": "+79991234567",
  "code": "123456",
  "new_password": "NewSecurePass456!"
}
```

**Validation Rules**:
- `new_password`: Minimum 8 characters, at least one uppercase, one lowercase, one digit

**Response** (200 OK):
```json
{
  "message": "Password reset successfully"
}
```

**Errors**:
- 400: Invalid code or password format
- 401: Expired or invalid verification code
- 500: Database error

**Rate Limit**: 10 requests/minute per IP

---

#### POST /api/v1/auth/refresh

Refresh access token using refresh token.

**Request Headers**:
```http
Authorization: Bearer <refresh_token>
```

**Response** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "access_expires_at": 1705320000,
  "refresh_expires_at": 1705838400
}
```

**Errors**:
- 401: Invalid or expired refresh token

**Rate Limit**: 100 requests/minute per user

---

#### GET /api/v1/auth/me

Get current authenticated user information.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "full_name": "John Doe",
    "email": "john.doe@example.com",
    "phone_number": "+79991234567",
    "phone_verified": true,
    "role": "client",
    "created_at": "2024-01-15T10:00:00Z",
    "qr_issued_at": "2024-01-15T12:00:00Z",
    "qr_expires_at": "2024-01-16T12:00:00Z"
  },
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANS..."
}
```

**Errors**:
- 401: Unauthorized (invalid or missing token)
- 404: User not found

**Rate Limit**: 100 requests/minute per user

---

#### POST /api/v1/users/:id/devices

Register device for push notifications (FCM token).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**URL Parameters**:
- `id`: User UUID (must match authenticated user)

**Request Body**:
```json
{
  "token": "fcm_device_token_here_very_long_string",
  "platform": "android"
}
```

**Validation Rules**:
- `token`: Required, FCM device token
- `platform`: Required, one of: `android`, `ios`, `web`

**Response** (200 OK):
```json
{
  "message": "Device registered"
}
```

**Errors**:
- 400: Invalid request format
- 401: Unauthorized
- 403: User ID mismatch (cannot register device for another user)
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

### Orders

#### POST /api/v1/orders

Create a new order for goods delivery.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Request Body**:
```json
{
  "good_id": "650e8400-e29b-41d4-a716-446655440000"
}
```

**Validation Rules**:
- `good_id`: Required, valid UUID of existing good
- User ID extracted from JWT token

**Response** (201 Created):
```json
{
  "id": "750e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "good_id": "650e8400-e29b-41d4-a716-446655440000",
  "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000",
  "locker_cell_id": null,
  "status": "pending",
  "created_at": "2024-01-15T12:00:00Z"
}
```

**Status Values**:
- `pending`: Order created, waiting for processing
- `processing`: Drone assigned, delivery in progress
- `delivered`: Cargo in locker cell, awaiting pickup
- `completed`: User picked up cargo
- `cancelled`: Order cancelled by user

**Business Logic**:
1. Validate good exists and quantity available > 0
2. Find nearest working parcel automat to user
3. Create order with status `pending`
4. Order worker will process and assign drone

**Errors**:
- 400: Invalid good_id format or good not available
- 401: Unauthorized
- 404: Good not found
- 409: Good out of stock
- 500: Database error

**Rate Limit**: 20 requests/minute per user

---

#### POST /api/v1/orders/batch

Create multiple orders at once.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Request Body**:
```json
{
  "good_ids": [
    "650e8400-e29b-41d4-a716-446655440000",
    "660e8400-e29b-41d4-a716-446655440000",
    "670e8400-e29b-41d4-a716-446655440000"
  ]
}
```

**Response** (201 Created):
```json
[
  {
    "id": "750e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "good_id": "650e8400-e29b-41d4-a716-446655440000",
    "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000",
    "status": "pending",
    "created_at": "2024-01-15T12:00:00Z"
  },
  {
    "id": "760e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "good_id": "660e8400-e29b-41d4-a716-446655440000",
    "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000",
    "status": "pending",
    "created_at": "2024-01-15T12:00:01Z"
  }
]
```

**Errors**:
- 400: Invalid request format or any good unavailable
- 401: Unauthorized
- 500: Database error

**Rate Limit**: 10 requests/minute per user

---

#### GET /api/v1/orders/:id

Get order details by ID.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**URL Parameters**:
- `id`: Order UUID

**Response** (200 OK):
```json
{
  "id": "750e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "good_id": "650e8400-e29b-41d4-a716-446655440000",
  "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000",
  "locker_cell_id": "950e8400-e29b-41d4-a716-446655440000",
  "status": "delivered",
  "created_at": "2024-01-15T12:00:00Z"
}
```

**Errors**:
- 400: Invalid order ID format
- 401: Unauthorized
- 404: Order not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### GET /api/v1/orders/user/:userId

Get all orders for a specific user with goods details.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**URL Parameters**:
- `userId`: User UUID

**Response** (200 OK):
```json
[
  {
    "id": "750e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "good_id": "650e8400-e29b-41d4-a716-446655440000",
    "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000",
    "status": "delivered",
    "created_at": "2024-01-15T12:00:00Z",
    "good": {
      "id": "650e8400-e29b-41d4-a716-446655440000",
      "name": "Small Package",
      "weight": 0.5,
      "height": 10.0,
      "length": 20.0,
      "width": 15.0,
      "quantity_available": 100
    }
  }
]
```

**Errors**:
- 400: Invalid user ID format
- 401: Unauthorized
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### POST /api/v1/orders/:id/return

Cancel order and return drone to base if delivery in progress.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**URL Parameters**:
- `id`: Order UUID

**Response** (200 OK):
```json
{
  "success": true
}
```

**Business Logic**:
1. Verify order belongs to authenticated user
2. Check order status is `processing`
3. Find associated delivery
4. Update order status to `cancelled`
5. Free locker cell if assigned
6. Send "return_to_base" command to drone via WebSocket
7. Update delivery status to `cancelled`

**Errors**:
- 400: Invalid order ID or order cannot be cancelled
- 401: Unauthorized
- 403: Order does not belong to user
- 404: Order not found
- 500: Database or WebSocket communication error

**Rate Limit**: 100 requests/minute per user

---

### Goods

#### GET /api/v1/goods

Get list of all available goods.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
[
  {
    "id": "650e8400-e29b-41d4-a716-446655440000",
    "name": "Small Package",
    "weight": 0.5,
    "height": 10.0,
    "length": 20.0,
    "width": 15.0,
    "quantity_available": 100
  },
  {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "name": "Medium Box",
    "weight": 2.0,
    "height": 20.0,
    "length": 30.0,
    "width": 25.0,
    "quantity_available": 50
  }
]
```

**Errors**:
- 401: Unauthorized
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### GET /api/v1/goods/:id

Get good details by ID.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**URL Parameters**:
- `id`: Good UUID

**Response** (200 OK):
```json
{
  "id": "650e8400-e29b-41d4-a716-446655440000",
  "name": "Small Package",
  "weight": 0.5,
  "height": 10.0,
  "length": 20.0,
  "width": 15.0,
  "quantity_available": 100
}
```

**Errors**:
- 400: Invalid good ID format
- 401: Unauthorized
- 404: Good not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### POST /api/v1/goods

Create new goods (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**Request Body**:
```json
{
  "name": "Large Parcel",
  "weight": 5.0,
  "height": 30.0,
  "length": 40.0,
  "width": 35.0,
  "quantity": 25
}
```

**Validation Rules**:
- `name`: Required, 1-255 characters
- `weight`: Required, positive decimal (kg)
- `height`, `length`, `width`: Required, positive decimals (cm)
- `quantity`: Required, non-negative integer

**Response** (201 Created):
```json
[
  {
    "id": "670e8400-e29b-41d4-a716-446655440001",
    "name": "Large Parcel",
    "weight": 5.0,
    "height": 30.0,
    "length": 40.0,
    "width": 35.0,
    "quantity_available": 25
  }
]
```

**Note**: Returns array because implementation creates goods in batch

**Errors**:
- 400: Validation error
- 401: Unauthorized
- 403: Not admin role
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### PATCH /api/v1/goods/:id

Update good details (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Good UUID

**Request Body** (all fields optional):
```json
{
  "name": "Updated Package Name",
  "weight": 0.6,
  "height": 12.0,
  "length": 22.0,
  "width": 16.0
}
```

**Response** (200 OK):
```json
{
  "id": "650e8400-e29b-41d4-a716-446655440000",
  "name": "Updated Package Name",
  "weight": 0.6,
  "height": 12.0,
  "length": 22.0,
  "width": 16.0,
  "quantity_available": 100
}
```

**Errors**:
- 400: Invalid good ID or validation error
- 401: Unauthorized
- 403: Not admin role
- 404: Good not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### DELETE /api/v1/goods/:id

Delete good (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Good UUID

**Response** (200 OK):
```json
{
  "message": "Good deleted successfully"
}
```

**Errors**:
- 400: Invalid good ID format
- 401: Unauthorized
- 403: Not admin role
- 404: Good not found
- 409: Good has active orders (cannot delete)
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

### Drones

#### GET /api/v1/drones

Get list of all drones.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**Response** (200 OK):
```json
[
  {
    "id": "450e8400-e29b-41d4-a716-446655440000",
    "model": "Clover 4",
    "ip_address": "192.168.1.100",
    "status": "idle",
    "battery_level": 95.5,
    "created_at": "2024-01-10T08:00:00Z",
    "updated_at": "2024-01-15T12:30:00Z"
  },
  {
    "id": "460e8400-e29b-41d4-a716-446655440000",
    "model": "Clover 4",
    "ip_address": "192.168.1.101",
    "status": "busy",
    "battery_level": 72.3,
    "created_at": "2024-01-10T08:00:00Z",
    "updated_at": "2024-01-15T12:35:00Z"
  }
]
```

**Drone Status Values**:
- `idle`: Ready for task assignment
- `busy`: Currently executing delivery
- `charging`: Charging at base station
- `maintenance`: Under maintenance
- `offline`: Not connected to system

**Errors**:
- 401: Unauthorized
- 403: Not admin role
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### GET /api/v1/drones/:id

Get drone details by ID.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Drone UUID

**Response** (200 OK):
```json
{
  "id": "450e8400-e29b-41d4-a716-446655440000",
  "model": "Clover 4",
  "ip_address": "192.168.1.100",
  "status": "idle",
  "battery_level": 95.5,
  "created_at": "2024-01-10T08:00:00Z",
  "updated_at": "2024-01-15T12:30:00Z"
}
```

**Errors**:
- 400: Invalid drone ID format
- 401: Unauthorized
- 403: Not admin role
- 404: Drone not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### GET /api/v1/drones/:id/status

Get real-time drone status including telemetry.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Drone UUID

**Response** (200 OK):
```json
{
  "drone_id": "450e8400-e29b-41d4-a716-446655440000",
  "status": "busy",
  "battery_level": 87.5,
  "position": {
    "latitude": 55.7558,
    "longitude": 37.6173,
    "altitude": 15.5
  },
  "current_delivery_id": "750e8400-e29b-41d4-a716-446655440000",
  "connection_status": "connected",
  "last_heartbeat": "2024-01-15T12:35:45Z"
}
```

**Response Fields**:
- `position`: Current GPS coordinates (null if offline)
- `current_delivery_id`: Active delivery UUID (null if idle)
- `connection_status`: `connected`, `disconnected`
- `last_heartbeat`: Last telemetry timestamp

**Errors**:
- 400: Invalid drone ID format
- 401: Unauthorized
- 403: Not admin role
- 500: Database or cache error

**Rate Limit**: 100 requests/minute per user

---

#### POST /api/v1/drones

Register new drone (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**Request Body**:
```json
{
  "model": "Clover 4",
  "ip_address": "192.168.1.102"
}
```

**Validation Rules**:
- `model`: Required, 1-255 characters
- `ip_address`: Required, valid IPv4 or IPv6 address

**Response** (201 Created):
```json
{
  "id": "470e8400-e29b-41d4-a716-446655440000",
  "model": "Clover 4",
  "ip_address": "192.168.1.102",
  "status": "idle",
  "battery_level": 100.0,
  "created_at": "2024-01-15T13:00:00Z",
  "updated_at": "2024-01-15T13:00:00Z"
}
```

**Errors**:
- 400: Validation error or invalid IP
- 401: Unauthorized
- 403: Not admin role
- 409: Drone with IP already exists
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### PUT /api/v1/drones/:id

Update drone information (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Drone UUID

**Request Body**:
```json
{
  "model": "Clover 4 Pro",
  "ip_address": "192.168.1.105"
}
```

**Response** (200 OK):
```json
{
  "id": "450e8400-e29b-41d4-a716-446655440000",
  "model": "Clover 4 Pro",
  "ip_address": "192.168.1.105",
  "status": "idle",
  "battery_level": 95.5,
  "created_at": "2024-01-10T08:00:00Z",
  "updated_at": "2024-01-15T13:05:00Z"
}
```

**Errors**:
- 400: Invalid drone ID or validation error
- 401: Unauthorized
- 403: Not admin role
- 404: Drone not found
- 409: IP address conflict
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### PATCH /api/v1/drones/:id/status

Update drone status (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Drone UUID

**Request Body**:
```json
{
  "status": "maintenance"
}
```

**Allowed Status Values**:
- `idle`, `busy`, `charging`, `maintenance`, `offline`

**Response** (200 OK):
```json
{
  "id": "450e8400-e29b-41d4-a716-446655440000",
  "model": "Clover 4",
  "ip_address": "192.168.1.100",
  "status": "maintenance",
  "battery_level": 95.5,
  "created_at": "2024-01-10T08:00:00Z",
  "updated_at": "2024-01-15T13:10:00Z"
}
```

**Errors**:
- 400: Invalid status value or drone ID
- 401: Unauthorized
- 403: Not admin role
- 404: Drone not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### DELETE /api/v1/drones/:id

Delete drone (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Drone UUID

**Response** (200 OK):
```json
{
  "message": "Drone deleted successfully"
}
```

**Errors**:
- 400: Invalid drone ID format
- 401: Unauthorized
- 403: Not admin role
- 404: Drone not found
- 409: Drone has active deliveries (cannot delete)
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

### Parcel Automats

#### GET /api/v1/automats

Get list of all parcel automats.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**Response** (200 OK):
```json
[
  {
    "id": "850e8400-e29b-41d4-a716-446655440000",
    "city": "Moscow",
    "address": "Red Square, 1",
    "number_of_cells": 20,
    "ip_address": "192.168.1.50",
    "coordinates": "55.7558,37.6173",
    "aruco_id": 101,
    "is_working": true
  },
  {
    "id": "860e8400-e29b-41d4-a716-446655440000",
    "city": "Moscow",
    "address": "Tverskaya Street, 10",
    "number_of_cells": 30,
    "ip_address": "192.168.1.51",
    "coordinates": "55.7600,37.6120",
    "aruco_id": 102,
    "is_working": true
  }
]
```

**Response Fields**:
- `coordinates`: GPS coordinates in format "latitude,longitude"
- `aruco_id`: ArUco marker ID for drone navigation
- `is_working`: Operational status

**Errors**:
- 401: Unauthorized
- 403: Not admin role
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### GET /api/v1/automats/working

Get list of operational parcel automats.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**Response** (200 OK):
```json
[
  {
    "id": "850e8400-e29b-41d4-a716-446655440000",
    "city": "Moscow",
    "address": "Red Square, 1",
    "number_of_cells": 20,
    "ip_address": "192.168.1.50",
    "coordinates": "55.7558,37.6173",
    "aruco_id": 101,
    "is_working": true
  }
]
```

**Filtering**: Returns only automats where `is_working = true`

**Errors**:
- 401: Unauthorized
- 403: Not admin role
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### GET /api/v1/automats/:id

Get parcel automat details by ID.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Automat UUID

**Response** (200 OK):
```json
{
  "id": "850e8400-e29b-41d4-a716-446655440000",
  "city": "Moscow",
  "address": "Red Square, 1",
  "number_of_cells": 20,
  "ip_address": "192.168.1.50",
  "coordinates": "55.7558,37.6173",
  "aruco_id": 101,
  "is_working": true
}
```

**Errors**:
- 400: Invalid automat ID format
- 401: Unauthorized
- 403: Not admin role
- 404: Automat not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### POST /api/v1/automats

Create new parcel automat (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**Request Body**:
```json
{
  "city": "Saint Petersburg",
  "address": "Nevsky Prospect, 20",
  "number_of_cells": 25,
  "ip_address": "192.168.1.52",
  "coordinates": "59.9343,30.3351",
  "aruco_id": 103,
  "cells": [
    {
      "height": 30.0,
      "length": 40.0,
      "width": 35.0,
      "cell_number": 1,
      "type": "external"
    },
    {
      "height": 30.0,
      "length": 40.0,
      "width": 35.0,
      "cell_number": 11,
      "type": "internal"
    }
  ]
}
```

**Validation Rules**:
- `city`: Required, 1-255 characters
- `address`: Required, 1-255 characters
- `number_of_cells`: Required, positive integer
- `ip_address`: Required, valid IPv4/IPv6
- `coordinates`: Required, format "lat,lon"
- `aruco_id`: Required, unique integer
- `cells`: Required, array length must equal `number_of_cells`
- `cells[].type`: `external` (user-accessible) or `internal` (drone drop)

**Response** (201 Created):
```json
{
  "id": "870e8400-e29b-41d4-a716-446655440000",
  "city": "Saint Petersburg",
  "address": "Nevsky Prospect, 20",
  "number_of_cells": 25,
  "ip_address": "192.168.1.52",
  "coordinates": "59.9343,30.3351",
  "aruco_id": 103,
  "is_working": true
}
```

**Errors**:
- 400: Validation error or cells count mismatch
- 401: Unauthorized
- 403: Not admin role
- 409: ArUco ID or IP already exists
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### PUT /api/v1/automats/:id

Update parcel automat information (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Automat UUID

**Request Body**:
```json
{
  "city": "Moscow",
  "address": "Red Square, 1A",
  "ip_address": "192.168.1.55",
  "coordinates": "55.7559,37.6174"
}
```

**Response** (200 OK):
```json
{
  "id": "850e8400-e29b-41d4-a716-446655440000",
  "city": "Moscow",
  "address": "Red Square, 1A",
  "number_of_cells": 20,
  "ip_address": "192.168.1.55",
  "coordinates": "55.7559,37.6174",
  "aruco_id": 101,
  "is_working": true
}
```

**Errors**:
- 400: Invalid automat ID or validation error
- 401: Unauthorized
- 403: Not admin role
- 404: Automat not found
- 409: IP address conflict
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### PATCH /api/v1/automats/:id/status

Update parcel automat operational status (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Automat UUID

**Request Body**:
```json
{
  "is_working": false
}
```

**Response** (200 OK):
```json
{
  "id": "850e8400-e29b-41d4-a716-446655440000",
  "city": "Moscow",
  "address": "Red Square, 1",
  "number_of_cells": 20,
  "ip_address": "192.168.1.50",
  "coordinates": "55.7558,37.6173",
  "aruco_id": 101,
  "is_working": false
}
```

**Use Cases**:
- Maintenance mode: Set `is_working = false` to prevent new deliveries
- Back online: Set `is_working = true` to resume operations

**Errors**:
- 400: Invalid automat ID or request format
- 401: Unauthorized
- 403: Not admin role
- 404: Automat not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### GET /api/v1/automats/:id/cells

Get all locker cells for automat.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Automat UUID

**Response** (200 OK):
```json
[
  {
    "id": "950e8400-e29b-41d4-a716-446655440000",
    "post_id": "850e8400-e29b-41d4-a716-446655440000",
    "height": 30.0,
    "length": 40.0,
    "width": 35.0,
    "status": "available",
    "cell_number": 1,
    "type": "external"
  },
  {
    "id": "960e8400-e29b-41d4-a716-446655440000",
    "post_id": "850e8400-e29b-41d4-a716-446655440000",
    "height": 30.0,
    "length": 40.0,
    "width": 35.0,
    "status": "occupied",
    "cell_number": 2,
    "type": "external"
  },
  {
    "id": "970e8400-e29b-41d4-a716-446655440000",
    "post_id": "850e8400-e29b-41d4-a716-446655440000",
    "height": 30.0,
    "length": 40.0,
    "width": 35.0,
    "status": "available",
    "cell_number": 11,
    "type": "internal"
  }
]
```

**Cell Status Values**:
- `available`: Ready for assignment
- `reserved`: Reserved for incoming delivery
- `occupied`: Contains cargo awaiting pickup
- `maintenance`: Out of service

**Cell Types**:
- `external`: User-accessible cells for pickup
- `internal`: Drone drop cells (not user-accessible)

**Errors**:
- 400: Invalid automat ID format
- 401: Unauthorized
- 403: Not admin role
- 404: Automat not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### PATCH /api/v1/automats/:id/cells/:cellId

Update locker cell dimensions (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Automat UUID
- `cellId`: Cell UUID

**Request Body**:
```json
{
  "height": 35.0,
  "length": 45.0,
  "width": 40.0
}
```

**Response** (200 OK):
```json
{
  "id": "950e8400-e29b-41d4-a716-446655440000",
  "post_id": "850e8400-e29b-41d4-a716-446655440000",
  "height": 35.0,
  "length": 45.0,
  "width": 40.0,
  "status": "available",
  "cell_number": 1,
  "type": "external"
}
```

**Errors**:
- 400: Invalid IDs or validation error
- 401: Unauthorized
- 403: Not admin role
- 404: Cell or automat not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### DELETE /api/v1/automats/:id

Delete parcel automat (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Automat UUID

**Response** (200 OK):
```json
{
  "message": "Parcel automat deleted successfully"
}
```

**Cascade Behavior**:
- Deletes all associated locker cells (external and internal)
- Orders referencing this automat are NOT deleted (foreign key with CASCADE)

**Errors**:
- 400: Invalid automat ID format
- 401: Unauthorized
- 403: Not admin role
- 404: Automat not found
- 409: Automat has active orders (cannot delete)
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### POST /api/v1/automats/qr-scan (Public)

Scan QR code at parcel automat (no authentication required).

**Request Body**:
```json
{
  "qr_data": "user_uuid_from_qr_code",
  "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000"
}
```

**Response** (200 OK):
```json
{
  "valid": true,
  "cell_id": "950e8400-e29b-41d4-a716-446655440000",
  "cell_number": 5,
  "order_id": "750e8400-e29b-41d4-a716-446655440000",
  "message": "Cell 5 opened. Please retrieve your package."
}
```

**Business Logic**:
1. Validate QR code (check expiration in `users` table)
2. Find order for user with status `delivered` at this automat
3. Get assigned `locker_cell_id` from order
4. Send HTTP request to locker-agent to open cell
5. Update order status to `completed`
6. Return cell number to user

**Errors**:
- 400: Invalid request format
- 401: QR code expired or invalid
- 404: No package found for user at this automat
- 500: Database or locker-agent communication error

**Rate Limit**: 30 requests/minute per IP

---

#### POST /api/v1/automats/confirm-pickup (Public)

Confirm cargo pickup and close cell (no authentication required).

**Request Body**:
```json
{
  "order_id": "750e8400-e29b-41d4-a716-446655440000",
  "cell_id": "950e8400-e29b-41d4-a716-446655440000"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Pickup confirmed. Thank you!"
}
```

**Business Logic**:
1. Verify order exists and status is `completed`
2. Verify cell_id matches order's locker_cell_id
3. Send HTTP request to locker-agent to close cell
4. Update locker cell status to `available`
5. Send push notification to user (if FCM token registered)

**Errors**:
- 400: Invalid request format
- 404: Order or cell not found
- 409: Order status not `completed`
- 500: Database or locker-agent error

**Rate Limit**: 30 requests/minute per IP

---

### QR Codes

#### POST /api/v1/qr/validate (Public)

Validate QR code (used by locker-agent, no authentication required).

**Request Body**:
```json
{
  "qr_data": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Validation Logic**:
- Extract user UUID from QR data
- Check `qr_expires_at > CURRENT_TIMESTAMP` in `users` table
- Return user details if valid

**Response** (200 OK):
```json
{
  "valid": true,
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john.doe@example.com",
  "full_name": "John Doe",
  "access_expires_at": 1705320000,
  "refresh_expires_at": 1705406400
}
```

**Response if Invalid** (401 Unauthorized):
```json
{
  "error": "Invalid or expired QR code"
}
```

**Errors**:
- 400: Invalid request format
- 401: QR code expired or user not found

**Rate Limit**: 30 requests/minute per IP

---

#### GET /api/v1/qr/me

Get current user's QR code (auto-refresh if expired).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANS...",
  "issued_at": "2024-01-15T12:00:00Z",
  "expires_at": "2024-01-16T12:00:00Z"
}
```

**Auto-Refresh Logic**:
- If `qr_expires_at < CURRENT_TIMESTAMP`, automatically generate new QR
- Update `qr_issued_at` and `qr_expires_at` in database
- Return new QR code

**QR Code Format**:
- Content: User UUID
- Image: PNG, 300x300 pixels
- Encoding: Base64 data URI
- Expiration: 24 hours from issue

**Errors**:
- 401: Unauthorized
- 500: QR generation or database error

**Rate Limit**: 100 requests/minute per user

---

#### POST /api/v1/qr/refresh

Manually refresh QR code (extends expiration by 24 hours).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Response** (200 OK):
```json
{
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANS...",
  "issued_at": "2024-01-15T14:00:00Z",
  "access_expires_at": "2024-01-15T14:00:00Z",
  "refresh_expires_at": "2024-01-16T14:00:00Z"
}
```

**Use Case**:
- User manually refreshes before QR expires
- Useful before going to pick up package

**Errors**:
- 401: Unauthorized
- 500: QR generation or database error

**Rate Limit**: 100 requests/minute per user

---

### Deliveries

#### GET /api/v1/deliveries/:id

Get delivery details by ID.

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Delivery UUID

**Response** (200 OK):
```json
{
  "id": "780e8400-e29b-41d4-a716-446655440000",
  "order_id": "750e8400-e29b-41d4-a716-446655440000",
  "drone_id": "450e8400-e29b-41d4-a716-446655440000",
  "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000",
  "internal_locker_cell_id": "970e8400-e29b-41d4-a716-446655440000",
  "status": "in_progress",
  "started_at": "2024-01-15T13:00:00Z",
  "completed_at": null
}
```

**Delivery Status Values**:
- `pending`: Created, no drone assigned
- `assigned`: Drone assigned, waiting to start
- `in_progress`: Drone executing delivery
- `delivered`: Cargo dropped in internal cell
- `completed`: Cargo moved to external cell
- `failed`: Delivery failed (drone error, weather, etc.)

**Errors**:
- 400: Invalid delivery ID format
- 401: Unauthorized
- 403: Not admin role
- 404: Delivery not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### PUT /api/v1/deliveries/:id/status

Update delivery status (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**URL Parameters**:
- `id`: Delivery UUID

**Request Body**:
```json
{
  "status": "completed"
}
```

**Allowed Transitions**:
- `pending` → `assigned` → `in_progress` → `delivered` → `completed`
- Any status → `failed`

**Response** (200 OK):
```json
{
  "success": true
}
```

**Errors**:
- 400: Invalid status or transition not allowed
- 401: Unauthorized
- 403: Not admin role
- 404: Delivery not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

#### POST /api/v1/deliveries/confirm-loaded

Confirm goods loaded from cell into drone (internal use).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Request Body**:
```json
{
  "order_id": "750e8400-e29b-41d4-a716-446655440000",
  "cell_id": "970e8400-e29b-41d4-a716-446655440000"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "delivery_id": "780e8400-e29b-41d4-a716-446655440000",
  "message": "Goods loaded, ready for takeoff"
}
```

**Business Logic**:
1. Find delivery for order
2. Verify cell matches internal_locker_cell_id
3. Update delivery status to `in_progress`
4. Update internal cell status to `available`
5. Update order status to `processing`

**Errors**:
- 400: Invalid request or cell mismatch
- 401: Unauthorized
- 404: Order or delivery not found
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

### Monitoring

#### GET /api/v1/monitoring/system-status

Get comprehensive system status (admin only).

**Request Headers**:
```http
Authorization: Bearer <access_token>
```

**Authorization**: Requires `admin` role

**Response** (200 OK):
```json
{
  "drones": [
    {
      "id": "450e8400-e29b-41d4-a716-446655440000",
      "model": "Clover 4",
      "status": "busy",
      "battery_level": 87.5
    }
  ],
  "automats": [
    {
      "parcel_automat": {
        "id": "850e8400-e29b-41d4-a716-446655440000",
        "city": "Moscow",
        "address": "Red Square, 1",
        "is_working": true
      },
      "cells": [
        {
          "id": "950e8400-e29b-41d4-a716-446655440000",
          "status": "available",
          "cell_number": 1,
          "type": "external"
        }
      ]
    }
  ],
  "active_deliveries": [
    {
      "id": "780e8400-e29b-41d4-a716-446655440000",
      "order_id": "750e8400-e29b-41d4-a716-446655440000",
      "drone_id": "450e8400-e29b-41d4-a716-446655440000",
      "status": "in_progress",
      "started_at": "2024-01-15T13:00:00Z"
    }
  ]
}
```

**Use Case**:
- Admin dashboard overview
- Real-time system monitoring
- Capacity planning

**Errors**:
- 401: Unauthorized
- 403: Not admin role
- 500: Database error

**Rate Limit**: 100 requests/minute per user

---

## gRPC API (Go-Orchestrator)

**Protocol**: gRPC (HTTP/2)  
**Protocol Buffers**: `proto/orchestrator.proto`  
**Endpoint**: `orchestrator:50051`

### Service: OrchestratorService

#### RequestCellOpen

Request opening of internal locker cell for cargo drop.

**Proto Definition**:
```protobuf
service OrchestratorService {
  rpc RequestCellOpen(CellOpenRequest) returns (CellOpenResponse);
}

message CellOpenRequest {
  string delivery_id = 1;
  string parcel_automat_id = 2;
  CellType cell_type = 3;
}

enum CellType {
  EXTERNAL = 0;
  INTERNAL = 1;
}

message CellOpenResponse {
  bool success = 1;
  string cell_id = 2;
  string error_message = 3;
}
```

**Request Example** (Go):
```go
req := &pb.CellOpenRequest{
    DeliveryId:       "780e8400-e29b-41d4-a716-446655440000",
    ParcelAutomatId:  "850e8400-e29b-41d4-a716-446655440000",
    CellType:         pb.CellType_INTERNAL,
}

resp, err := client.RequestCellOpen(ctx, req)
```

**Response Example**:
```go
// Success
&CellOpenResponse{
    Success:      true,
    CellId:       "970e8400-e29b-41d4-a716-446655440000",
    ErrorMessage: "",
}

// Failure
&CellOpenResponse{
    Success:      false,
    CellId:       "",
    ErrorMessage: "No available internal cells",
}
```

**Business Logic**:
1. Find available internal cell at parcel automat
2. Update cell status to `reserved`
3. Send HTTP POST to locker-agent `/api/cells/internal/:number/open`
4. Update delivery record with `internal_locker_cell_id`
5. Return cell UUID

**gRPC Error Codes**:
- `OK`: Successful cell opening
- `NOT_FOUND`: Parcel automat or delivery not found
- `UNAVAILABLE`: No available cells or locker-agent offline
- `INTERNAL`: Database or network error

**Retry Policy**:
- Max attempts: 3
- Backoff: Exponential (100ms, 200ms, 400ms)
- Timeout: 5 seconds per attempt

---

## WebSocket API (Drone-Service)

**Protocol**: WebSocket (RFC 6455)  
**Base URL**: `ws://drone-service:8081/ws`  
**Message Format**: JSON

### WebSocket: /ws/drone

Drone agent connection for task assignment and telemetry.

**Connection**:
```javascript
const ws = new WebSocket('ws://drone-service:8081/ws/drone');
```

**Message Types** (Drone → Service):

#### 1. Register Message

Sent immediately after connection to register drone.

```json
{
  "type": "register",
  "payload": {
    "drone_id": "450e8400-e29b-41d4-a716-446655440000",
    "model": "Clover 4",
    "battery_level": 95.5
  },
  "timestamp": "2024-01-15T12:00:00Z"
}
```

**Server Response**:
```json
{
  "type": "register_ack",
  "payload": {
    "status": "connected",
    "message": "Drone registered successfully"
  },
  "timestamp": "2024-01-15T12:00:01Z"
}
```

---

#### 2. Telemetry Message

Sent every 3 seconds with drone status.

```json
{
  "type": "telemetry",
  "payload": {
    "drone_id": "450e8400-e29b-41d4-a716-446655440000",
    "battery_level": 87.5,
    "position": {
      "latitude": 55.7558,
      "longitude": 37.6173,
      "altitude": 15.5
    },
    "status": "flying"
  },
  "timestamp": "2024-01-15T12:05:30Z"
}
```

**Status Values**:
- `idle`: On ground, ready
- `flying`: In flight
- `hovering`: Hovering at position
- `landing`: Landing sequence
- `error`: Error state

---

#### 3. Event Message

Sent when delivery events occur.

```json
{
  "type": "event",
  "payload": {
    "drone_id": "450e8400-e29b-41d4-a716-446655440000",
    "delivery_id": "780e8400-e29b-41d4-a716-446655440000",
    "event_type": "arrived",
    "timestamp": "2024-01-15T12:10:00Z"
  },
  "timestamp": "2024-01-15T12:10:00Z"
}
```

**Event Types**:
- `task_assigned`: Delivery task received
- `navigating`: En route to destination
- `arrived`: Arrived at parcel automat
- `waiting_drop`: Waiting for cell to open
- `drop_confirmed`: Cargo released
- `returning`: Returning to base
- `completed`: Delivery completed
- `error`: Error occurred

---

#### 4. Video Frame Message

Sent continuously for video streaming.

```json
{
  "type": "video_frame",
  "payload": {
    "drone_id": "450e8400-e29b-41d4-a716-446655440000",
    "frame": "/9j/4AAQSkZJRgABAQAAAQABAAD...",
    "timestamp": "2024-01-15T12:05:30.123Z"
  },
  "timestamp": "2024-01-15T12:05:30.123Z"
}
```

**Frame Format**:
- Encoding: JPEG
- Resolution: 640x480
- Format: Base64 string
- Rate: ~10 FPS

---

**Message Types** (Service → Drone):

#### 1. Task Message

Delivery task assignment.

```json
{
  "type": "task",
  "payload": {
    "delivery_id": "780e8400-e29b-41d4-a716-446655440000",
    "order_id": "750e8400-e29b-41d4-a716-446655440000",
    "destination": {
      "latitude": 55.7558,
      "longitude": 37.6173,
      "aruco_id": 101
    },
    "priority": "normal"
  },
  "timestamp": "2024-01-15T12:00:00Z"
}
```

**Priority Values**:
- `normal`: Standard delivery queue
- `priority`: High-priority queue

---

#### 2. Command Message

Control commands to drone.

```json
{
  "type": "command",
  "payload": {
    "command": "drop_cargo",
    "delivery_id": "780e8400-e29b-41d4-a716-446655440000"
  },
  "timestamp": "2024-01-15T12:10:00Z"
}
```

**Command Types**:
- `drop_cargo`: Release cargo via servo
- `return_to_base`: Abort delivery and return
- `emergency_land`: Emergency landing
- `pause`: Pause current operation

---

### WebSocket: /ws/admin

Admin panel connection for monitoring.

**Connection**:
```javascript
const ws = new WebSocket('ws://drone-service:8081/ws/admin');
```

**Message Types** (Service → Admin):

#### 1. Drone Status Update

```json
{
  "type": "drone_status",
  "payload": {
    "drone_id": "450e8400-e29b-41d4-a716-446655440000",
    "status": "busy",
    "battery_level": 87.5,
    "position": {
      "latitude": 55.7558,
      "longitude": 37.6173,
      "altitude": 15.5
    },
    "current_delivery_id": "780e8400-e29b-41d4-a716-446655440000"
  },
  "timestamp": "2024-01-15T12:05:30Z"
}
```

---

#### 2. Delivery Update

```json
{
  "type": "delivery_update",
  "payload": {
    "delivery_id": "780e8400-e29b-41d4-a716-446655440000",
    "status": "in_progress",
    "drone_id": "450e8400-e29b-41d4-a716-446655440000",
    "progress": 45
  },
  "timestamp": "2024-01-15T12:05:30Z"
}
```

**Progress**: Percentage (0-100) of delivery completion

---

#### 3. System Alert

```json
{
  "type": "alert",
  "payload": {
    "severity": "warning",
    "message": "Drone battery low: 15%",
    "drone_id": "450e8400-e29b-41d4-a716-446655440000"
  },
  "timestamp": "2024-01-15T12:05:30Z"
}
```

**Severity Levels**:
- `info`: Informational message
- `warning`: Warning condition
- `error`: Error condition
- `critical`: Critical system issue

---

### WebSocket: /ws/drone/:drone_id/video

Video stream relay to admin panel.

**Connection**:
```javascript
const droneId = "450e8400-e29b-41d4-a716-446655440000";
const ws = new WebSocket(`ws://drone-service:8081/ws/drone/${droneId}/video`);
```

**Message Format**:
```json
{
  "type": "video_frame",
  "payload": {
    "frame": "/9j/4AAQSkZJRgABAQAAAQABAAD...",
    "timestamp": "2024-01-15T12:05:30.123Z"
  }
}
```

**Stream Properties**:
- Format: JPEG frames
- Encoding: Base64
- Rate: ~10 FPS
- Latency: < 500ms

---

## HTTP API (Locker-Agent)

**Base URL**: `http://orange-pi-ip:8082/api`  
**Protocol**: HTTP/1.1  
**Content-Type**: `application/json`

### POST /api/cells/sync

Synchronize cell mapping with orchestrator.

**Request Body**:
```json
{
  "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000",
  "external_cells": {
    "1": "950e8400-e29b-41d4-a716-446655440000",
    "2": "960e8400-e29b-41d4-a716-446655440000"
  },
  "internal_cells": {
    "11": "970e8400-e29b-41d4-a716-446655440000",
    "12": "980e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Mapping Format**:
- Key: Physical cell number (integer as string)
- Value: Cell UUID from database

**Response** (200 OK):
```json
{
  "message": "Cells synchronized successfully",
  "cells_count": 2,
  "internal_cells_count": 2,
  "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000"
}
```

**Purpose**: Initialize locker agent with UUID mapping on startup

---

### GET /api/cells/mapping

Get current cell mapping.

**Response** (200 OK):
```json
{
  "mapping": {
    "1": {
      "cell_uuid": "950e8400-e29b-41d4-a716-446655440000",
      "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000"
    },
    "2": {
      "cell_uuid": "960e8400-e29b-41d4-a716-446655440000",
      "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000"
    }
  },
  "cells_count": 2,
  "internal_mapping": {
    "11": {
      "cell_uuid": "970e8400-e29b-41d4-a716-446655440000",
      "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000"
    }
  },
  "internal_cells_count": 1,
  "parcel_automat_id": "850e8400-e29b-41d4-a716-446655440000",
  "initialized": true
}
```

---

### POST /api/cells/:number/open

Open external cell door.

**URL Parameters**:
- `number`: Physical cell number (integer)

**Request Body** (optional):
```json
{
  "order_number": "ORD-12345"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "cell_number": 5,
  "cell_uuid": "950e8400-e29b-41d4-a716-446655440000",
  "opened_at": "2024-01-15T12:15:00Z"
}
```

**Hardware Action**:
- Sends serial command to Arduino: `OPEN_5`
- Arduino controls servo/solenoid to unlock cell

**Errors**:
- 400: Invalid cell number or not initialized
- 500: Arduino communication error

---

### POST /api/cells/internal/:number/open

Open internal cell door for drone cargo drop.

**URL Parameters**:
- `number`: Physical cell number (integer)

**Request Body**:
```json
{
  "delivery_id": "780e8400-e29b-41d4-a716-446655440000"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "cell_number": 11,
  "cell_uuid": "970e8400-e29b-41d4-a716-446655440000",
  "opened_at": "2024-01-15T12:10:00Z"
}
```

**Hardware Action**:
- Sends serial command: `OPEN_11`
- Used by drone service via gRPC → orchestrator → locker-agent flow

---

### POST /api/cells/prepare

Prepare cell for incoming delivery (reserve cell).

**Request Body**:
```json
{
  "cell_uuid": "970e8400-e29b-41d4-a716-446655440000",
  "delivery_id": "780e8400-e29b-41d4-a716-446655440000"
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "cell_number": 11,
  "message": "Cell prepared and reserved"
}
```

**Purpose**: Reserve cell before drone arrives

---

### GET /api/cells/count

Get total cell count.

**Response** (200 OK):
```json
{
  "external_cells": 10,
  "internal_cells": 10,
  "total": 20
}
```

---

### POST /api/qr/scan

Scan QR code with camera.

**Request Body**:
```json
{
  "timeout": 10
}
```

**Timeout**: Seconds to wait for QR detection (default: 10)

**Response** (200 OK):
```json
{
  "success": true,
  "qr_data": "550e8400-e29b-41d4-a716-446655440000",
  "scanned_at": "2024-01-15T12:20:00Z"
}
```

**Hardware**: Uses OpenCV with QR camera to detect codes

**Errors**:
- 408: Timeout - no QR code detected
- 500: Camera error

---

### GET /health

Health check endpoint.

**Response** (200 OK):
```json
{
  "status": "healthy",
  "service": "locker-agent",
  "initialized": true,
  "arduino_connected": true
}
```

---

## Appendix

### Error Code Reference

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| INVALID_REQUEST | 400 | Request validation failed |
| UNAUTHORIZED | 401 | Missing or invalid authentication |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource does not exist |
| CONFLICT | 409 | Resource conflict or duplicate |
| RATE_LIMIT_EXCEEDED | 429 | Too many requests |
| INTERNAL_ERROR | 500 | Server-side error |
| SERVICE_UNAVAILABLE | 503 | Service temporarily down |

### Rate Limit Tiers

| Endpoint Category | Limit | Window |
|-------------------|-------|--------|
| Authentication | 10 req | 1 minute per IP |
| Order Creation | 20 req | 1 minute per user |
| QR Operations | 30 req | 1 minute per IP |
| General API | 100 req | 1 minute per user |
| Admin API | 200 req | 1 minute per user |

### Webhook Events (Future)

Planned webhook support for external integrations:

- `order.created`
- `order.completed`
- `delivery.started`
- `delivery.completed`
- `drone.battery_low`
- `automat.offline`

---

**Documentation Version**: 1.0  
**Last Updated**: December 13, 2025  
**Maintainers**: SkyPost Development Team
