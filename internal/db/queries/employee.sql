-- name: GetEmployeeAuthByEmail :one
SELECT e.id, e.default_company_id, e.password_hash, COALESCE(r.is_admin, false) AS is_admin
FROM employee e
LEFT JOIN role r ON r.id = e.role_id
WHERE e.email = $1 AND e.is_active = true
LIMIT 1;

-- name: GetEmployee :one
SELECT * FROM employee WHERE id = $1;

-- name: UpdateEmployeePassword :exec
UPDATE employee SET password_hash = $2, updated_at = $3 WHERE id = $1;
