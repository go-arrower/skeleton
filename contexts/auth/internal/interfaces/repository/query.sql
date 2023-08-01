------------------
------ User ------
------------------

-- name: AllUsers :many
SELECT *
FROM auth.user
ORDER BY login;

-- name: FindUserByID :one
SELECT *
FROM auth.user
WHERE id = $1;

-- name: FindUserByLogin :one
SELECT *
FROM auth.user
WHERE login = $1;

-- name: CreateUser :one
INSERT INTO auth.user (login, password_hash, verified_at)
VALUES ($1, $2, $3)
RETURNING *;

-- DeleteUser :exec
DELETE
FROM auth.user
WHERE id = $1;