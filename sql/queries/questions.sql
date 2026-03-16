-- name: CreateQuestion :one
INSERT INTO questions (id, created_at, updated_at, title, polls_id, questions)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3
)
RETURNING *;