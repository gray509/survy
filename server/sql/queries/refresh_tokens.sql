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
SELECT users.*
FROM refresh_tokens JOIN users ON users.id = refresh_token.user_id
WHERE refresh_tokens.token_hash = $1
AND refresh_tokens.revoked_at = NUll
AND refresh_tokens.expires_at > NOW();