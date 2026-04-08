-- name: CreateQuestion :one
INSERT INTO questions (id, created_at, updated_at, title, types, is_required ,polls_id, options)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING id;

-- name: GetQuestionsWithPollid :many
select * 
from questions 
where polls_id = $1;
