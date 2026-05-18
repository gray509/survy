-- +goose Up
CREATE TABLE questions(
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    survey_id UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    question_type TEXT NOT NULL,
    is_required BOOLEAN NOT NULL,
    choice TEXT[]
);

ALTER TABLE surveys
DROP COLUMN questions;

ALTER TABLE responses
ADD COLUMN question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE;

ALTER TABLE responses
ADD COLUMN choice JSONB NOT NULL DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE surveys
ADD COLUMN questions JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE responses
DROP COLUMN question_id;

ALTER TABLE responses
DROP COLUMN choice;

DROP TABLE questions;

