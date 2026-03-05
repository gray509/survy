-- +goose Up
CREATE TABLE polls(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    questions_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    is_public BOOLEAN NOT NULL,
    expiration TIMESTAMP
);

-- +goose Down
DROP TABLE polls;