# Accounts and Self-Service Onboarding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a platform admin create a client and its owner in one call, and let that owner log in with zero companies and create their own.

**Architecture:** A new `account` entity groups companies and employees. Cross-account membership is made impossible by composite foreign keys rather than by a middleware check. Login learns to issue a company-less token when an employee has an account but no membership yet, and company creation takes its `account_id` from the caller's identity.

**Tech Stack:** Go 1.22+, Gin, pgx/v5, sqlc, golang-migrate, Postgres 16. Tests are `go test ./...` — unit tests only, no database required. Migrations are verified against the local Compose stack.

**Spec:** `docs/superpowers/specs/2026-09-10-accounts-and-self-service-onboarding-design.md`

## Global Constraints

- **Migration numbering:** the next free number is **000019**. Both `.up.sql` and `.down.sql` are required; `sqlc` reads the migrations directory as its schema and ignores `.down.sql` automatically.
- **After any change under `internal/db/queries/` or `internal/db/migrations/`, run `sqlc generate`** and commit the regenerated `internal/db/gen/`. CI runs it and fails on drift.
- **`emit_pointers_for_null_types: true`** — a nullable column generates `*T`, not `sql.NullT`.
- **A company-less session omits `company_id` from the token claims entirely.** It is never encoded as `0`. `Claims.CompanyID` is `*int64`.
- **`company.account_id` is taken from the caller's identity, never from a request body.** A body-supplied `account_id` is ignored, not honoured.
- **Passwords are never echoed in a response and never logged.**
- **An employee with `account_id IS NULL` is platform staff** and can hold no company memberships — the composite FK makes that structural.
- **Never break an existing login.** Every employee who can log in today must still be able to after the migration.
- Local Postgres for migration checks: `docker compose up -d db migrate`, database on **port 5433**, DSN `postgres://postgres:postgres@localhost:5433/fleet?sslmode=disable`.

---

## File Structure

| File | Responsibility |
|---|---|
| `internal/db/migrations/000019_accounts.up.sql` / `.down.sql` | The `account` table, the three `account_id` columns, the backfill, and the composite constraints |
| `internal/db/queries/account.sql` | Account CRUD and the owner lookup |
| `internal/db/queries/employee.sql` | `GetEmployeeIdentity` gains `account_id` and account-ownership; membership check becomes nullable |
| `internal/db/queries/company.sql` | `CreateCompany` gains `account_id` |
| `internal/auth/token.go` | `Claims.CompanyID *int64`, `Claims.AccountID *int64`, new `Issue` signature |
| `internal/http/handler/credential.go` | `resolveLoginCompany` returns an optional company |
| `internal/http/middleware/rbac.go` | `Identity.AccountID`, `RequireAccountOwner` semantics, `AccountFromContext` |
| `internal/http/handler/company.go` | Account-scoped creation, `default_company_id` seeding |
| `internal/http/dto/account.go` | Account request/response types |
| `internal/http/handler/admin_account.go` | The three `/admin/accounts` routes |
| `internal/http/handler/router.go` | Route registration |

---

### Task 1: Migration 000019 — the account table, columns, backfill and constraints

**Files:**
- Create: `internal/db/migrations/000019_accounts.up.sql`
- Create: `internal/db/migrations/000019_accounts.down.sql`

**Interfaces:**
- Consumes: nothing.
- Produces: table `account(id, name, owner_employee_id, is_active, created_at)`; columns `company.account_id NOT NULL`, `employee.account_id NULL`, `employee_companies.account_id NOT NULL`; composite FKs preventing cross-account membership.

There is no database-backed test harness in this repository — every existing test is a pure unit test and `go test ./...` runs without Postgres. So this task is verified by applying the migration to a real database and asserting the constraint actually refuses a cross-account insert. That assertion is the point of the task; do not skip it.

- [ ] **Step 1: Write the up migration**

