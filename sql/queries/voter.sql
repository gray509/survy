-- name: CreateVoter :one
INSERT INTO voter (id, created_at, updated_at, ip, hash)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;