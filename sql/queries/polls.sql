-- name: CreatePoll :one
INSERT INTO polls (id, created_at, updated_at, title, questions_id, user_id, config)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3,
    $4
)
RETURNING *;
