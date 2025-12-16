-- name: GetDroneIDByIP :one
SELECT id
FROM drones
WHERE ip_address = $1;
-- name: UpdateDroneBattery :exec
UPDATE drones
SET battery_level = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
-- name: SaveDroneState :exec
UPDATE drones
SET status = $2,
    battery_level = $3,
    latitude = $4,
    longitude = $5,
    altitude = $6,
    speed = $7,
    current_delivery_id = $8,
    error_message = $9,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
-- name: GetDroneState :one
SELECT id,
    status,
    battery_level,
    latitude,
    longitude,
    altitude,
    speed,
    current_delivery_id,
    error_message,
    updated_at
FROM drones
WHERE id = $1;