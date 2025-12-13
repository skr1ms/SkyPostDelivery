# Database Documentation

## Overview

SkyPost Delivery uses PostgreSQL 17 as the primary relational database management system. The database schema is designed to support the complete delivery workflow, from user authentication and order creation to drone delivery execution and parcel locker management.

**Database Name**: `skypost_delivery`  
**PostgreSQL Version**: 17-alpine  
**Character Encoding**: UTF-8  
**Extensions**: `uuid-ossp` for UUID generation

## Database Schema

### Entity-Relationship Diagram

![ERD Diagram](image/ERD.png)

## Table Definitions

### 1. users

Stores user account information for authentication and authorization.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone_number VARCHAR(20) UNIQUE,
    pass_hash VARCHAR(255),
    phone_verified BOOLEAN NOT NULL DEFAULT false,
    verification_code VARCHAR(10),
    code_expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role VARCHAR(50) NOT NULL DEFAULT 'client',
    qr_issued_at TIMESTAMP,
    qr_expires_at TIMESTAMP
);
```

**Columns**:
- `id`: Primary key, UUID generated automatically
- `full_name`: User's full name (required)
- `email`: Email address (unique, optional for phone-only auth)
- `phone_number`: Phone number (unique, for SMS-based auth)
- `pass_hash`: Bcrypt hashed password (optional for phone-only auth)
- `phone_verified`: Whether phone number is verified via SMS code
- `verification_code`: Temporary SMS verification code (6 digits)
- `code_expires_at`: Expiration timestamp for verification code (5 minutes)
- `created_at`: Account creation timestamp
- `role`: User role (`client`, `admin`)
- `qr_issued_at`: Timestamp when QR code was last issued
- `qr_expires_at`: QR code expiration timestamp (24 hours from issue)

**Indexes**:
- `idx_users_phone_number`: Fast lookup by phone number for authentication

**Constraints**:
- `email` OR `phone_number` must be provided (application-level validation)
- `verification_code` is 6 digits, expires after 5 minutes
- `qr_expires_at` is 24 hours after `qr_issued_at`

**Sample Queries**:
```sql
-- Find user by phone number
SELECT * FROM users WHERE phone_number = '+79991234567';

-- Verify phone code
UPDATE users 
SET phone_verified = true, verification_code = NULL 
WHERE phone_number = '+79991234567' 
  AND verification_code = '123456' 
  AND code_expires_at > CURRENT_TIMESTAMP;

-- Check QR code validity
SELECT * FROM users 
WHERE id = 'uuid' 
  AND qr_expires_at > CURRENT_TIMESTAMP;
```

### 2. user_devices

Stores FCM (Firebase Cloud Messaging) tokens for push notifications.

```sql
CREATE TABLE user_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    platform VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Columns**:
- `id`: Primary key
- `user_id`: Foreign key to `users` table (CASCADE delete)
- `token`: FCM device token (unique)
- `platform`: Device platform (`ios`, `android`, `web`)
- `created_at`: Token registration timestamp
- `updated_at`: Last token update timestamp

**Indexes**:
- `idx_user_devices_user_id`: Fast lookup of user's devices

**Constraints**:
- Foreign key: `user_id` → `users(id)` with CASCADE delete
- Unique constraint on `token`

**Sample Queries**:
```sql
-- Get all devices for user
SELECT * FROM user_devices WHERE user_id = 'uuid';

-- Upsert device token
INSERT INTO user_devices (user_id, token, platform)
VALUES ('uuid', 'fcm_token', 'android')
ON CONFLICT (token) DO UPDATE SET updated_at = CURRENT_TIMESTAMP;
```

### 3. goods

Catalog of available goods for delivery.

```sql
CREATE TABLE goods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    weight DECIMAL(10, 2) NOT NULL,
    height DECIMAL(10, 2) NOT NULL,
    length DECIMAL(10, 2) NOT NULL,
    width DECIMAL(10, 2) NOT NULL,
    quantity_available INTEGER NOT NULL DEFAULT 0
);
```

**Columns**:
- `id`: Primary key
- `name`: Good name/description
- `weight`: Weight in kilograms (up to 99,999,999.99 kg)
- `height`: Height in centimeters
- `length`: Length in centimeters
- `width`: Width in centimeters
- `quantity_available`: Current stock quantity

**Constraints**:
- All dimensions and weight must be positive (application-level validation)
- `quantity_available` cannot be negative

