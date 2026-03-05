-- +goose Up
CREATE TABLE voter(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    ip TEXT,
    hash TEXT
);

-- +goose Down
DROP TABLE voter;