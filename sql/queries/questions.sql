-- name: CreateQuestion :one
INSERT INTO questions (id, created_at, updated_at, title, types, is_required ,polls_id, options)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING id;

-- name: GetQuestionsWithPollid :many
select * 
from questions 
where polls_id = $1;

-- name: QuestionsBulkInsert :copyfrom
INSERT INTO questions (id, created_at, updated_at, title, types, is_required ,polls_id, options)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
);