```sql
-- 000019_accounts.up.sql
-- An account is a client organisation: the thing that owns companies. Until now
-- the model was flat — employee <-> employee_companies <-> company — so nothing
-- recorded which companies belonged to which client, and a client could not be
-- scoped, billed or isolated as a unit.
--
-- owner_employee_id is nullable because an account and its owner reference each
-- other: creation inserts the account, then the owner employee carrying
-- account_id, then sets this. It stays nullable so ownership can be transferred
-- without a window where the column has no legal value.
CREATE TABLE account (
    id                bigserial    PRIMARY KEY,
    name              varchar(200) NOT NULL,
    owner_employee_id bigint,
    is_active         boolean      NOT NULL DEFAULT true,
    created_at        timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE company            ADD COLUMN account_id bigint;
ALTER TABLE employee           ADD COLUMN account_id bigint;
ALTER TABLE employee_companies ADD COLUMN account_id bigint;

-- Everything that exists today belongs to one client: there has never been a
-- second. Create that account and adopt the existing rows into it. Guarded on
-- there being data at all, so a fresh development database is not given a
-- phantom account.
DO $$
DECLARE acct_id bigint;
BEGIN
    IF EXISTS (SELECT 1 FROM company) OR EXISTS (SELECT 1 FROM employee) THEN
        INSERT INTO account (name, created_at)
        VALUES ('Go Logistics', now())
        RETURNING id INTO acct_id;

        UPDATE company            SET account_id = acct_id WHERE account_id IS NULL;
        UPDATE employee           SET account_id = acct_id WHERE account_id IS NULL;
        UPDATE employee_companies SET account_id = acct_id WHERE account_id IS NULL;

        UPDATE account
           SET owner_employee_id = (
               SELECT id FROM employee WHERE is_account_owner ORDER BY id LIMIT 1
           )
         WHERE id = acct_id;
    END IF;
END $$;

ALTER TABLE company            ALTER COLUMN account_id SET NOT NULL;
ALTER TABLE employee_companies ALTER COLUMN account_id SET NOT NULL;
-- employee.account_id stays nullable: NULL means platform staff, who belong to
-- no client and reach tenant data through /api/v1/admin/*.

ALTER TABLE company  ADD CONSTRAINT fk_company_account
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE RESTRICT;
ALTER TABLE employee ADD CONSTRAINT fk_employee_account
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE RESTRICT;
ALTER TABLE account  ADD CONSTRAINT fk_account_owner
    FOREIGN KEY (owner_employee_id) REFERENCES employee(id) ON DELETE SET NULL;

-- The composite foreign keys below need these as their referenced unique keys.
ALTER TABLE employee ADD CONSTRAINT uq_employee_account UNIQUE (id, account_id);
ALTER TABLE company  ADD CONSTRAINT uq_company_account  UNIQUE (id, account_id);

-- The isolation invariant, enforced by the database rather than by middleware:
-- a membership may only link an employee and a company in the SAME account.
-- A cross-account membership becomes impossible to insert — from the API, from
-- the CLI, from a hand-written INSERT during an incident. A middleware check
-- cannot give that guarantee, because every future endpoint that writes a
-- membership would have to remember it.
--
-- The single-column FKs are dropped: the composite ones subsume them, and
-- keeping both would mean two cascade paths for the same relationship.
ALTER TABLE employee_companies DROP CONSTRAINT fk_employee_companies_employee;
ALTER TABLE employee_companies DROP CONSTRAINT fk_employee_companies_company;

ALTER TABLE employee_companies ADD CONSTRAINT fk_ec_employee_account
    FOREIGN KEY (employee_id, account_id) REFERENCES employee(id, account_id) ON DELETE CASCADE;
ALTER TABLE employee_companies ADD CONSTRAINT fk_ec_company_account
    FOREIGN KEY (company_id, account_id) REFERENCES company(id, account_id) ON DELETE CASCADE;

CREATE INDEX idx_company_account  ON company (account_id);
CREATE INDEX idx_employee_account ON employee (account_id);
```

- [ ] **Step 2: Write the down migration**

```sql
-- 000019_accounts.down.sql
ALTER TABLE employee_companies DROP CONSTRAINT fk_ec_company_account;
ALTER TABLE employee_companies DROP CONSTRAINT fk_ec_employee_account;

ALTER TABLE employee_companies ADD CONSTRAINT fk_employee_companies_employee
    FOREIGN KEY (employee_id) REFERENCES employee(id) ON DELETE CASCADE;
ALTER TABLE employee_companies ADD CONSTRAINT fk_employee_companies_company
    FOREIGN KEY (company_id) REFERENCES company(id) ON DELETE CASCADE;

ALTER TABLE company  DROP CONSTRAINT uq_company_account;
ALTER TABLE employee DROP CONSTRAINT uq_employee_account;

ALTER TABLE account  DROP CONSTRAINT fk_account_owner;
ALTER TABLE employee DROP CONSTRAINT fk_employee_account;
ALTER TABLE company  DROP CONSTRAINT fk_company_account;

DROP INDEX IF EXISTS idx_employee_account;
DROP INDEX IF EXISTS idx_company_account;

ALTER TABLE employee_companies DROP COLUMN account_id;
ALTER TABLE employee           DROP COLUMN account_id;
ALTER TABLE company            DROP COLUMN account_id;

DROP TABLE account;
```

- [ ] **Step 3: Apply it to a real database**

```bash
docker compose up -d db
docker compose run --rm migrate
```

Expected: the migrate container exits 0. If it reports a dirty version, the migration failed partway — read the error, fix the SQL, then `migrate force <previous version>` before retrying.

- [ ] **Step 4: Prove the constraint actually refuses a cross-account membership**

This is the assertion the whole task exists for. Run:

```bash
docker compose exec -T db psql -U postgres -d fleet <<'SQL'
BEGIN;
INSERT INTO account (name) VALUES ('Client A') RETURNING id \gset a_
INSERT INTO account (name) VALUES ('Client B') RETURNING id \gset b_
INSERT INTO company (name, tax_id, address, created_at, phone, email, website,
                     city, region, postal_code, account_id)
VALUES ('A Co','TAXA','addr',now(),'','','','','','', :a_id) RETURNING id \gset comp_
INSERT INTO employee (first_name,last_name,employee_id,email,mobile_phone,work_phone,
                      job_title,license_class,license_number,license_state,
                      street_address,city,region,postal_code,country,updated_at,account_id)
VALUES ('B','Person','B1','b@example.invalid','','','','','','','','','','','',now(), :b_id) RETURNING id \gset emp_
-- Client B's employee, Client A's company: must be rejected.
INSERT INTO employee_companies (employee_id, company_id, account_id)
VALUES (:emp_id, :comp_id, :b_id);
ROLLBACK;
SQL
```

Expected: `ERROR: insert or update on table "employee_companies" violates foreign key constraint "fk_ec_company_account"`.

If that insert **succeeds**, the constraint is wrong and the task is not done. Do not proceed.

- [ ] **Step 5: Verify the down migration reverses cleanly**

```bash
docker compose run --rm migrate -path /migrations -database "postgres://postgres:postgres@db:5432/fleet?sslmode=disable" down 1
docker compose run --rm migrate
```

Expected: both exit 0. This proves the migration is reversible under pressure, which matters because it runs against production.

- [ ] **Step 6: Commit**

```bash
git add internal/db/migrations/000019_accounts.up.sql internal/db/migrations/000019_accounts.down.sql
git commit -m "feat(db): add accounts and enforce cross-account isolation structurally"
```

---

