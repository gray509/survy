-- name: CreateQuestion :one
INSERT INTO questions (id, created_at, updated_at, title, types, is_required ,surveys_id, options)
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

-- name: GetQuestionsWithSurveyid :many
select * 
from questions 
where surveys_id = $1;

-- name: QuestionsBulkInsert :copyfrom
INSERT INTO questions (id, created_at, updated_at, title, types, is_required ,surveys_id, options)
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

-- name: GetQuestionBySurveyId :many
Select *
FROM questions
WHERE surveys_id = $1;