-- name: BulkEnterQuestions :copyfrom
INSERT INTO questions (id, created_at, updated_at, title, question_type, choice, is_required, survey_id)
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

-- name: GetQuestionsFromSurveyId :many
SELECT id, created_at, updated_at, title, question_type, choice, is_required
FROM questions
WHERE survey_id = $1;
