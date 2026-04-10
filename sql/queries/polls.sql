-- name: CreateSurvey :one
INSERT INTO surveys (id, created_at, updated_at, title, user_id, config)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING id;

