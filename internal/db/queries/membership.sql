-- Membership carries the role, so granting access and choosing what it permits
-- are one operation.

-- name: GrantMembership :exec
INSERT INTO employee_companies (employee_id, company_id, account_id, role_id)
SELECT sqlc.arg(employee_id), c.id, c.account_id, sqlc.narg(role_id)
FROM company c
WHERE c.id = sqlc.arg(company_id)
ON CONFLICT (employee_id, company_id)
DO UPDATE SET role_id = EXCLUDED.role_id;

-- account_id is derived from the company rather than taken as a parameter, so it
-- cannot disagree with the composite foreign key.

-- name: RevokeMembershipsNotIn :exec
DELETE FROM employee_companies
WHERE employee_id = sqlc.arg(employee_id)
  AND NOT (company_id = ANY(sqlc.arg(company_ids)::bigint[]));

-- name: ListAccountEmployeeMemberships :many
SELECT e.id AS employee_id, e.first_name, e.last_name, e.email, e.is_active,
       ec.company_id, c.name AS company_name,
       ec.role_id, r.name AS role_name, COALESCE(r.is_admin, false) AS role_is_admin
FROM employee e
LEFT JOIN employee_companies ec ON ec.employee_id = e.id
LEFT JOIN company c ON c.id = ec.company_id
LEFT JOIN role r    ON r.id = ec.role_id
WHERE e.account_id = sqlc.arg(account_id)::bigint
ORDER BY e.id, ec.company_id;

-- name: RoleBelongsToCompany :one
SELECT EXISTS (
    SELECT 1 FROM role r
    WHERE r.id = sqlc.arg(role_id) AND r.company_id = sqlc.arg(company_id)
);

-- name: CompanyInAccount :one
SELECT EXISTS (
    SELECT 1 FROM company c
    WHERE c.id = sqlc.arg(company_id) AND c.account_id = sqlc.arg(account_id)
);
