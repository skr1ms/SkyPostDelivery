-- name: UpsertDevice :one
INSERT INTO user_devices (user_id, token, platform)
VALUES ($1, $2, $3) ON CONFLICT (token) DO
UPDATE
SET user_id = EXCLUDED.user_id,
    platform = EXCLUDED.platform,
    updated_at = NOW()
RETURNING *;
-- name: ListDevicesByUserID :many
SELECT *
FROM user_devices
WHERE user_id = $1;
-- name: DeleteDeviceByToken :exec
DELETE FROM user_devices
WHERE token = $1;