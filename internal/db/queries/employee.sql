-- name: GetEmployeeAuthByEmail :one
SELECT e.id, e.default_company_id, e.account_id, e.password_hash
FROM employee e
WHERE e.email = $1 AND e.is_active = true
LIMIT 1;

-- name: UpdateEmployeePassword :exec
UPDATE employee SET password_hash = $2, updated_at = $3 WHERE id = $1;

-- GetEmployeeIdentity backs the authorization middleware, resolved from the
-- database on every request so a role change or deactivation takes effect at
-- once rather than at the next token refresh.
--
-- The role now comes through the membership, so one join answers both "is this
-- person a member of this company" and "what may they do here". company_id is
-- nullable: a company-less session has no tenant, and sqlc.narg makes the join
-- miss rather than requiring a sentinel id.
--
-- Both ownership expressions are COALESCEd because platform staff have
-- account_id IS NULL, so the account join yields no row and the comparison is
-- NULL. In SQL `false OR NULL` is NULL — an authorization value decided by NULL
-- semantics rather than by a rule.
-- name: GetEmployeeIdentity :one
SELECT
    e.id,
    e.is_active,
    e.account_id,
    e.is_platform_admin,
    ec.role_id,
    (ec.id IS NOT NULL)::boolean         AS is_member,
    (COALESCE(r.is_admin, false)
        OR COALESCE(a.owner_employee_id = e.id, false))::boolean AS is_admin,
    COALESCE(r.permissions, '{}'::jsonb) AS permissions,
    COALESCE(a.owner_employee_id = e.id, false)::boolean AS is_account_owner
FROM employee e
LEFT JOIN employee_companies ec ON ec.employee_id = e.id
                               AND ec.company_id = sqlc.narg(company_id)
LEFT JOIN role r    ON r.id = ec.role_id
LEFT JOIN account a ON a.id = e.account_id
WHERE e.id = sqlc.arg(id);

-- Employee CRUD is scoped by company membership via the employee_companies m2m,
-- and never reads/writes password_hash (managed only by the auth queries above).

-- name: GetEmployee :one
SELECT e.* FROM employee e
WHERE e.id = sqlc.arg(id)
  AND EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.employee_id = e.id AND ec.company_id = sqlc.arg(company_id));

-- name: ListEmployees :many
SELECT e.* FROM employee e
WHERE EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.employee_id = e.id AND ec.company_id = sqlc.arg(company_id))
ORDER BY e.last_name, e.first_name, e.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountEmployees :one
SELECT count(*) FROM employee e
WHERE EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.employee_id = e.id AND ec.company_id = sqlc.arg(company_id));

-- name: CreateEmployee :one
-- account_id is a required argument, never a column the caller's JSON body can
-- reach (see dto.CreateEmployeeRequest and CreateEmployee's one call site,
-- EmployeeStore.Create, which sources it from AccountFromContext — the same
-- rule company creation follows). Without it the employee row is left with a
-- NULL account_id, and the composite FK fk_ec_employee_account rejects the
-- membership AddEmployeeCompany then tries to create for it.
INSERT INTO employee (
    account_id, user_id, default_company_id, first_name, last_name, employee_id, is_active,
    email, mobile_phone, work_phone, job_title, start_date, leave_date, birth_date,
    hourly_labor_rate, is_technician, is_vehicle_operator, is_account_owner, license_class,
    license_number, license_state, license_expiry, street_address, city, region, postal_code,
    country, group_id, custom_fields, table_preferences, dashboard_preferences, updated_at
) VALUES (
    sqlc.arg(account_id), sqlc.arg(user_id), sqlc.arg(default_company_id), sqlc.arg(first_name), sqlc.arg(last_name),
    sqlc.arg(employee_id), sqlc.arg(is_active), sqlc.arg(email),
    sqlc.arg(mobile_phone), sqlc.arg(work_phone), sqlc.arg(job_title), sqlc.arg(start_date),
    sqlc.arg(leave_date), sqlc.arg(birth_date), sqlc.arg(hourly_labor_rate), sqlc.arg(is_technician),
    sqlc.arg(is_vehicle_operator), sqlc.arg(is_account_owner), sqlc.arg(license_class),
    sqlc.arg(license_number), sqlc.arg(license_state), sqlc.arg(license_expiry),
    sqlc.arg(street_address), sqlc.arg(city), sqlc.arg(region), sqlc.arg(postal_code),
    sqlc.arg(country), sqlc.arg(group_id), sqlc.arg(custom_fields), sqlc.arg(table_preferences),
    sqlc.arg(dashboard_preferences), sqlc.arg(updated_at)
)
RETURNING *;

-- CreateAccountOwnerEmployee provisions the account's owner employee with the
-- minimum the employee table requires (see the NOT NULL columns with no
-- default in 000001_init.up.sql) plus account_id and password_hash. It is
-- separate from CreateEmployee because an owner is provisioned before there is
-- any company, role or profile detail to give it; updated_at has no default
-- either, so it is stamped here rather than threaded through as a param.
-- name: CreateAccountOwnerEmployee :one
INSERT INTO employee (
    account_id, first_name, last_name, employee_id, email, mobile_phone,
    work_phone, job_title, license_class, license_number, license_state,
    street_address, city, region, postal_code, country, password_hash, updated_at
) VALUES (
    sqlc.arg(account_id), sqlc.arg(first_name), sqlc.arg(last_name), '', sqlc.arg(email),
    '', '', '', '', '', '', '', '', '', '', '', sqlc.arg(password_hash), now()
)
RETURNING id;

