-- name: CreateUser :one
INSERT INTO auth.users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM auth.users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM auth.users
WHERE id = $1;
