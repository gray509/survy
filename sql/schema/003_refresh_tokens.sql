-- +goose Up
CREATE TABLE refresh_tokens(
    id UUID PRIMARY KEY,
    token_hash TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

ALTER TABLE users
DROP COLUMN refresh_token;

-- +goose Down
DROP TABLE refresh_tokens;

ALTER TABLE users
ADD COLUMN refresh_token TEXT;