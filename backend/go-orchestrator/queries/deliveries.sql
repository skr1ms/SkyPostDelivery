-- name: CreateDelivery :one
INSERT INTO deliveries (
        order_id,
        drone_id,
        parcel_automat_id,
        internal_locker_cell_id,
        status
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
-- name: GetDeliveryByID :one
SELECT *
FROM deliveries
WHERE id = $1;
-- name: GetDeliveryByOrderID :one
SELECT *
FROM deliveries
WHERE order_id = $1;
-- name: ListDeliveries :many
SELECT *
FROM deliveries
ORDER BY id DESC;
-- name: ListDeliveriesByStatus :many
SELECT *
FROM deliveries
WHERE status = $1
ORDER BY id DESC;
-- name: UpdateDeliveryStatus :one
UPDATE deliveries
SET status = $2
WHERE id = $1
RETURNING *;
-- name: UpdateDeliveryDrone :one
UPDATE deliveries
SET drone_id = $2
WHERE id = $1
RETURNING *;
-- name: DeleteDelivery :exec
DELETE FROM deliveries
WHERE id = $1;