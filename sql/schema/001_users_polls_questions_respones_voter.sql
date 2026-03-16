-- +goose Up
CREATE TABLE users(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL
);

CREATE TABLE polls(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    --questions_id UUID NOT NULL REFERENCES questions(id),
    --user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    config JSONB
);

CREATE TABLE questions(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    --polls_id UUID NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    questions JSONB
);

CREATE TABLE responses(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    response JSONB NOT NULL
    --voter_id UUID NOT NULL REFERENCES voter(id) ON DELETE CASCADE,
    --questions_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    --polls_id UUID NOT NULL REFERENCES polls(id) ON DELETE CASCADE
);

CREATE TABLE voter(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    ip TEXT,
    hash TEXT
);

-- +goose Down
DROP TABLE users;
DROP TABLE polls;
DROP TABLE questions;
DROP TABLE responses;
DROP TABLE voter;