-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: BulkCreateUser :copyfrom
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);

-- name: ResetAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UserExist :one
SELECT EXISTS (
  SELECT 1
  FROM users
  WHERE email = $1
);

-- name: DeleteTestUsers :exec
DELETE FROM users WHERE email LIKE 'user%@testsurvy.com';