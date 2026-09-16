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
    city, region, postal_code, country, timezone, currency, system_of_measurement, account_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, sqlc.arg(account_id)
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

-- name: BootstrapEmployeeDefaultCompany :exec
-- Mirrors Django: only fills default_company when the employee has none. The
-- role half moved to the membership.
UPDATE employee SET
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
-- A company's owner is the owner of the account it belongs to. Ownership lives
-- on account.owner_employee_id, not on the employee row, so an account has
-- exactly one owner and it is the same for every company in it. No row when
-- the account has no owner.
SELECT e.id, e.first_name, e.last_name, e.email, e.job_title
FROM company c
JOIN account a  ON a.id = c.account_id
JOIN employee e ON e.id = a.owner_employee_id
WHERE c.id = sqlc.arg(company_id);

-- name: ListAdminCompanies :many
-- Cross-tenant company list for the admin namespace. ListCompanies is scoped to
-- the caller's memberships, and platform staff hold none. account_id narrows it
-- to one client; a null argument means every client.
SELECT sqlc.embed(c),
    a.name AS account_name,
    (SELECT count(*) FROM employee_companies ec WHERE ec.company_id = c.id) AS employee_count
FROM company c
JOIN account a ON a.id = c.account_id
WHERE sqlc.narg(account_id)::bigint IS NULL OR c.account_id = sqlc.narg(account_id)
ORDER BY c.name, c.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountAdminCompanies :one
SELECT count(*) FROM company c
WHERE sqlc.narg(account_id)::bigint IS NULL OR c.account_id = sqlc.narg(account_id);

-- name: GetAdminCompany :one
-- Single-company read with the same account fields as ListAdminCompanies.
SELECT sqlc.embed(c),
    a.name AS account_name,
    (SELECT count(*) FROM employee_companies ec WHERE ec.company_id = c.id) AS employee_count
FROM company c
JOIN account a ON a.id = c.account_id
WHERE c.id = sqlc.arg(id);