### Task 2: Queries and regeneration

**Files:**
- Create: `internal/db/queries/account.sql`
- Modify: `internal/db/queries/employee.sql` (the `GetEmployeeIdentity` block, lines 11-30)
- Modify: `internal/db/queries/company.sql` (`CreateCompany`)
- Modify: `internal/db/gen/**` (regenerated, not hand-edited)

**Interfaces:**
- Consumes: the schema from Task 1.
- Produces (sqlc-generated Go):
  - `CreateAccount(ctx, CreateAccountParams{Name string}) (Account, error)`
  - `SetAccountOwner(ctx, SetAccountOwnerParams{ID int64, OwnerEmployeeID *int64}) error`
  - `GetAccount(ctx, id int64) (Account, error)`
  - `ListAccountsWithCounts(ctx, ListAccountsWithCountsParams{Limit, Offset int32}) ([]ListAccountsWithCountsRow, error)` — row carries `ID, Name, OwnerEmployeeID, IsActive, CreatedAt, CompanyCount, EmployeeCount`
  - `CountAccounts(ctx) (int64, error)`
  - `GetEmployeeIdentity(ctx, GetEmployeeIdentityParams{ID int64, CompanyID *int64}) (GetEmployeeIdentityRow, error)` — row gains `AccountID *int64` and `IsAccountOwner bool`
  - `CreateCompany` params gain `AccountID int64`

- [ ] **Step 1: Write the account queries**

```sql
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
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: CountAccounts :one
SELECT count(*) FROM account;

-- IsEmployeeInAccount backs the ownership-transfer check: a new owner must
-- already belong to the account. The composite constraint makes this one read.
-- name: IsEmployeeInAccount :one
SELECT EXISTS (
    SELECT 1 FROM employee e
    WHERE e.id = sqlc.arg(employee_id) AND e.account_id = sqlc.arg(account_id)
);
```

- [ ] **Step 2: Rewrite GetEmployeeIdentity**

Replace the `GetEmployeeIdentity` block in `internal/db/queries/employee.sql` (lines 11-30) with:

```sql
-- GetEmployeeIdentity backs the authorization middleware. Django resolved
-- request.user.employee (and its role) from the database on every request, so
-- revoking a role or deactivating an employee takes effect immediately rather
-- than at the next token refresh; this query preserves that.
--
-- company_id is nullable: a company-less session (an account owner who has not
-- created a company yet) has no tenant to scope to. sqlc.narg makes the
-- membership EXISTS resolve to false for such a caller rather than requiring a
-- sentinel id that does not exist.
--
-- is_account_owner no longer reads employee.is_account_owner, which was a global
-- boolean meaning only "may create companies". It now means what it says: this
-- employee owns the account they belong to.
-- name: GetEmployeeIdentity :one
SELECT
    e.id,
    e.is_active,
    e.account_id,
    e.role_id,
    e.is_platform_admin,
    COALESCE(r.is_admin, false)          AS is_admin,
    COALESCE(r.permissions, '{}'::jsonb) AS permissions,
    EXISTS (
        SELECT 1 FROM employee_companies ec
        WHERE ec.employee_id = e.id AND ec.company_id = sqlc.narg(company_id)
    ) AS is_member,
    EXISTS (
        SELECT 1 FROM account a
        WHERE a.id = e.account_id AND a.owner_employee_id = e.id
    ) AS is_account_owner
FROM employee e
LEFT JOIN role r ON r.id = e.role_id
WHERE e.id = sqlc.arg(id);
```

- [ ] **Step 3: Add account_id to CreateCompany**

In `internal/db/queries/company.sql`, add `account_id` to the `CreateCompany` insert's column list and `sqlc.arg(account_id)` to its `VALUES` list, keeping the existing column order otherwise unchanged.

- [ ] **Step 4: Regenerate and confirm the code compiles**

```bash
sqlc generate
go build ./...
```

Expected: `sqlc generate` succeeds. `go build` **fails** — every caller of `GetEmployeeIdentity` now passes an `int64` where a `*int64` is wanted, and `CreateCompany` is missing a field. That failure is the correct state at this point; Tasks 3-6 fix each call site. Note the exact list of failing files, it is your work list.

- [ ] **Step 5: Commit**

```bash
git add internal/db/queries internal/db/gen
git commit -m "feat(db): account queries, nullable company scope on identity"
```

---

### Task 3: Optional company scope in the token

**Files:**
- Modify: `internal/auth/token.go`
- Test: `internal/auth/token_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `Claims{CompanyID *int64, AccountID *int64, IsAdmin bool, Type string, jwt.RegisteredClaims}`
  - `Issue(employeeID int64, companyID *int64, accountID *int64, isAdmin bool) (TokenPair, error)`

- [ ] **Step 1: Write the failing test**

```go
// internal/auth/token_test.go
package auth

import (
	"testing"
	"time"
)

func i64(v int64) *int64 { return &v }

// NewTokenService takes the secret as a string (it converts internally) —
// verified against internal/auth/token.go:50.
func newTestService(t *testing.T) *TokenService {
	t.Helper()
	return NewTokenService("test-secret-value", "fleet-test", time.Minute, time.Hour)
}

func TestIssueAndParseCarriesCompanyAndAccount(t *testing.T) {
	s := newTestService(t)
	pair, err := s.Issue(42, i64(7), i64(3), true)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	claims, err := s.Parse(pair.AccessToken, TypeAccess)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.EmployeeID() != 42 {
		t.Fatalf("employee id = %d, want 42", claims.EmployeeID())
	}
	if claims.CompanyID == nil || *claims.CompanyID != 7 {
		t.Fatalf("company id = %v, want 7", claims.CompanyID)
	}
	if claims.AccountID == nil || *claims.AccountID != 3 {
		t.Fatalf("account id = %v, want 3", claims.AccountID)
	}
}