-- name: AddEmployeeCompany :exec
-- account_id is derived from the company row via a scalar subquery rather
-- than taken as a parameter: that way it can never disagree with the
-- company's own account, and it always satisfies the composite FK invariant
-- fk_ec_company_account by construction rather than by trusting the caller.
--
-- A scalar subquery, not INSERT...SELECT...FROM company: the earlier shape
-- silently inserted zero rows for a company_id that names no company, which
-- the caller had no way to distinguish from an already-satisfied ON CONFLICT.
-- Here a nonexistent company_id makes the subquery yield NULL, which the
-- NOT NULL constraint on employee_companies.account_id rejects outright — a
-- real, surfaced error, the same as the foreign key gave before this column
-- existed, instead of a membership the database never actually created.
INSERT INTO employee_companies (employee_id, company_id, account_id)
VALUES (
    sqlc.arg(employee_id),
    sqlc.arg(company_id),
    (SELECT account_id FROM company WHERE id = sqlc.arg(company_id))
)
ON CONFLICT (employee_id, company_id) DO NOTHING;

-- name: ListEmployeeCompanyIDs :many
-- The companies an employee actually belongs to. Login resolves its company_id
-- claim through this rather than trusting employee.default_company_id, which is
-- a plain writable field and can name a company the employee is not a member of.
SELECT company_id FROM employee_companies
WHERE employee_id = sqlc.arg(employee_id)
ORDER BY company_id;

-- name: UpdateEmployee :one
UPDATE employee e SET
    user_id = sqlc.arg(user_id), default_company_id = sqlc.arg(default_company_id),
    first_name = sqlc.arg(first_name), last_name = sqlc.arg(last_name), employee_id = sqlc.arg(employee_id),
    is_active = sqlc.arg(is_active), email = sqlc.arg(email),
    mobile_phone = sqlc.arg(mobile_phone), work_phone = sqlc.arg(work_phone), job_title = sqlc.arg(job_title),
    start_date = sqlc.arg(start_date), leave_date = sqlc.arg(leave_date), birth_date = sqlc.arg(birth_date),
    hourly_labor_rate = sqlc.arg(hourly_labor_rate), is_technician = sqlc.arg(is_technician),
    is_vehicle_operator = sqlc.arg(is_vehicle_operator), is_account_owner = sqlc.arg(is_account_owner),
    license_class = sqlc.arg(license_class), license_number = sqlc.arg(license_number),
    license_state = sqlc.arg(license_state), license_expiry = sqlc.arg(license_expiry),
    street_address = sqlc.arg(street_address), city = sqlc.arg(city), region = sqlc.arg(region),
    postal_code = sqlc.arg(postal_code), country = sqlc.arg(country), group_id = sqlc.arg(group_id),
    custom_fields = sqlc.arg(custom_fields), table_preferences = sqlc.arg(table_preferences),
    dashboard_preferences = sqlc.arg(dashboard_preferences), updated_at = sqlc.arg(updated_at)
WHERE e.id = sqlc.arg(id)
  AND EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.employee_id = e.id AND ec.company_id = sqlc.arg(company_id))
RETURNING e.*;

-- name: DeleteEmployee :exec
DELETE FROM employee e
WHERE e.id = sqlc.arg(id)
  AND EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.employee_id = e.id AND ec.company_id = sqlc.arg(company_id));

-- name: ListAllEmployees :many
-- Cross-company employee list for the admin namespace. Unlike ListEmployees
-- this is deliberately unscoped: it exists to answer "who exists anywhere",
-- which the company-scoped route cannot.
--
-- company_id narrows it to one tenant's staff, through employee_companies -
-- membership, not default_company_id. "Belongs to company X" is what a
-- membership row says; default_company_id only says where a session lands, and
-- an employee can belong to a company that is not their default. A null
-- argument means no filter, so one query serves both questions.
SELECT e.* FROM employee e
WHERE (
    sqlc.narg(company_id)::bigint IS NULL
    OR EXISTS (
        SELECT 1 FROM employee_companies ec
        WHERE ec.employee_id = e.id AND ec.company_id = sqlc.narg(company_id)
    )
)
ORDER BY e.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountAllEmployees :one
SELECT count(*) FROM employee e
WHERE (
    sqlc.narg(company_id)::bigint IS NULL
    OR EXISTS (
        SELECT 1 FROM employee_companies ec
        WHERE ec.employee_id = e.id AND ec.company_id = sqlc.narg(company_id)
    )
);

-- name: GetEmployeeByID :one
-- Unscoped single-employee read for the admin namespace.
SELECT * FROM employee WHERE id = sqlc.arg(id);

-- name: RemoveEmployeeCompaniesNotIn :exec
-- The delete half of a full-replace over employee_companies.
DELETE FROM employee_companies
WHERE employee_id = sqlc.arg(employee_id)
  AND company_id <> ALL(sqlc.arg(company_ids)::bigint[]);

-- name: SetEmployeeDefaultCompany :exec
UPDATE employee SET default_company_id = sqlc.narg(default_company_id), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: SetEmployeeDefaultCompanyIfUnset :exec
UPDATE employee
   SET default_company_id = sqlc.narg(default_company_id)
 WHERE id = sqlc.arg(id) AND default_company_id IS NULL;
