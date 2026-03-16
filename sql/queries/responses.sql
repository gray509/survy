-- name: CreateResponse :one
INSERT INTO responses (id, created_at, updated_at, response, polls_id, questions_id, voter_id)
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