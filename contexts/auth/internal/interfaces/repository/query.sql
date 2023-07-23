-- name: CreateTenant :exec
INSERT INTO auth.tenant (name)
VALUES ($1);

-- name: AllTenants :many
SELECT *
FROM auth.tenant
ORDER BY name;

-- name: FindTenantByID :one
SELECT *
FROM auth.tenant
WHERE id = $1;