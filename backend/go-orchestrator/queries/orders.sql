-- name: CreateOrder :one
INSERT INTO orders (
        user_id,
        good_id,
        parcel_automat_id,
        locker_cell_id,
        status
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
-- name: GetOrderByID :one
SELECT *
FROM orders
WHERE id = $1;
-- name: GetOrderByLockerCellID :one
SELECT *
FROM orders
WHERE locker_cell_id = $1;
-- name: ListOrdersByUserID :many
SELECT o.id,
    o.user_id,
    o.good_id,
    o.parcel_automat_id,
    o.locker_cell_id,
    o.status,
    o.created_at,
    g.id as "good.id",
    g.name as "good.name",
    g.weight as "good.weight",
    g.height as "good.height",
    g.length as "good.length",
    g.width as "good.width",
    g.quantity_available as "good.quantity_available"
FROM orders o
    LEFT JOIN goods g ON o.good_id = g.id
WHERE o.user_id = $1
ORDER BY o.created_at DESC;
-- name: ListOrders :many
SELECT *
FROM orders
ORDER BY created_at DESC;
-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2
WHERE id = $1
RETURNING *;
-- name: DeleteOrder :exec
DELETE FROM orders
WHERE id = $1;