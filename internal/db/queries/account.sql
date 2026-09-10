-- internal/db/queries/account.sql
-- An account is a client organisation. Companies and employees belong to one;
-- platform staff belong to none.

-- name: CreateAccount :one
INSERT INTO account (name) VALUES (sqlc.arg(name)) RETURNING *;

-- name: GetAccount :one
SELECT * FROM account WHERE id = sqlc.arg(id);

-- SetAccountOwner is separate from CreateAccount because the owner employee does
-- not exist yet when the account is inserted: the employee carries account_id,
-- so the account must exist first. Both run in one transaction.
-- name: SetAccountOwner :exec
UPDATE account SET owner_employee_id = sqlc.narg(owner_employee_id) WHERE id = sqlc.arg(id);

-- name: ListAccountsWithCounts :many
SELECT
    a.*,
    (SELECT count(*) FROM company  c WHERE c.account_id = a.id) AS company_count,
    (SELECT count(*) FROM employee e WHERE e.account_id = a.id) AS employee_count
FROM account a
ORDER BY a.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountAccounts :one
SELECT count(*) FROM account;

-- IsEmployeeInAccount backs the ownership-transfer check: a new owner must
-- already belong to the account. The composite constraint makes this one read.
-- name: IsEmployeeInAccount :one
SELECT EXISTS (
    SELECT 1 FROM employee e
    WHERE e.id = sqlc.arg(employee_id) AND e.account_id = sqlc.arg(account_id)
);
