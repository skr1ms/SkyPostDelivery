-- name: SaveDeliveryTask :exec
UPDATE deliveries
SET drone_id = $2,
    status = $3,
    started_at = COALESCE(started_at, CURRENT_TIMESTAMP)
WHERE id = $1;
-- name: GetDeliveryTask :one
SELECT d.id,
    d.order_id,
    o.good_id,
    o.locker_cell_id,
    d.parcel_automat_id,
    d.internal_locker_cell_id,
    d.status,
    d.started_at,
    d.completed_at,
    d.drone_id
FROM deliveries d
    JOIN orders o ON d.order_id = o.id
WHERE d.id = $1;
-- name: UpdateDeliveryStatus :exec
UPDATE deliveries
SET status = $2,
    completed_at = CASE
        WHEN $2 = 'completed' THEN CURRENT_TIMESTAMP
        ELSE completed_at
    END
WHERE id = $1;