-- name: CreateParcelAutomat :one
INSERT INTO parcel_automats (
        city,
        address,
        number_of_cells,
        ip_address,
        coordinates,
        aruco_id,
        is_working
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
-- name: GetParcelAutomatByID :one
SELECT *
FROM parcel_automats
WHERE id = $1;
-- name: ListParcelAutomats :many
SELECT *
FROM parcel_automats
ORDER BY id;
-- name: DeleteParcelAutomat :exec
DELETE FROM parcel_automats
WHERE id = $1;
-- name: UpdateParcelAutomatStatus :one
UPDATE parcel_automats
SET is_working = $2
WHERE id = $1
RETURNING *;
-- name: UpdateParcelAutomat :one
UPDATE parcel_automats
SET city = $2,
    address = $3,
    ip_address = $4,
    coordinates = $5
WHERE id = $1
RETURNING *;
-- name: ListWorkingParcelAutomats :many
SELECT *
FROM parcel_automats
WHERE is_working = true
ORDER BY city,
    address;