-- name: CreateInternalLockerCell :one
INSERT INTO locker_cells_internal (
    post_id,
    height,
    length,
    width,
    status,
    cell_number
  )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: GetInternalLockerCellByID :one
SELECT *
FROM locker_cells_internal
WHERE id = $1;
-- name: ListInternalLockerCells :many
SELECT *
FROM locker_cells_internal
ORDER BY id;
-- name: ListInternalLockerCellsByPostID :many
SELECT *
FROM locker_cells_internal
WHERE post_id = $1
ORDER BY cell_number;
-- name: UpdateInternalLockerCellStatus :one
UPDATE locker_cells_internal
SET status = $2
WHERE id = $1
RETURNING *;
-- name: UpdateInternalLockerCellDimensions :one
UPDATE locker_cells_internal
SET height = $2,
  length = $3,
  width = $4
WHERE id = $1
RETURNING *;
-- name: FindAvailableInternalCell :one
SELECT *
FROM locker_cells_internal
WHERE status = 'available'
  AND height >= $1
  AND length >= $2
  AND width >= $3
ORDER BY (height * length * width)
LIMIT 1;
-- name: DeleteInternalLockerCell :exec
DELETE FROM locker_cells_internal
WHERE id = $1;