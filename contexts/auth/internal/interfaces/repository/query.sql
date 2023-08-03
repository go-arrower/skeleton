---------------------
------ Session ------
---------------------

-- name: AllSessions :many
SELECT *
FROM auth.session
ORDER BY created_at ASC;

-- name: FindSessionDataByKey :one
SELECT data
FROM auth.session
WHERE key = $1;

-- name: DeleteSessionByKey :exec
DELETE
FROM auth.session
WHERE key = $1;

-- name: UpsertSession :exec
INSERT INTO auth.session (key, data, expires_at, user_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (key) DO UPDATE SET data       = $2,
                                expires_at = $3,
                                user_id    = $4;



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
INSERT INTO auth.user (login, password_hash, verified_at, blocked_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- DeleteUser :exec
DELETE
FROM auth.user
WHERE id = $1;