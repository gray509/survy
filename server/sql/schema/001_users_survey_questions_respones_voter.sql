-- +goose Up
CREATE TABLE users(
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    refresh_token TEXT
);

CREATE TABLE surveys(
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    title TEXT NOT NULL,
    --user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expiration_time TIMESTAMPTZ,
    indentified BOOLEAN NOT NULL,
    max_response INTEGER,
    is_published BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE TABLE questions(
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    title TEXT NOT NULL,
    --surveys_id UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
    options JSONB DEFAULT '{}'::jsonb,
    is_required BOOLEAN NOT NUll,
    types TEXT NOT NULL
);

CREATE TABLE responses(
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    response JSONB NOT NULL
    --voter_id UUID NOT NULL REFERENCES voter(id) ON DELETE CASCADE,
    --questions_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    --surveys_id UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE
);

CREATE TABLE voter(
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    ip TEXT,
    hash TEXT
);

-- +goose Down
DROP TABLE users;
DROP TABLE surveys;
DROP TABLE questions;
DROP TABLE responses;
DROP TABLE voter;