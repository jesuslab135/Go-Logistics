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
    oe.email AS owner_email,
    (SELECT count(*) FROM company  c WHERE c.account_id = a.id) AS company_count,
    (SELECT count(*) FROM employee e WHERE e.account_id = a.id) AS employee_count
FROM account a
LEFT JOIN employee oe ON oe.id = a.owner_employee_id
ORDER BY a.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- ListAccountCompanies is every company of one account, whether or not the
-- caller holds a membership in it. The account owner needs the whole set to
-- appoint directors; /me/permissions only lists companies they belong to.
-- name: ListAccountCompanies :many
SELECT c.id, c.name, c.logo
FROM company c
WHERE c.account_id = sqlc.arg(account_id)
ORDER BY c.name, c.id;

-- ListOwnersAmong reports which of the given employees own an account, so an
-- employee listing can say who the owner is without a column on employee.
-- name: ListOwnersAmong :many
SELECT a.owner_employee_id::bigint AS employee_id
FROM account a
WHERE a.owner_employee_id = ANY(sqlc.arg(employee_ids)::bigint[]);

-- name: CountAccounts :one
SELECT count(*) FROM account;

-- IsEmployeeInAccount backs the ownership-transfer check: a new owner must
-- already belong to the account. The composite constraint makes this one read.
-- name: IsEmployeeInAccount :one
SELECT EXISTS (
    SELECT 1 FROM employee e
    WHERE e.id = sqlc.arg(employee_id) AND e.account_id = sqlc.arg(account_id)
);

-- GetAccountWithCounts is ListAccountsWithCounts for one account. The select list
-- must stay identical to ListAccountsWithCounts: the handler converts this row
-- to that row type, which only compiles while the two match.
-- name: GetAccountWithCounts :one
SELECT
    a.*,
    oe.email AS owner_email,
    (SELECT count(*) FROM company  c WHERE c.account_id = a.id) AS company_count,
    (SELECT count(*) FROM employee e WHERE e.account_id = a.id) AS employee_count
FROM account a
LEFT JOIN employee oe ON oe.id = a.owner_employee_id
WHERE a.id = sqlc.arg(id);
