------------------
----- Tenant -----
------------------

-- name: AllTenants :many
SELECT *
FROM auth.tenant
ORDER BY name;

-- name: CreateTenant :exec
INSERT INTO auth.tenant (name)
VALUES ($1);

-- name: FindTenantByID :one
SELECT *
FROM auth.tenant
WHERE id = $1;



------------------
------ User ------
------------------

-- name: AllUsers :many
SELECT *
FROM auth.user
WHERE credential_type = 'user'
ORDER BY is_admin, user_login;

-- name: FindUserByID :one
SELECT *
FROM auth.user
WHERE id = $1;

-- name: FindUserByLogin :one
SELECT *
FROM auth.user
WHERE credential_type = 'user'
  AND user_login = $1;

-- name: CreateUser :one
INSERT INTO auth.user (credential_type, user_login, user_password_hash)
VALUES ('user', $1, $2)
RETURNING *;

-- name: UpsertUser :one
INSERT INTO auth.user (credential_type, user_login, user_password_hash)
VALUES ('user', $1, $2)
ON CONFLICT (user_login) DO UPDATE SET is_active              = $3,
                                       user_password_hash     = $4,
                                       user_login_verified_at = $5,
                                       is_admin               = $6
RETURNING *;

-- DeleteUser :exec
DELETE
FROM auth.user
WHERE id = $1;