-- Queries backing /api/v1/me/permissions. The authorization middleware already
-- resolves the caller (GetEmployeeIdentity in employee.sql); these add the
-- descriptive fields a client needs to render the shell: who am I, what is my
-- role called, and which companies can I switch to.

-- GetMeProfile's role now comes through the membership for the caller's
-- current company, the same join GetEmployeeIdentity uses. company_id is
-- nullable for the same reason: an account owner with no company yet has no
-- membership row to resolve a role from.
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
    e.default_company_id,
    ec.role_id,
    r.name                               AS role_name,
    COALESCE(r.is_admin, false)          AS role_is_admin,
    COALESCE(r.permissions, '{}'::jsonb) AS permissions
FROM employee e
LEFT JOIN employee_companies ec ON ec.employee_id = e.id
                               AND ec.company_id = sqlc.narg(company_id)
                               AND ec.is_active
LEFT JOIN role r ON r.id = ec.role_id
WHERE e.id = sqlc.arg(id);

-- ListMyCompanies returns every company the caller belongs to, unpaginated: a
-- company switcher needs the whole set in one request, and membership counts are
-- bounded by how many companies a person actually works for.
-- name: ListMyCompanies :many
SELECT c.id, c.name, c.logo
FROM company c
JOIN employee_companies ec ON ec.company_id = c.id
WHERE ec.employee_id = sqlc.arg(employee_id)
  AND ec.is_active
ORDER BY c.name, c.id;
