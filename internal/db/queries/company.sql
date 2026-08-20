-- Companies are scoped by the caller's employee_companies membership, not by the
-- JWT's single company_id: an employee may belong to several companies and must
-- see all of them. Django left this queryset unfiltered, which let any company
-- admin read and DELETE other tenants' companies (cascading to their data).

-- name: GetCompany :one
SELECT c.* FROM company c
WHERE c.id = sqlc.arg(id)
  AND EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.company_id = c.id AND ec.employee_id = sqlc.arg(employee_id));

-- name: ListCompanies :many
SELECT c.* FROM company c
WHERE EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.company_id = c.id AND ec.employee_id = sqlc.arg(employee_id))
ORDER BY c.name, c.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountCompanies :one
SELECT count(*) FROM company c
WHERE EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.company_id = c.id AND ec.employee_id = sqlc.arg(employee_id));

-- name: CreateCompany :one
INSERT INTO company (
    name, tax_id, address, created_at, phone, email, website, logo,
    city, region, postal_code, country, timezone, currency, system_of_measurement
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: UpdateCompany :one
UPDATE company c SET
    name = sqlc.arg(name), tax_id = sqlc.arg(tax_id), address = sqlc.arg(address),
    phone = sqlc.arg(phone), email = sqlc.arg(email), website = sqlc.arg(website),
    logo = sqlc.arg(logo), city = sqlc.arg(city), region = sqlc.arg(region),
    postal_code = sqlc.arg(postal_code), country = sqlc.arg(country),
    timezone = sqlc.arg(timezone), currency = sqlc.arg(currency),
    system_of_measurement = sqlc.arg(system_of_measurement)
WHERE c.id = sqlc.arg(id)
  AND EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.company_id = c.id AND ec.employee_id = sqlc.arg(employee_id))
RETURNING c.*;

-- name: DeleteCompany :exec
DELETE FROM company c
WHERE c.id = sqlc.arg(id)
  AND EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.company_id = c.id AND ec.employee_id = sqlc.arg(employee_id));

-- Company bootstrap. Django's CompanyViewSet.perform_create ran these in one
-- transaction with the insert, so a company is never created unusable. The
-- ON CONFLICT clauses make each seed idempotent, matching get_or_create.

-- name: SeedAssetStatus :exec
INSERT INTO asset_status (company_id, name, color_code)
VALUES (sqlc.arg(company_id), sqlc.arg(name), sqlc.arg(color_code))
ON CONFLICT (company_id, name) DO NOTHING;

-- name: SeedWorkOrderStatus :exec
INSERT INTO work_order_status (
    company_id, name, description, color, is_default, marks_as_completed, position
) VALUES (
    sqlc.arg(company_id), sqlc.arg(name), '', sqlc.arg(color),
    sqlc.arg(is_default), sqlc.arg(marks_as_completed), sqlc.arg(position)
)
ON CONFLICT (company_id, name) DO NOTHING;

-- name: BootstrapEmployeeCompany :exec
-- Mirrors Django: only fills role/default_company when the employee has none.
UPDATE employee SET
    role_id            = COALESCE(role_id, sqlc.arg(role_id)),
    default_company_id = COALESCE(default_company_id, sqlc.arg(company_id)),
    updated_at         = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: CountCompaniesByIDs :one
-- Pre-flight for a membership replace: a mismatch against the requested count
-- means at least one id names no company, which is a 422 rather than the 409 a
-- foreign-key violation would surface as.
SELECT count(*) FROM company WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: GetCompanyByID :one
-- Unscoped single-company read for the admin namespace.
SELECT * FROM company WHERE id = sqlc.arg(id);

-- name: GetCompanyOwner :one
-- employee.is_account_owner is a single global boolean, so a company's owner is
-- the member of that company carrying the flag. LIMIT 1 keeps the read total
-- even where historical data has more than one — SetCompanyAccountOwner is what
-- makes that impossible going forward.
SELECT e.id, e.first_name, e.last_name, e.email, e.job_title
FROM employee e
JOIN employee_companies ec ON ec.employee_id = e.id AND ec.company_id = sqlc.arg(company_id)
WHERE e.is_account_owner = true
ORDER BY e.id
LIMIT 1;

-- name: ClearCompanyAccountOwner :exec
-- The clear half of set-owner. Runs in the same transaction as the set so a
-- company never has two owners, which plain CRUD on is_account_owner allows.
UPDATE employee e SET is_account_owner = false, updated_at = sqlc.arg(updated_at)
FROM employee_companies ec
WHERE ec.employee_id = e.id
  AND ec.company_id = sqlc.arg(company_id)
  AND e.is_account_owner = true;

-- name: SetCompanyAccountOwner :one
-- The membership EXISTS guard makes "not a member of this company" return no
-- row, so the caller cannot make an outsider the owner of a tenant.
UPDATE employee SET is_account_owner = true, updated_at = sqlc.arg(updated_at)
WHERE employee.id = sqlc.arg(id)
  AND EXISTS (
      SELECT 1 FROM employee_companies ec
      WHERE ec.employee_id = employee.id AND ec.company_id = sqlc.arg(company_id)
  )
RETURNING id, first_name, last_name, email, job_title;
