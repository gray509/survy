-- name: CreateResponse :exec
INSERT INTO responses (id, created_at, updated_at, response, survey_id, voter_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
);