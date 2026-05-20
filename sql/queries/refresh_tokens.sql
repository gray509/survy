-- name: AddRefreshToken :exec
INSERT INTO refresh_tokens (id, token_hash, user_id, created_at, expires_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token_hash = $1;

-- name: GetUserFromRefreshToken :one
SELECT user_id
FROM refresh_tokens
WHERE token_hash = $1
AND revoked_at is NULL
AND expires_at > NOW();