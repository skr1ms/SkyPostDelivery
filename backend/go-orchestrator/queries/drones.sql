-- name: CreateDrone :one
INSERT INTO drones (model, status, ip_address)
VALUES ($1, $2, $3)
RETURNING *;
-- name: GetDroneByID :one
SELECT *
FROM drones
WHERE id = $1;
-- name: ListDrones :many
SELECT *
FROM drones
ORDER BY id;
-- name: UpdateDroneStatus :one
UPDATE drones
SET status = $2
WHERE id = $1
RETURNING *;
-- name: UpdateDrone :one
UPDATE drones
SET model = $2,
    ip_address = $3,
    status = $4
WHERE id = $1
RETURNING *;
-- name: GetAvailableDrone :one
SELECT *
FROM drones
WHERE status = 'idle'
LIMIT 1;
-- name: DeleteDrone :exec
DELETE FROM drones
WHERE id = $1;