// A company-less session is the whole point of this change: an account owner who
// has not created a company yet must still receive a usable token. The claim is
// ABSENT rather than zero — a zero would flow into a scoped query as a real
// argument and silently scope to a company that cannot exist.
func TestIssueOmitsCompanyWhenThereIsNone(t *testing.T) {
	s := newTestService(t)
	pair, err := s.Issue(42, nil, i64(3), false)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	claims, err := s.Parse(pair.AccessToken, TypeAccess)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.CompanyID != nil {
		t.Fatalf("company id = %v, want nil", *claims.CompanyID)
	}
	if claims.AccountID == nil || *claims.AccountID != 3 {
		t.Fatalf("account id = %v, want 3", claims.AccountID)
	}
}
```

If `NewTokenService`'s signature differs from the call above, match the real one — read `internal/auth/token.go` and adapt the helper, not the assertions.

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./internal/auth/ -run TestIssue -v`
Expected: compile failure — `Issue` takes `int64`, not `*int64`.

- [ ] **Step 3: Change the claims and the signature**

In `internal/auth/token.go`:

```go
// Claims are the JWT payload. Subject holds the employee id (string per the JWT
// spec); EmployeeID decodes it.
//
// CompanyID is a pointer and omitted when absent. A company-less session — an
// account owner who has not created a company yet — has no tenant to scope to,
// and encoding 0 would be worse than encoding nothing: it would flow into
// scoped queries as a real argument and quietly resolve to "not a member",
// which is the right answer for the wrong reason.
type Claims struct {
	CompanyID *int64 `json:"company_id,omitempty"`
	AccountID *int64 `json:"account_id,omitempty"`
	IsAdmin   bool   `json:"is_admin"`
	Type      string `json:"typ"`
	jwt.RegisteredClaims
}
```

Change `Issue` and the private `sign` to take `companyID *int64, accountID *int64`, threading both into the claims unchanged.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/auth/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/auth
git commit -m "feat(auth): make the company claim optional, add the account claim"
```

---

### Task 4: Company-less login

**Files:**
- Modify: `internal/http/handler/credential.go`
- Modify: `internal/http/handler/auth.go` (the `Identity` struct and the three `Issue` call sites)
- Test: `internal/http/handler/credential_test.go`

**Interfaces:**
- Consumes: `auth.Issue` from Task 3.
- Produces:
  - `resolveLoginCompany(defaultCompanyID *int64, memberships []int64) *int64` — returns nil when there is no membership, instead of the old `(int64, bool)`
  - `handler.Identity{EmployeeID int64, CompanyID *int64, AccountID *int64, IsAdmin bool}`

- [ ] **Step 1: Write the failing test**

Replace the existing `TestResolveLoginCompany` in `internal/http/handler/credential_test.go` with:

```go
package handler

import "testing"

func ptr(v int64) *int64 { return &v }

