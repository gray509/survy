-- name: CreateVoter :one
INSERT INTO voter (id, created_at, updated_at, ip, hash)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;