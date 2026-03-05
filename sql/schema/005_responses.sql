-- +goose Up
CREATE TABLE response(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    response TEXT NOT NULL,
    answer_id UUID NOT NULL REFERENCES answer(id) ON DELETE CASCADE,
    voter_id UUID NOT NULL REFERENCES voter(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE response;