-- name: GetDroneIDByIP :one
SELECT id FROM drones
WHERE ip_address = $1;

-- name: UpdateDroneBattery :exec
UPDATE drones
SET battery_level = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: SaveDroneState :exec
UPDATE drones
SET 
    status = $2,
    battery_level = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetDroneState :one
SELECT id, status, battery_level
FROM drones
WHERE id = $1;

