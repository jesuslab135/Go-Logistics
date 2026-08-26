-- Membership changes confer access to a whole tenant — and, because
-- employee.role_id is a single global FK, administrator access there if the
-- employee's role carries is_admin. They are recorded here, append-only: there
-- is deliberately no update or delete.

-- name: RecordMembershipChange :exec
INSERT INTO membership_audit (actor_employee_id, subject_employee_id, company_id, action, occurred_at)
VALUES (sqlc.narg(actor_employee_id), sqlc.arg(subject_employee_id), sqlc.arg(company_id), sqlc.arg(action), sqlc.arg(occurred_at));

-- name: ListMembershipAuditForEmployee :many
SELECT * FROM membership_audit
WHERE subject_employee_id = sqlc.arg(subject_employee_id)
ORDER BY occurred_at DESC, id DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountMembershipAuditForEmployee :one
SELECT count(*) FROM membership_audit WHERE subject_employee_id = sqlc.arg(subject_employee_id);

-- Platform administrators are granted from the CLI only; these back that.

-- name: SetEmployeePlatformAdmin :one
UPDATE employee SET is_platform_admin = sqlc.arg(is_platform_admin), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)
RETURNING id, first_name, last_name, email, is_platform_admin;

-- name: FindEmployeeByEmail :one
SELECT id, first_name, last_name, email, is_platform_admin FROM employee WHERE lower(email) = lower(sqlc.arg(email));

-- name: ListPlatformAdmins :many
SELECT id, first_name, last_name, email FROM employee WHERE is_platform_admin ORDER BY id;
