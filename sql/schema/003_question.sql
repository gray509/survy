-- +goose Up
CREATE TABLE questions(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    polls_id UUID NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    answer_id UUID NOT NULL REFERENCES answer(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE questions;