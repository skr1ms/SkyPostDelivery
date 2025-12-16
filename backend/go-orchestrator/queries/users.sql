-- name: CreateUser :one
INSERT INTO users (full_name, email, phone_number, pass_hash, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
-- name: CreateUserWithCustomDate :one
INSERT INTO users (
        full_name,
        email,
        phone_number,
        pass_hash,
        role,
        created_at
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;
-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;
-- name: GetUserByPhone :one
SELECT *
FROM users
WHERE phone_number = $1;
-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at DESC;
-- name: UpdateUser :one
UPDATE users
SET full_name = $2,
    email = $3,
    phone_number = $4
WHERE id = $1
RETURNING *;
-- name: UpdateUserPassword :one
UPDATE users
SET pass_hash = $2
WHERE id = $1
RETURNING *;
-- name: UpdateUserVerificationCode :one
UPDATE users
SET verification_code = $2,
    code_expires_at = $3
WHERE id = $1
RETURNING *;
-- name: VerifyUserPhone :one
UPDATE users
SET phone_verified = true,
    verification_code = NULL,
    code_expires_at = NULL
WHERE id = $1
RETURNING *;
-- name: UpdateUserQR :one
UPDATE users
SET qr_issued_at = $2,
    qr_expires_at = $3
WHERE id = $1
RETURNING *;
-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;