**Sample Queries**:
```sql
-- Find goods that fit in a locker cell
SELECT * FROM goods g
WHERE EXISTS (
    SELECT 1 FROM locker_cells_out lc
    WHERE lc.status = 'available'
      AND lc.height >= g.height
      AND lc.length >= g.length
      AND lc.width >= g.width
);

-- Decrease quantity after order
UPDATE goods 
SET quantity_available = quantity_available - 1
WHERE id = 'uuid' AND quantity_available > 0;
```

### 4. drones

Drone fleet registry with status and telemetry.

```sql
CREATE TABLE drones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'idle',
    battery_level DECIMAL(5, 2) DEFAULT 100.0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Columns**:
- `id`: Primary key
- `model`: Drone model (e.g., "Clover 4")
- `ip_address`: Drone's network IP address (IPv4/IPv6)
- `status`: Current status (see status values below)
- `battery_level`: Battery percentage (0.00 to 100.00)
- `created_at`: Drone registration timestamp
- `updated_at`: Last telemetry update timestamp

**Status Values**:
- `idle`: Ready for task assignment
- `busy`: Currently executing delivery
- `charging`: Charging at base station
- `maintenance`: Under maintenance
- `offline`: Not connected

**Indexes**:
- `idx_drones_ip_address`: Fast lookup by IP
- `idx_drones_status`: Fast filtering by status

**Sample Queries**:
```sql
-- Find available drones with sufficient battery
SELECT * FROM drones 
WHERE status = 'idle' 
  AND battery_level > 30.0
ORDER BY battery_level DESC
LIMIT 1;

-- Update drone status after task assignment
UPDATE drones 
SET status = 'busy', updated_at = CURRENT_TIMESTAMP
WHERE id = 'uuid';
```

### 5. parcel_automats

Parcel locker locations with coordinates and ArUco markers.

```sql
CREATE TABLE parcel_automats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    city VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    number_of_cells INTEGER NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    coordinates VARCHAR(255) NOT NULL,
    aruco_id INTEGER NOT NULL,
    is_working BOOLEAN NOT NULL DEFAULT true
);
```

**Columns**:
- `id`: Primary key
- `city`: City name
- `address`: Street address
- `number_of_cells`: Total number of cells (external + internal)
- `ip_address`: Orange Pi IP address for locker agent
- `coordinates`: GPS coordinates (format: "55.7558,37.6173")
- `aruco_id`: ArUco marker ID for drone navigation
- `is_working`: Operational status

**Indexes**:
- `idx_parcel_automats_ip_address`: Fast lookup by IP

**Constraints**:
- `aruco_id` should be unique (application-level validation)
- `coordinates` format validated by application

**Sample Queries**:
```sql
-- Find nearest parcel automat (application logic)
SELECT * FROM parcel_automats 
WHERE is_working = true
ORDER BY ST_Distance(
    ST_SetSRID(ST_MakePoint(lon, lat), 4326),
    ST_SetSRID(ST_MakePoint(:user_lon, :user_lat), 4326)
)
LIMIT 1;

-- Get automat by ArUco ID
SELECT * FROM parcel_automats WHERE aruco_id = 123;
```

### 6. locker_cells_out

External locker cells accessible to users for cargo pickup.

```sql
CREATE TABLE locker_cells_out (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES parcel_automats(id) ON DELETE CASCADE,
    height DECIMAL(10, 2) NOT NULL,
    length DECIMAL(10, 2) NOT NULL,
    width DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    cell_number INTEGER
);
```

**Columns**:
- `id`: Primary key (UUID used for QR code validation)
- `post_id`: Foreign key to `parcel_automats`
- `height`, `length`, `width`: Cell dimensions in centimeters
- `status`: Cell availability status
- `cell_number`: Physical cell number (for Arduino control)

**Status Values**:
- `available`: Ready for assignment
- `occupied`: Contains cargo awaiting pickup
- `reserved`: Reserved for incoming delivery
- `maintenance`: Out of service

**Indexes**:
- `idx_locker_cells_out_status`: Fast filtering by status

**Constraints**:
- Foreign key: `post_id` → `parcel_automats(id)` with CASCADE delete

**Sample Queries**:
```sql
-- Find available cell that fits good dimensions
SELECT * FROM locker_cells_out
WHERE post_id = 'uuid'
  AND status = 'available'
  AND height >= :good_height
  AND length >= :good_length
  AND width >= :good_width
ORDER BY (height * length * width) ASC
LIMIT 1;

