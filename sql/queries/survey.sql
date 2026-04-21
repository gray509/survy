-- name: CreateSurvey :one
INSERT INTO surveys (id, created_at, updated_at, title, user_id, expiration_time, indentified, max_response)
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

-- name: BulkCreateSurvey :copyfrom
INSERT INTO surveys (id, created_at, updated_at, title, user_id, expiration_time, indentified, max_response)
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
-- name: GetSurveyByIdUserId :one
Select *
FROM surveys
WHERE surveys.id = $1 AND surveys.user_id = $2;

-- name: SetIsPublish :execresult
UPDATE surveys
SET is_published = $1
WHERE id = $2 AND user_id = $3;


