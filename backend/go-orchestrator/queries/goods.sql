-- name: CreateGood :one
INSERT INTO goods (
        name,
        weight,
        height,
        length,
        width,
        quantity_available
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: GetGoodByID :one
SELECT *
FROM goods
WHERE id = $1;
-- name: ListGoods :many
SELECT *
FROM goods
ORDER BY id;
-- name: UpdateGood :one
UPDATE goods
SET name = $2,
    weight = $3,
    height = $4,
    length = $5,
    width = $6
WHERE id = $1
RETURNING *;
-- name: UpdateGoodQuantity :one
UPDATE goods
SET quantity_available = quantity_available + $2
WHERE id = $1
RETURNING *;
-- name: ListAvailableGoods :many
SELECT *
FROM goods
WHERE quantity_available > 0
ORDER BY name;
-- name: DeleteGood :exec
DELETE FROM goods
WHERE id = $1;