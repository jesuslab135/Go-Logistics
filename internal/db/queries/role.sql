-- name: GetRole :one
SELECT * FROM role WHERE id = $1 AND company_id = $2;

-- name: ListRoles :many
SELECT * FROM role WHERE company_id = $1 ORDER BY name LIMIT $2 OFFSET $3;

-- name: CountRoles :one
SELECT count(*) FROM role WHERE company_id = $1;

-- name: FindRoleByName :one
SELECT * FROM role WHERE company_id = $1 AND name = $2 LIMIT 1;

-- name: CreateRole :one
INSERT INTO role (
    company_id, name, is_admin, permissions
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateRole :one
UPDATE role SET name = $3, is_admin = $4, permissions = $5
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM role WHERE id = $1 AND company_id = $2;
