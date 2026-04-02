-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: ResetAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: SetRefreshTokenById :exec
UPDATE users
SET updated_at = NOW(), refresh_token = $1
WHERE id = $2;

-- name: UserExit :one
SELECT EXISTS (
  SELECT 1
  FROM users
  WHERE id = $1
);