-- name: CreateLockerCell :one
INSERT INTO locker_cells_out (
    post_id,
    height,
    length,
    width,
    status,
    cell_number
  )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: GetLockerCellByID :one
SELECT *
FROM locker_cells_out
WHERE id = $1;
-- name: ListLockerCells :many
SELECT *
FROM locker_cells_out
ORDER BY id;
-- name: ListLockerCellsByPostID :many
SELECT *
FROM locker_cells_out
WHERE post_id = $1
ORDER BY cell_number;
-- name: UpdateLockerCellStatus :one
UPDATE locker_cells_out
SET status = $2
WHERE id = $1
RETURNING *;
-- name: UpdateLockerCellDimensions :one
UPDATE locker_cells_out
SET height = $2,
  length = $3,
  width = $4
WHERE id = $1
RETURNING *;
-- name: FindAvailableCell :one
SELECT *
FROM locker_cells_out
WHERE status = 'available'
  AND height >= $1
  AND length >= $2
  AND width >= $3
ORDER BY (height * length * width)
LIMIT 1;
-- name: DeleteLockerCell :exec
DELETE FROM locker_cells_out
WHERE id = $1;