-- Reserve cell for delivery
UPDATE locker_cells_out
SET status = 'reserved'
WHERE id = 'uuid' AND status = 'available';
```

### 7. locker_cells_internal

Internal locker cells for drone cargo drop (not accessible to users).

```sql
CREATE TABLE locker_cells_internal (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES parcel_automats(id) ON DELETE CASCADE,
    height DECIMAL(10, 2) NOT NULL,
    length DECIMAL(10, 2) NOT NULL,
    width DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    cell_number INTEGER
);
```

**Columns**: Same as `locker_cells_out`

**Purpose**: Internal cells are opened by drone service for cargo drop. After delivery completion, cargo is manually moved to external cells by automat staff (or automated conveyor in future versions).

**Status Values**: Same as `locker_cells_out`

**Indexes**:
- `idx_locker_cells_internal_status`: Fast filtering by status

**Sample Queries**:
```sql
-- Find available internal cell for drone drop
SELECT * FROM locker_cells_internal
WHERE post_id = 'uuid'
  AND status = 'available'
ORDER BY cell_number ASC
LIMIT 1;
```

### 8. orders

User orders linking goods, users, and delivery destinations.

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    good_id UUID NOT NULL REFERENCES goods(id) ON DELETE CASCADE,
    parcel_automat_id UUID NOT NULL REFERENCES parcel_automats(id) ON DELETE CASCADE,
    locker_cell_id UUID REFERENCES locker_cells_out(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Columns**:
- `id`: Primary key
- `user_id`: Foreign key to `users`
- `good_id`: Foreign key to `goods`
- `parcel_automat_id`: Destination automat
- `locker_cell_id`: Assigned external cell (NULL until delivery)
- `status`: Order status (see status values below)
- `created_at`: Order creation timestamp

**Status Values**:
- `pending`: Awaiting processing
- `processing`: Drone assigned, delivery in progress
- `delivered`: Cargo in locker, awaiting pickup
- `completed`: User picked up cargo
- `cancelled`: Order cancelled

**Indexes**:
- `idx_orders_user_id`: Fast user order lookup
- `idx_orders_status`: Fast filtering by status

**Constraints**:
- Foreign key: `user_id` → `users(id)` with CASCADE delete
- Foreign key: `good_id` → `goods(id)` with CASCADE delete
- Foreign key: `parcel_automat_id` → `parcel_automats(id)` with CASCADE delete
- Foreign key: `locker_cell_id` → `locker_cells_out(id)` with SET NULL

**Sample Queries**:
```sql
-- Get user's orders
SELECT o.*, g.name AS good_name, pa.address 
FROM orders o
JOIN goods g ON o.good_id = g.id
JOIN parcel_automats pa ON o.parcel_automat_id = pa.id
WHERE o.user_id = 'uuid'
ORDER BY o.created_at DESC;

