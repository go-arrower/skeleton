---------------------
------ Session ------
---------------------

-- name: AllSessions :many
SELECT *
FROM auth.session
ORDER BY created_at ASC;

-- name: FindSessionsByUserID :many
SELECT *
FROM auth.session
WHERE user_id = $1
ORDER BY created_at;

-- name: FindSessionDataByKey :one
SELECT data
FROM auth.session
WHERE key = $1;

-- name: DeleteSessionByKey :exec
DELETE
FROM auth.session
WHERE key = $1;

-- name: UpsertSessionData :exec
INSERT INTO auth.session (key, data, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE SET (data, expires_at) = ($2, $3);

-- name: UpsertNewSession :exec
INSERT INTO auth.session (key, user_id, user_agent)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE SET (user_id, user_agent) = ($2, $3);



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
INSERT INTO auth.user (id, login, password_hash, verified_at, blocked_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- DeleteUser :exec
DELETE
FROM auth.user
WHERE id = $1;