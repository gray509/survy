-- +goose Up
CREATE TABLE answer(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    body TEXT NOT NULL,
    questions_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    response_id UUID NOT NULL REFERENCES response(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE answer;