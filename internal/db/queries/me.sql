-- Queries backing /api/v1/me/permissions. The authorization middleware already
-- resolves the caller (GetEmployeeIdentity in employee.sql); these add the
-- descriptive fields a client needs to render the shell: who am I, what is my
-- role called, and which companies can I switch to.

-- name: GetMeProfile :one
SELECT
    e.id,
    e.first_name,
    e.last_name,
    e.email,
    e.job_title,
    e.is_active,
    e.is_technician,
    e.is_vehicle_operator,
    e.is_account_owner,
    e.default_company_id,
    e.role_id,
    r.name                               AS role_name,
    COALESCE(r.is_admin, false)          AS role_is_admin,
    COALESCE(r.permissions, '{}'::jsonb) AS permissions
FROM employee e
LEFT JOIN role r ON r.id = e.role_id
WHERE e.id = sqlc.arg(id);

-- ListMyCompanies returns every company the caller belongs to, unpaginated: a
-- company switcher needs the whole set in one request, and membership counts are
-- bounded by how many companies a person actually works for.
-- name: ListMyCompanies :many
SELECT c.id, c.name, c.logo
FROM company c
JOIN employee_companies ec ON ec.company_id = c.id
WHERE ec.employee_id = sqlc.arg(employee_id)
ORDER BY c.name, c.id;