// A login token is only safe if its company_id claim names a company the
// employee is actually a member of: every downstream query trusts that claim.
// With no membership the answer is now "no company" rather than "no token" —
// an account owner who has not created a company yet must be able to log in.
func TestResolveLoginCompany(t *testing.T) {
	tests := []struct {
		name        string
		defaultID   *int64
		memberships []int64
		want        *int64
	}{
		{"default company that is a real membership wins", ptr(7), []int64{3, 7, 9}, ptr(7)},
		{"nil default falls back to the lowest membership", nil, []int64{9, 3, 7}, ptr(3)},
		{"default naming a company the employee does not belong to is ignored", ptr(42), []int64{9, 3}, ptr(3)},
		{"no membership yields no company rather than a failure", ptr(7), nil, nil},
		{"no membership and no default yields no company", nil, nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveLoginCompany(tc.defaultID, tc.memberships)
			switch {
			case tc.want == nil && got != nil:
				t.Fatalf("got %d, want no company", *got)
			case tc.want != nil && got == nil:
				t.Fatalf("got no company, want %d", *tc.want)
			case tc.want != nil && *got != *tc.want:
				t.Fatalf("got %d, want %d", *got, *tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./internal/http/handler/ -run TestResolveLoginCompany -v`
Expected: compile failure — `resolveLoginCompany` returns two values.

- [ ] **Step 3: Implement**

In `internal/http/handler/credential.go`:

```go
// resolveLoginCompany picks the company a session is scoped to. The employee's
// default_company_id wins when it names a real membership; otherwise the lowest
// membership id does, so the choice is deterministic across logins.
//
// An employee with no membership resolves to nil, which is a company-less
// session rather than a refusal: an account owner provisioned by a platform
// admin has no company until they create one, and refusing them a token is the
// deadlock this change exists to break. Whether they may log in at all is
// decided by Verify, on whether they belong to an account.
func resolveLoginCompany(defaultCompanyID *int64, memberships []int64) *int64 {
	if len(memberships) == 0 {
		return nil
	}
	if defaultCompanyID != nil && slices.Contains(memberships, *defaultCompanyID) {
		return defaultCompanyID
	}
	lowest := slices.Min(memberships)
	return &lowest
}
```

Then in `Verify`, replace the `if !ok { return Identity{}, ErrNoCompanyMembership }` block with:

```go
	companyID := resolveLoginCompany(row.DefaultCompanyID, memberships)

	// An employee belonging to no account and holding no membership has nowhere
	// to be: refuse, as before. An employee with an account but no company yet
	// gets a company-less session, which can reach only /me/permissions,
	// POST /companies and /auth/switch-company.
	if companyID == nil && row.AccountID == nil {
		return Identity{}, ErrNoCompanyMembership
	}

	return Identity{
		EmployeeID: row.ID,
		CompanyID:  companyID,
		AccountID:  row.AccountID,
		IsAdmin:    row.IsAdmin,
	}, nil
```

`GetEmployeeAuthByEmail` must return `account_id`. Add `e.account_id` to that query's select list in `internal/db/queries/employee.sql` and re-run `sqlc generate`.

Update `handler.Identity` in `internal/http/handler/auth.go`:

```go
type Identity struct {
	EmployeeID int64
	CompanyID  *int64
	AccountID  *int64
	IsAdmin    bool
}
```

Update the three `h.tokens.Issue(...)` call sites — `Login`, `Refresh`, `SwitchCompany` — to pass the pointers. In `SwitchCompany` the target company is always known, so pass `&req.CompanyID`.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/http/handler/ -run TestResolveLoginCompany -v`
Expected: PASS, 5 subtests.

- [ ] **Step 5: Commit**

```bash
git add internal/http/handler/credential.go internal/http/handler/credential_test.go internal/http/handler/auth.go internal/db/queries/employee.sql internal/db/gen
git commit -m "feat(auth): issue a company-less session to an owner with no company yet"
```

---

### Task 5: Identity carries the account, and account ownership is real

**Files:**
- Modify: `internal/http/middleware/rbac.go`
- Modify: `internal/http/handler/identity.go`
- Test: `internal/http/middleware/rbac_test.go`

**Interfaces:**
- Consumes: `GetEmployeeIdentity` from Task 2.
- Produces:
  - `middleware.Identity` gains `AccountID *int64`; `IsAccountOwner` now means "owns the account they belong to"
  - `middleware.AccountFromContext(ctx context.Context) *int64`
  - `IdentityLoader.LoadIdentity(ctx, employeeID int64, companyID *int64) (Identity, error)`

- [ ] **Step 1: Write the failing test**

```go
// internal/http/middleware/rbac_test.go — add to the existing file
package middleware

import "testing"

// RequireAccountOwner used to read a global boolean whose only meaning was "may
// create companies". It now means the caller owns the account they belong to,
// so the predicate must be false for a member of an account they do not own.
func TestAccountOwnerPredicate(t *testing.T) {
	owner := Identity{IsActive: true, IsAccountOwner: true}
	if !owner.IsAccountOwner {
		t.Fatal("an account owner should satisfy the predicate")
	}
	member := Identity{IsActive: true, IsAccountOwner: false}
	if member.IsAccountOwner {
		t.Fatal("a non-owner must not satisfy the predicate")
	}
}

// A company-less session has no tenant. It must not be reported as a member of
// anything, because every scoped query trusts that flag.
func TestCompanylessIdentityIsNotAMember(t *testing.T) {
	id := Identity{IsActive: true, AccountID: i64ptr(3), IsMember: false}
	if id.IsMember {
		t.Fatal("a company-less identity must not be a member")
	}
	if id.AccountID == nil {
		t.Fatal("a company-less identity still belongs to an account")
	}
}

func i64ptr(v int64) *int64 { return &v }
```

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./internal/http/middleware/ -run 'TestAccountOwner|TestCompanyless' -v`
Expected: compile failure — `Identity` has no `AccountID` field.

- [ ] **Step 3: Implement**

In `internal/http/middleware/rbac.go`, add to `Identity`:

```go
	// AccountID is the client organisation this person belongs to. NULL means
	// platform staff, who belong to no client and hold no company memberships.
	AccountID *int64
```

Change the `IsAccountOwner` doc comment to say it means ownership of `AccountID`, not the retired global boolean. `RequireAccountOwner`'s body does not change — the predicate is the same, its meaning now comes from the query.

Change `IdentityLoader`'s interface method and `RequireIdentity`'s call to pass `claims.CompanyID` (already `*int64`) straight through.

Add the accessor next to the existing `EmployeeFromContext`:

```go
// AccountFromContext returns the client organisation of the caller, or nil for
// platform staff. Company creation reads it so a new company can never be
// planted inside another client's account.
func AccountFromContext(ctx context.Context) *int64 {
	id, ok := ctx.Value(identityContextKey{}).(Identity)
	if !ok {
		return nil
	}
	return id.AccountID
}
```

In `internal/http/handler/identity.go`, thread `AccountID` and the new `IsAccountOwner` through from the row, and change the signature to accept `companyID *int64`.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/http/middleware/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/http/middleware internal/http/handler/identity.go
git commit -m "feat(authz): identity carries the account; ownership means owning it"
```

---

### Task 6: Company creation is account-scoped

**Files:**
- Modify: `internal/http/handler/company.go` (`CompanyStore.Create`, lines 69-105)
- Test: `internal/http/handler/company_account_test.go`

**Interfaces:**
- Consumes: `middleware.AccountFromContext` (Task 5), `CreateCompany` with `AccountID` (Task 2).
- Produces: no new exported symbols.

- [ ] **Step 1: Write the failing test**

```go
// internal/http/handler/company_account_test.go
package handler

import (
	"context"
	"testing"

	"fleet/internal/http/middleware"
)

// The account a company lands in comes from the caller's identity, never from
// the request body. Honouring a body-supplied account would let an owner plant
// a company inside another client.
func TestAccountForNewCompanyComesFromIdentity(t *testing.T) {
	ctx := middleware.ContextWithIdentity(context.Background(), middleware.Identity{
		EmployeeID: 5, AccountID: ptr(3), IsAccountOwner: true, IsActive: true,
	})
	got := middleware.AccountFromContext(ctx)
	if got == nil || *got != 3 {
		t.Fatalf("account = %v, want 3", got)
	}
}

// Platform staff belong to no account, so they have no account to create a
// company into. The handler must refuse rather than insert a NULL.
func TestPlatformStaffHaveNoAccountForCompanyCreation(t *testing.T) {
	ctx := middleware.ContextWithIdentity(context.Background(), middleware.Identity{
		EmployeeID: 1, AccountID: nil, IsPlatformAdmin: true, IsActive: true,
	})
	if middleware.AccountFromContext(ctx) != nil {
		t.Fatal("platform staff must resolve to no account")
	}
}
```

This needs an exported way to build a context in tests. Add to `internal/http/middleware/rbac.go`:

```go
// ContextWithIdentity injects an identity. RequireIdentity uses it in the
// request path; tests use it to construct a caller without a live database.
func ContextWithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityContextKey{}, id)
}
```

and have `RequireIdentity` call it instead of inlining `context.WithValue`, so there is one definition rather than two.

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./internal/http/handler/ -run 'TestAccount|TestPlatformStaff' -v`
Expected: compile failure — `middleware.ContextWithIdentity` is undefined.

- [ ] **Step 3: Implement**

Add `ContextWithIdentity` as above. Then in `CompanyStore.Create`, before opening the transaction:

```go
	// The account comes from the caller, never from the body: honouring a
	// body-supplied account_id would let an owner plant a company inside
	// another client. Platform staff belong to no account and so have no
	// account to create into.
	accountID := middleware.AccountFromContext(ctx)
	if accountID == nil {
		return dto.CompanyResponse{}, apierr.Forbidden(
			"only a member of a client account can create a company")
	}
```

Pass `AccountID: *accountID` in `CreateCompanyParams`. Then, after `bootstrap.Company` and before the commit:

```go
	// A company-less owner has no default tenant, so their next login would
	// issue another company-less token and onboarding would look broken. Seed it
	// with the company they just created. Left alone when already set: silently
	// repointing someone's default tenant is a surprise, not a convenience.
	employeeID := middleware.EmployeeFromContext(ctx)
	if err := qtx.SetEmployeeDefaultCompanyIfUnset(ctx, gen.SetEmployeeDefaultCompanyIfUnsetParams{
		ID:               employeeID,
		DefaultCompanyID: &row.ID,
	}); err != nil {
		return dto.CompanyResponse{}, err
	}
```

Add that query to `internal/db/queries/employee.sql` and re-run `sqlc generate`:

```sql
-- name: SetEmployeeDefaultCompanyIfUnset :exec
UPDATE employee
   SET default_company_id = sqlc.narg(default_company_id)
 WHERE id = sqlc.arg(id) AND default_company_id IS NULL;
```

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/http/handler/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/http/handler/company.go internal/http/handler/company_account_test.go internal/http/middleware/rbac.go internal/db/queries/employee.sql internal/db/gen
git commit -m "feat(companies): scope creation to the caller's account, seed their default"
```

---

### Task 7: Platform-admin account provisioning

**Files:**
- Create: `internal/http/dto/account.go`
- Create: `internal/http/handler/admin_account.go`
- Modify: `internal/http/handler/router.go` (`registerAdminRoutes`)
- Test: `internal/http/handler/admin_account_test.go`

**Interfaces:**
- Consumes: the account queries (Task 2), `auth.HashPassword`.
- Produces:
  - `NewAdminAccountHandler(q *gen.Queries, pool *pgxpool.Pool) *AdminAccountHandler`
  - methods `Create`, `List`, `SetOwner`
  - `dto.CreateAccountRequest{Name, OwnerFirstName, OwnerLastName, OwnerEmail, OwnerPassword string}`
  - `dto.AccountResponse{ID, Name, OwnerEmployeeID *int64, IsActive, CreatedAt, CompanyCount, EmployeeCount}`

- [ ] **Step 1: Write the failing test**

```go
// internal/http/handler/admin_account_test.go
package handler

import (
	"strings"
	"testing"

	"fleet/internal/http/dto"
)

// Provisioning is one call because the two-call alternative can leave an account
// whose owner exists but cannot log in. Every field the owner needs must
// therefore be present and validated up front.
func TestValidateCreateAccountRequest(t *testing.T) {
	tests := []struct {
		name string
		in   dto.CreateAccountRequest
		want string
	}{
		{"complete request passes", dto.CreateAccountRequest{
			Name: "Acme", OwnerFirstName: "Ada", OwnerLastName: "Byron",
			OwnerEmail: "ada@acme.test", OwnerPassword: "correct-horse-battery",
		}, ""},
		{"missing account name", dto.CreateAccountRequest{
			OwnerFirstName: "Ada", OwnerLastName: "Byron",
			OwnerEmail: "ada@acme.test", OwnerPassword: "correct-horse-battery",
		}, "name"},
		{"missing owner email", dto.CreateAccountRequest{
			Name: "Acme", OwnerFirstName: "Ada", OwnerLastName: "Byron",
			OwnerPassword: "correct-horse-battery",
		}, "owner_email"},
		{"short password is refused", dto.CreateAccountRequest{
			Name: "Acme", OwnerFirstName: "Ada", OwnerLastName: "Byron",
			OwnerEmail: "ada@acme.test", OwnerPassword: "short",
		}, "owner_password"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCreateAccount(tc.in)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("expected valid, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want it to name %q", err, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./internal/http/handler/ -run TestValidateCreateAccount -v`
Expected: compile failure — `validateCreateAccount` and `dto.CreateAccountRequest` are undefined.

- [ ] **Step 3: Write the DTOs**

```go
// internal/http/dto/account.go
package dto

import "time"

// CreateAccountRequest provisions a client and its owner in one act. The
// password is accepted here rather than in a follow-up call because a
// two-step flow can leave an account whose owner exists but cannot log in.
type CreateAccountRequest struct {
	Name           string `json:"name"`
	OwnerFirstName string `json:"owner_first_name"`
	OwnerLastName  string `json:"owner_last_name"`
	OwnerEmail     string `json:"owner_email"`
	// OwnerPassword is never echoed back and never logged.
	OwnerPassword string `json:"owner_password"`
}

type AccountResponse struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	OwnerEmployeeID *int64    `json:"owner_employee_id"`
	OwnerEmail      string    `json:"owner_email,omitempty"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	CompanyCount    int64     `json:"company_count"`
	EmployeeCount   int64     `json:"employee_count"`
}

type AccountPage struct {
	Data    []AccountResponse `json:"data"`
	Total   int64             `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasNext bool              `json:"has_next"`
}

type SetAccountOwnerRequest struct {
	EmployeeID int64 `json:"employee_id"`
}
```

- [ ] **Step 4: Write the handler**

```go
// internal/http/handler/admin_account.go
package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/auth"
	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/platform/apierr"
)

// AdminAccountHandler serves the cross-client /admin/accounts routes. Creating a
// client is a platform act: nothing inside a tenant can reach these.
type AdminAccountHandler struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewAdminAccountHandler(q *gen.Queries, pool *pgxpool.Pool) *AdminAccountHandler {
	return &AdminAccountHandler{q: q, pool: pool}
}

// minOwnerPasswordLength matches what an operator can reasonably be asked to
// generate. It is not a policy engine; it is a floor that stops an unusable
// credential being provisioned by accident.
const minOwnerPasswordLength = 12

func validateCreateAccount(in dto.CreateAccountRequest) error {
	missing := map[string]string{}
	if strings.TrimSpace(in.Name) == "" {
		missing["name"] = "this field is required"
	}
	if strings.TrimSpace(in.OwnerFirstName) == "" {
		missing["owner_first_name"] = "this field is required"
	}
	if strings.TrimSpace(in.OwnerLastName) == "" {
		missing["owner_last_name"] = "this field is required"
	}
	if strings.TrimSpace(in.OwnerEmail) == "" {
		missing["owner_email"] = "this field is required"
	}
	if len(in.OwnerPassword) < minOwnerPasswordLength {
		missing["owner_password"] = "must be at least 12 characters"
	}
	if len(missing) == 0 {
		return nil
	}
	return apierr.Validation(missing)
}

// Create provisions the account, its owner employee and the owner's password in
// one transaction. The three exist together or not at all: an account with an
// owner who cannot log in is the state this design exists to prevent.
func (h *AdminAccountHandler) Create(c *gin.Context) {
	var in dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		apierr.Abort(c, apierr.BadRequest("invalid request body").Wrap(err))
		return
	}
	if err := validateCreateAccount(in); err != nil {
		apierr.Abort(c, err)
		return
	}

	hash, err := auth.HashPassword(in.OwnerPassword)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	ctx := c.Request.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	defer tx.Rollback(ctx)
	qtx := h.q.WithTx(tx)

	account, err := qtx.CreateAccount(ctx, in.Name)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	// The owner carries account_id, so the account must exist first; the account
	// then learns its owner. That ordering is why owner_employee_id is nullable.
	owner, err := qtx.CreateAccountOwnerEmployee(ctx, gen.CreateAccountOwnerEmployeeParams{
		AccountID:    &account.ID,
		FirstName:    in.OwnerFirstName,
		LastName:     in.OwnerLastName,
		Email:        in.OwnerEmail,
		PasswordHash: hash,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if err := qtx.SetAccountOwner(ctx, gen.SetAccountOwnerParams{
		ID:              account.ID,
		OwnerEmployeeID: &owner.ID,
	}); err != nil {
		apierr.Abort(c, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		apierr.Abort(c, err)
		return
	}

	// The password is deliberately absent from this response.
	c.JSON(http.StatusCreated, dto.AccountResponse{
		ID:              account.ID,
		Name:            account.Name,
		OwnerEmployeeID: &owner.ID,
		OwnerEmail:      in.OwnerEmail,
		IsActive:        account.IsActive,
		CreatedAt:       account.CreatedAt,
		EmployeeCount:   1,
	})
}
```

Write `List` and `SetOwner` in the same shape as the existing `AdminCompanyHandler` methods — `List` over `ListAccountsWithCounts` + `CountAccounts` using `paginate`, `SetOwner` validating with `IsEmployeeInAccount` and refusing `422` when the target does not belong to the account.

Add `CreateAccountOwnerEmployee` to `internal/db/queries/employee.sql`, inserting the minimum an employee row needs plus `account_id` and `password_hash`, returning `id`. Re-run `sqlc generate`.

- [ ] **Step 5: Register the routes**

In `registerAdminRoutes` in `internal/http/handler/router.go`, alongside the existing admin wiring:

```go
	accounts := NewAdminAccountHandler(d.Queries, d.Pool)
	admin.POST("/admin/accounts", accounts.Create)
	admin.GET("/admin/accounts", accounts.List)
	admin.POST("/admin/accounts/:id/set-owner", accounts.SetOwner)
```

- [ ] **Step 6: Run the tests**

Run: `go test ./... 2>&1 | tail -20`
Expected: PASS across every package.

- [ ] **Step 7: Commit**

```bash
git add internal/http/dto/account.go internal/http/handler/admin_account.go internal/http/handler/admin_account_test.go internal/http/handler/router.go internal/db/queries/employee.sql internal/db/gen
git commit -m "feat(admin): provision a client and its owner in one call"
```

---

### Task 8: End-to-end verification against a live stack

**Files:**
- Modify: `README.md` (the onboarding section)
- Modify: `docs/swagger.json`, `docs/docs.go`, `docs/swagger.yaml` (regenerated)

**Interfaces:**
- Consumes: everything above.
- Produces: no new symbols.

This task proves the deadlock is actually broken. Unit tests cannot show that; only a running stack can.

- [ ] **Step 1: Bring up a clean stack**

```bash
docker compose down -v
docker compose up -d db migrate
docker compose up -d api
```

Expected: all healthy, migrate exits 0.

- [ ] **Step 2: Grant yourself platform admin**

```bash
go run ./cmd/cli platform-admin --email <your-email>
```

On a fresh database with no employees, first create one via `fleet-cli` or a direct insert, then grant. Record which you did.

- [ ] **Step 3: Provision a client**

```bash
curl -s -X POST http://localhost:8080/api/v1/admin/accounts \
  -H "Authorization: Bearer $PLATFORM_TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Acme Fleet","owner_first_name":"Ada","owner_last_name":"Byron",
       "owner_email":"ada@acme.test","owner_password":"correct-horse-battery"}'
```

Expected: `201` with an account id and owner employee id. **Confirm the response does not contain the password.**

- [ ] **Step 4: Log in as the owner, who has no company**

```bash
curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@acme.test","password":"correct-horse-battery"}'
```

Expected: `200` with a token pair. This is the deadlock broken — before this change it was `403 no_company_membership`.

Decode the access token's payload and confirm `company_id` is **absent**, not `0`.

- [ ] **Step 5: Confirm the session can reach only what it should**

```bash
curl -s -o /dev/null -w 'me:      %{http_code}\n' -H "Authorization: Bearer $OWNER" http://localhost:8080/api/v1/me/permissions
curl -s -o /dev/null -w 'assets:  %{http_code}\n' -H "Authorization: Bearer $OWNER" http://localhost:8080/api/v1/assets
```

Expected: `me: 200`, `assets: 403`. A company-less session must reach no tenant data.

- [ ] **Step 6: Create a company and enter it**

```bash
curl -s -X POST http://localhost:8080/api/v1/companies \
  -H "Authorization: Bearer $OWNER" -H 'Content-Type: application/json' \
  -d '{"name":"Acme North","tax_id":"ACME-N-001","address":"1 Road"}'
```

Expected: `201`. Then log in again and confirm the new token **does** carry `company_id`, proving `default_company_id` was seeded. Then `GET /api/v1/assets` with it: expect `200`.

- [ ] **Step 7: Prove the isolation boundary holds through the API**

Provision a second account, log in as its owner, create a company, and confirm that owner's `GET /api/v1/companies` shows only their own. Then confirm the first owner cannot switch into the second's company:

```bash
curl -s -o /dev/null -w '%{http_code}\n' -X POST http://localhost:8080/auth/switch-company \
  -H "Authorization: Bearer $OWNER_A" -H 'Content-Type: application/json' \
  -d '{"company_id":<company id belonging to account B>}'
```

Expected: `403`.

- [ ] **Step 8: Regenerate the API docs**

```bash
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

Match the flags CI uses — check `.github/workflows/ci.yml` and use exactly those.

- [ ] **Step 9: Document the onboarding flow in the README**

Add a short section covering: platform admin creates the account, the owner logs in with no company, creates one, and is placed into it. Note that platform staff hold no memberships, and that employee 1 currently holds both roles pending the follow-up split.

- [ ] **Step 10: Commit**

```bash
git add README.md docs/
git commit -m "docs: self-service onboarding flow and regenerated API docs"
```

---

## Self-Review

**Spec coverage:**

| Spec section | Task |
|---|---|
| `account` table | 1 |
| `company.account_id`, `employee.account_id` | 1 |
| Composite-FK isolation invariant | 1 (steps 1, 4) |
| Migration and backfill, ordering | 1 |
| Company-less session, claims shape | 3, 4 |
| Three reachable routes for such a session | 8 step 5 |
| `RequireAccountOwner` new meaning | 2 (query), 5 (identity) |
| Company creation account-scoped, body ignored | 6 |
| `default_company_id` seeded when unset | 6 |
| `POST /admin/accounts` one transaction, password not echoed | 7 |
| `GET /admin/accounts` | 7 |
| `POST /admin/accounts/{id}/set-owner`, membership check | 7 |
| Error handling: company-less 403 | 8 step 5 |
| Testing: all five listed cases | 3, 4, 5, 6, 8 |
| Success criteria | 8 |

Two gaps found and closed while writing:

1. The spec says `GetEmployeeAuthByEmail` must expose the account so `Verify` can distinguish "no account" from "account but no company". That query change was not in any task; it is now folded into Task 4 step 3.
2. The spec's testing section asks for a test that a cross-account membership insert fails. There is no database-backed test harness in this repository, so it cannot be a `go test`. It is Task 1 step 4 — a `psql` assertion that must error — and the task states explicitly that a successful insert means the task is not done.

One spec requirement is deliberately **not** implemented: removing `employee.is_account_owner`. The spec defers it, so no task touches it. The column stays and is no longer read by any gate.

**Placeholder scan:** no TBD, TODO, or "similar to Task N". Two steps say "in the same shape as the existing handler" (Task 7 step 4) and "match the flags CI uses" (Task 8 step 8) — both name the exact file to read rather than leaving the engineer to invent one.

**Type consistency:** `*int64` is used for `Claims.CompanyID`, `Claims.AccountID`, `handler.Identity.CompanyID/AccountID`, `middleware.Identity.AccountID`, and `resolveLoginCompany`'s return, consistently across Tasks 3-6. `Issue(employeeID int64, companyID *int64, accountID *int64, isAdmin bool)` is declared in Task 3 and called that way in Task 4. `middleware.ContextWithIdentity` is introduced in Task 6 and used only there.