-- Get pending orders for order worker
SELECT * FROM orders 
WHERE status = 'pending'
ORDER BY created_at ASC
FOR UPDATE SKIP LOCKED;
```

### 9. deliveries

Delivery execution records linking orders to drones and cells.

```sql
CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    drone_id UUID REFERENCES drones(id) ON DELETE SET NULL,
    parcel_automat_id UUID NOT NULL REFERENCES parcel_automats(id) ON DELETE CASCADE,
    internal_locker_cell_id UUID REFERENCES locker_cells_internal(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

**Columns**:
- `id`: Primary key
- `order_id`: Foreign key to `orders`
- `drone_id`: Assigned drone (NULL if unassigned)
- `parcel_automat_id`: Destination automat
- `internal_locker_cell_id`: Internal cell for cargo drop
- `status`: Delivery status (see status values below)
- `started_at`: Delivery start timestamp
- `completed_at`: Delivery completion timestamp

**Status Values**:
- `pending`: Created, no drone assigned
- `assigned`: Drone assigned, waiting to start
- `in_progress`: Drone executing delivery
- `delivered`: Cargo dropped in internal cell
- `completed`: Cargo moved to external cell
- `failed`: Delivery failed (drone error, weather, etc.)

**Indexes**:
- `idx_deliveries_status`: Fast filtering by status
- `idx_deliveries_drone_id`: Fast drone delivery lookup

**Constraints**:
- Foreign key: `order_id` → `orders(id)` with CASCADE delete
- Foreign key: `drone_id` → `drones(id)` with SET NULL
- Foreign key: `parcel_automat_id` → `parcel_automats(id)` with CASCADE delete
- Foreign key: `internal_locker_cell_id` → `locker_cells_internal(id)` with SET NULL

**Sample Queries**:
```sql
-- Get active deliveries for drone
SELECT * FROM deliveries 
WHERE drone_id = 'uuid' 
  AND status IN ('assigned', 'in_progress');

-- Update delivery status
UPDATE deliveries
SET status = 'in_progress', started_at = CURRENT_TIMESTAMP
WHERE id = 'uuid';

-- Get delivery statistics
SELECT 
    status, 
    COUNT(*) AS count,
    AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) AS avg_duration_seconds
FROM deliveries
WHERE completed_at IS NOT NULL
GROUP BY status;
```

## Stored Functions

### update_drone_battery

Updates drone battery level and timestamp.

```sql
CREATE OR REPLACE FUNCTION update_drone_battery(
    p_drone_id UUID,
    p_battery_level DECIMAL(5,2)
) RETURNS void AS $$
BEGIN
    UPDATE drones
    SET 
        battery_level = p_battery_level,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = p_drone_id;
END;
$$ LANGUAGE plpgsql;
```

**Usage**:
```sql
SELECT update_drone_battery('drone-uuid', 87.5);
```

**Purpose**: Atomic battery level update from telemetry reports.

## Indexes

### Primary Indexes (Automatically Created)

All tables have primary key indexes on `id` column (UUID).

### Secondary Indexes

**Performance Optimization**:
- `idx_orders_user_id`: Fast user order queries
- `idx_orders_status`: Order worker filtering
- `idx_deliveries_status`: Delivery status filtering
- `idx_deliveries_drone_id`: Drone task queries
- `idx_locker_cells_out_status`: Cell availability queries
- `idx_locker_cells_internal_status`: Internal cell queries
- `idx_users_phone_number`: Phone-based authentication
- `idx_user_devices_user_id`: User device lookup
- `idx_parcel_automats_ip_address`: Automat IP lookup
- `idx_drones_ip_address`: Drone IP lookup
- `idx_drones_status`: Available drone queries

**Index Usage Examples**:
```sql
-- Using idx_orders_status
EXPLAIN ANALYZE
SELECT * FROM orders WHERE status = 'pending';

-- Using idx_drones_status
EXPLAIN ANALYZE
SELECT * FROM drones WHERE status = 'idle';
```

## Relationships and Foreign Keys

### Cascade Delete Relationships

**users → user_devices**: When user is deleted, all device tokens are deleted  
**users → orders**: When user is deleted, all orders are deleted  
**goods → orders**: When good is deleted, all orders are deleted  
**parcel_automats → orders**: When automat is deleted, all orders are deleted  
**parcel_automats → locker_cells_out**: When automat is deleted, all external cells are deleted  
**parcel_automats → locker_cells_internal**: When automat is deleted, all internal cells are deleted  
**orders → deliveries**: When order is deleted, delivery record is deleted

### Set NULL Relationships

**locker_cells_out → orders**: When cell is deleted, order remains with NULL `locker_cell_id`  
**drones → deliveries**: When drone is deleted, delivery record remains with NULL `drone_id`  
**locker_cells_internal → deliveries**: When cell is deleted, delivery remains with NULL `internal_locker_cell_id`

### Referential Integrity

All foreign keys enforce referential integrity at database level. No orphaned records are allowed.

## Transactions and Concurrency

### ACID Properties

- **Atomicity**: All-or-nothing transactions
- **Consistency**: Foreign key constraints enforced
- **Isolation**: Read Committed isolation level (default)
- **Durability**: Write-Ahead Logging (WAL)

### Locking Strategy

**Pessimistic Locking** (Order Worker):
```sql
BEGIN;
SELECT * FROM orders 
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- Process order...

UPDATE orders SET status = 'processing' WHERE id = 'uuid';
COMMIT;
```

`FOR UPDATE SKIP LOCKED` prevents race conditions when multiple order workers run concurrently.

**Optimistic Locking** (Cell Assignment):
```sql
UPDATE locker_cells_out
SET status = 'reserved'
WHERE id = 'uuid' AND status = 'available'
RETURNING *;
```

Returns empty result if cell was already reserved by another transaction.

### Transaction Isolation

**Default**: Read Committed
- No dirty reads
- No phantom reads for single-row operations
- Allows concurrent order processing

**Critical Sections**:
- Order creation: Serializable isolation to prevent double-booking
- Cell assignment: Read Committed with optimistic locking

## Data Migrations

### Migration Tool

**golang-migrate/migrate** is used for version-controlled schema changes.

**Migration Files**:
- `000001_init_schema.up.sql`: Initial schema creation
- `000001_init_schema.down.sql`: Schema rollback

### Migration Commands

```bash
# Apply migrations
migrate -path migrations -database "postgresql://user:pass@host:5432/db?sslmode=disable" up

# Rollback migrations
migrate -path migrations -database "postgresql://user:pass@host:5432/db?sslmode=disable" down

# Check migration version
migrate -path migrations -database "postgresql://user:pass@host:5432/db?sslmode=disable" version
```

### Migration Best Practices

1. **Backward Compatibility**: Ensure migrations don't break running services
2. **Idempotency**: Use `CREATE TABLE IF NOT EXISTS`
3. **Rollback Support**: Always provide `.down.sql` files
4. **Data Preservation**: Use transactions for data migrations
5. **Testing**: Test migrations on staging environment first

## Query Optimization

### SQLC Code Generation

Type-safe SQL queries are generated using SQLC from `.sql` files in `queries/` directory.

**Example SQLC Query**:
```sql
-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone_number = $1 LIMIT 1;

-- name: CreateOrder :one
INSERT INTO orders (user_id, good_id, parcel_automat_id, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPendingOrders :many
SELECT * FROM orders WHERE status = 'pending' ORDER BY created_at ASC;
```

**Generated Go Code**:
```go
func (q *Queries) GetUserByPhone(ctx context.Context, phoneNumber string) (User, error)
func (q *Queries) CreateOrder(ctx context.Context, arg CreateOrderParams) (Order, error)
func (q *Queries) GetPendingOrders(ctx context.Context) ([]Order, error)
```

### Query Performance Tips

1. **Use Indexes**: Ensure WHERE clauses use indexed columns
2. **Avoid SELECT \***: Select only needed columns
3. **Use LIMIT**: Paginate large result sets
4. **Batch Operations**: Use INSERT/UPDATE with multiple rows
5. **Connection Pooling**: Reuse database connections

**Example Optimized Query**:
```sql
-- Bad: No index, SELECT *
SELECT * FROM orders WHERE good_id = 'uuid';

-- Good: Uses index, specific columns
SELECT id, user_id, status, created_at 
FROM orders 
WHERE user_id = 'uuid' AND status = 'pending'
LIMIT 10;
```

### Explain Analyze

Use `EXPLAIN ANALYZE` to identify slow queries:

```sql
EXPLAIN ANALYZE
SELECT o.id, g.name, d.status 
FROM orders o
JOIN goods g ON o.good_id = g.id
LEFT JOIN deliveries d ON o.id = d.order_id
WHERE o.user_id = 'uuid';
```

## Backup and Recovery

### Backup Strategy

**Daily Full Backups**:
```bash
pg_dump -U postgres -h localhost -d skypost_delivery -F c -f backup_$(date +%Y%m%d).dump
```

**Continuous WAL Archiving**:
```ini
# postgresql.conf
wal_level = replica
archive_mode = on
archive_command = 'cp %p /var/lib/postgresql/wal_archive/%f'
```

**Backup Retention**: 30 days

### Point-in-Time Recovery (PITR)

Restore to specific timestamp:
```bash
# Stop PostgreSQL
systemctl stop postgresql

# Restore base backup
pg_restore -U postgres -d skypost_delivery backup_20240115.dump

# Apply WAL files up to target time
# Edit recovery.conf
recovery_target_time = '2024-01-15 12:00:00'

# Start PostgreSQL
systemctl start postgresql
```

### Disaster Recovery Plan

1. **Primary Failure**: Promote standby replica to master
2. **Data Corruption**: Restore from last known good backup
3. **Hardware Failure**: Migrate to new server using PITR
4. **Human Error**: Rollback to pre-error backup

## Database Configuration

### PostgreSQL Settings (postgresql.conf)

```ini
# Connection Settings
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200

# Logging
log_statement = 'all'
log_duration = on
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '

# Performance
work_mem = 4MB
maintenance_work_mem = 64MB
autovacuum = on
```

### Connection Pooling (Application)

**Go Configuration**:
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

**PgBouncer** (Optional):
```ini
[databases]
skypost_delivery = host=localhost port=5432 dbname=skypost_delivery

[pgbouncer]
pool_mode = transaction
max_client_conn = 100
default_pool_size = 25
```

## Security Considerations

### Database Access Control

**User Roles**:
- `skypost_admin`: Full access (DDL + DML)
- `skypost_app`: Application user (DML only)
- `skypost_readonly`: Read-only access

```sql
-- Create application user
CREATE USER skypost_app WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE skypost_delivery TO skypost_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO skypost_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO skypost_app;

-- Create read-only user
CREATE USER skypost_readonly WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE skypost_delivery TO skypost_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO skypost_readonly;
```

### Data Encryption

**At Rest**: PostgreSQL data directory encryption (LUKS)  
**In Transit**: SSL/TLS connections  
**Password Storage**: Bcrypt hashing with cost factor 10

**SSL Configuration**:
```ini
ssl = on
ssl_cert_file = '/etc/ssl/certs/server.crt'
ssl_key_file = '/etc/ssl/private/server.key'
ssl_ca_file = '/etc/ssl/certs/ca.crt'
```

### Sensitive Data

**Passwords**: Never stored in plaintext, always bcrypt hashed  
**Verification Codes**: Automatically deleted after verification  
**QR Codes**: UUID-based, expire after 24 hours  
**FCM Tokens**: Encrypted in transit, stored securely

## Monitoring and Maintenance

### Database Metrics

**Key Metrics**:
- Connection count: `SELECT count(*) FROM pg_stat_activity;`
- Slow queries: `SELECT * FROM pg_stat_statements ORDER BY total_time DESC;`
- Table sizes: `SELECT pg_size_pretty(pg_total_relation_size('orders'));`
- Index usage: `SELECT * FROM pg_stat_user_indexes;`

**Prometheus Exporter**:
```yaml
# postgres-exporter configuration
DATA_SOURCE_NAME: "postgresql://user:pass@localhost:5432/skypost_delivery?sslmode=disable"
```

**Collected Metrics**:
- `pg_up`: Database availability
- `pg_stat_database_numbackends`: Connection count
- `pg_stat_database_tup_inserted`: Insert rate
- `pg_database_size_bytes`: Database size

### Maintenance Tasks

**Vacuum**: Reclaim storage and update statistics
```sql
VACUUM ANALYZE orders;
```

**Reindex**: Rebuild indexes for performance
```sql
REINDEX TABLE orders;
```

**Autovacuum Configuration**:
```ini
autovacuum = on
autovacuum_naptime = 1min
autovacuum_vacuum_threshold = 50
autovacuum_analyze_threshold = 50
```

## Sample Data

### Seed Data Script

```sql
-- Insert test users
INSERT INTO users (full_name, email, phone_number, phone_verified, role)
VALUES 
    ('Admin User', 'admin@skypost.com', '+79991234567', true, 'admin'),
    ('Test User', 'user@example.com', '+79997654321', true, 'client');

-- Insert test goods
INSERT INTO goods (name, weight, height, length, width, quantity_available)
VALUES 
    ('Small Package', 0.5, 10, 20, 15, 100),
    ('Medium Box', 2.0, 20, 30, 25, 50),
    ('Large Parcel', 5.0, 30, 40, 35, 25);

-- Insert test drones
INSERT INTO drones (model, ip_address, status, battery_level)
VALUES 
    ('Clover 4', '192.168.1.100', 'idle', 95.0),
    ('Clover 4', '192.168.1.101', 'idle', 88.5);

-- Insert test parcel automat
INSERT INTO parcel_automats (city, address, number_of_cells, ip_address, coordinates, aruco_id)
VALUES 
    ('Moscow', 'Red Square, 1', 20, '192.168.1.50', '55.7558,37.6173', 101);

-- Insert locker cells
INSERT INTO locker_cells_out (post_id, height, length, width, status, cell_number)
SELECT id, 30, 40, 35, 'available', generate_series(1, 10)
FROM parcel_automats LIMIT 1;

INSERT INTO locker_cells_internal (post_id, height, length, width, status, cell_number)
SELECT id, 30, 40, 35, 'available', generate_series(11, 20)
FROM parcel_automats LIMIT 1;
```

## Conclusion

This database schema provides:
- **Normalization**: 3NF compliance for data integrity
- **Referential Integrity**: Foreign key constraints
- **Performance**: Strategic indexing
- **Scalability**: UUID-based primary keys
- **Auditability**: Timestamp columns
- **Flexibility**: JSON fields for extensibility
- **Security**: Role-based access control
- **Maintainability**: Migration versioning
