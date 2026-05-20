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
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    title TEXT NOT NULL,
    expiration_time TIMESTAMPTZ,
    max_response INTEGER,
    is_published BOOLEAN DEFAULT FALSE NOT NULL,
    questions JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE voters(
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    ip TEXT,
    hash TEXT
);

CREATE TABLE responses(
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    response JSONB NOT NULL,
    voter_id UUID NOT NULL REFERENCES voters(id) ON DELETE CASCADE,
    survey_id UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE
);



CREATE TABLE refresh_tokens(
    id UUID PRIMARY KEY,
    token_hash TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE users CASCADE;
DROP TABLE surveys CASCADE;
DROP TABLE responses CASCADE;
DROP TABLE voters CASCADE;
DROP TABLE refresh_tokens;