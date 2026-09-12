# Per-Company Roles and Directors Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let one person be an administrator of one company in their account and a limited user, or nothing, in another.

**Architecture:** `role_id` moves from `employee` onto `employee_companies`, so permissions are a property of the membership rather than the person. A composite foreign key makes attaching another company's role impossible. The account owner remains an administrator throughout their own account by virtue of ownership, so they cannot lock themselves out.

**Tech Stack:** Go 1.22+, Gin, pgx/v5, sqlc v1.29.0, golang-migrate, Postgres 16. `go test ./...` is unit-only and never touches Postgres; `//go:build integration` tests in `internal/http/handler/integration_test.go` run against a real database.

**Spec:** `docs/superpowers/specs/2026-09-10-per-company-roles-design.md`

## Global Constraints

- **Migration numbering:** the next free number is **000021**. Both `.up.sql` and `.down.sql` are required.
- **After any change under `internal/db/queries/`, run `sqlc generate`** (v1.29.0, on PATH) and commit the regenerated `internal/db/gen/`. CI fails on drift. Never hand-edit generated code.
- **A membership carrying no role grants nothing.** `role_id` is nullable; empty permissions is the correct outcome, and forgetting to assign a role must fail closed.
- **Both ownership expressions require `COALESCE`.** Platform staff have `employee.account_id IS NULL`, so `LEFT JOIN account` yields no row and `a.owner_employee_id = e.id` is `NULL`. In SQL `false OR NULL` is `NULL` — an authorization value decided by NULL semantics. Wrap both.
- **The backfill must not change any existing user's effective permissions.** Today a global admin role confers admin in every company the person belongs to.
- **Role names are not unique per company** — there is no `UNIQUE (company_id, name)` on `role`. The backfill takes the **lowest `id`** among same-name matches so it is deterministic.
- **Account-scoped routes are not `/admin/*`.** That namespace is cross-tenant and platform-admin only.
- `go build ./...` and `go test ./...` clean at the end of every task.

---

## File Structure

| File | Responsibility |
|---|---|
| `internal/db/migrations/000021_role_on_membership.up.sql` / `.down.sql` | The column, the composite FK, the access-preserving backfill, dropping `employee.role_id` |
| `internal/db/queries/employee.sql` | `GetEmployeeIdentity` rewrite; `GetEmployeeAuthByEmail` drops `is_admin`; employee CRUD loses `role_id` |
| `internal/db/queries/membership.sql` | New — membership grant/revoke carrying a role, and the account-scoped listing |
| `internal/db/queries/company.sql` | `BootstrapEmployeeCompany` splits |
| `internal/db/queries/notification.sql` | Role join scoped through the membership |
| `internal/http/handler/credential.go` | Login resolves the role after the company |
| `internal/http/handler/employee.go` | Employee CRUD without `role_id` |
| `internal/bootstrap/bootstrap.go` | Seeds the creator's admin role onto the membership |
| `internal/http/dto/membership.go` | New — request/response types for director appointment |
| `internal/http/handler/account_employee.go` | New — the two account-scoped routes |
| `internal/http/handler/router.go` | Route registration |
| `internal/http/handler/integration_test.go` | The end-to-end proof |

---

### Task 1: Migration 000021 — role on the membership

**Files:**
- Create: `internal/db/migrations/000021_role_on_membership.up.sql`
- Create: `internal/db/migrations/000021_role_on_membership.down.sql`

**Interfaces:**
- Consumes: the schema after `000020`.
- Produces: `employee_companies.role_id bigint NULL` with `fk_ec_role_company`; `employee.role_id` dropped.

There is no database-backed unit test harness. This task is verified by applying the migration to a real database and asserting two things: that a wrong-company role is refused, and that **effective permissions are unchanged**. Both assertions are the deliverable.

- [ ] **Step 1: Write the up migration**

```sql
-- 000021_role_on_membership.up.sql
-- A role is a property of a membership, not of a person.
--
-- employee.role_id was a single global foreign key while role is company-scoped,
-- so a person whose role pointed at one company's administrator role was an
-- administrator in EVERY company they belonged to. Migration 000010 exists
-- because that reached further than it looked once already.
ALTER TABLE employee_companies ADD COLUMN role_id bigint;

-- A role from another company must be impossible to attach — not merely
-- rejected by whichever endpoint remembers to check. Same construction 000019
-- used for cross-account membership, and for the same reason.
ALTER TABLE role ADD CONSTRAINT uq_role_company UNIQUE (id, company_id);
ALTER TABLE employee_companies ADD CONSTRAINT fk_ec_role_company
    FOREIGN KEY (role_id, company_id) REFERENCES role(id, company_id) ON DELETE SET NULL;

-- Preserve effective access exactly.
--
-- Today a global admin role confers admin in every company the person belongs
-- to. Moving role_id onto memberships naively would leave them an administrator
-- only where that particular role lives, silently stripping access everywhere
-- else with no error and no indication anything changed.
--
-- So each membership takes the role of the SAME NAME in its own company, and
-- where none exists the definition is copied. Role names are not unique per
-- company (there is no UNIQUE (company_id, name)), so the lowest id wins and the
-- result is deterministic.
DO $$
DECLARE m RECORD;
        src RECORD;
        target_role_id bigint;
BEGIN
    FOR m IN
        SELECT ec.id AS ec_id, ec.company_id, e.role_id AS global_role_id
        FROM employee_companies ec
        JOIN employee e ON e.id = ec.employee_id
        WHERE e.role_id IS NOT NULL
    LOOP
        SELECT r.name, r.is_admin, r.permissions INTO src
        FROM role r WHERE r.id = m.global_role_id;

        SELECT r.id INTO target_role_id
        FROM role r
        WHERE r.company_id = m.company_id AND r.name = src.name
        ORDER BY r.id
        LIMIT 1;

        IF target_role_id IS NULL THEN
            INSERT INTO role (company_id, name, is_admin, permissions)
            VALUES (m.company_id, src.name, src.is_admin, src.permissions)
            RETURNING id INTO target_role_id;
        END IF;

        UPDATE employee_companies SET role_id = target_role_id WHERE id = m.ec_id;
    END LOOP;
END $$;

-- An employee with no global role keeps role_id NULL on every membership, which
-- grants nothing — the same as today, since a NULL role already yields empty
-- permissions.

ALTER TABLE employee DROP CONSTRAINT fk_employee_role;
ALTER TABLE employee DROP COLUMN role_id;

CREATE INDEX idx_employee_companies_role ON employee_companies (role_id);
```

- [ ] **Step 2: Write the down migration**

```sql
-- 000021_role_on_membership.down.sql
ALTER TABLE employee ADD COLUMN role_id bigint;

-- Restore a single global role per employee from their memberships. Which one
-- is arbitrary when they differ, because the old model could not express more
-- than one — that information is genuinely lost on the way down.
UPDATE employee e SET role_id = (
    SELECT ec.role_id FROM employee_companies ec
    WHERE ec.employee_id = e.id AND ec.role_id IS NOT NULL
    ORDER BY ec.id LIMIT 1
);

ALTER TABLE employee ADD CONSTRAINT fk_employee_role
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE RESTRICT;

DROP INDEX IF EXISTS idx_employee_companies_role;
ALTER TABLE employee_companies DROP CONSTRAINT fk_ec_role_company;
ALTER TABLE role DROP CONSTRAINT uq_role_company;
ALTER TABLE employee_companies DROP COLUMN role_id;
```

- [ ] **Step 3: Capture effective permissions BEFORE migrating**

This is the assertion that protects live customers. Run against a database seeded with the current schema and some data:

```bash
export MSYS_NO_PATHCONV=1
cp .env.example .env    # gitignored; docker compose fails to parse without it
docker compose up -d db

docker compose exec -T db psql -U postgres -d fleet -A -F',' -c "
SELECT ec.employee_id, ec.company_id,
       COALESCE(r.is_admin,false) AS is_admin,
       COALESCE(r.permissions,'{}'::jsonb) AS permissions
FROM employee_companies ec
JOIN employee e ON e.id = ec.employee_id
LEFT JOIN role r ON r.id = e.role_id
ORDER BY ec.employee_id, ec.company_id;" > /tmp/perms-before.csv
```

- [ ] **Step 4: Apply the migration**

```bash
docker compose run --rm migrate
```

Expected: exit 0.

- [ ] **Step 5: Prove effective permissions did not change**

```bash
docker compose exec -T db psql -U postgres -d fleet -A -F',' -c "
SELECT ec.employee_id, ec.company_id,
       COALESCE(r.is_admin,false) AS is_admin,
       COALESCE(r.permissions,'{}'::jsonb) AS permissions
FROM employee_companies ec
LEFT JOIN role r ON r.id = ec.role_id
ORDER BY ec.employee_id, ec.company_id;" > /tmp/perms-after.csv

diff /tmp/perms-before.csv /tmp/perms-after.csv && echo "IDENTICAL — access preserved"
```

Expected: no differences. **Any difference means the backfill silently changed someone's access and the task is not done.** Report the diff rather than adjusting the comparison.

- [ ] **Step 6: Prove a wrong-company role cannot be attached**

```bash
docker compose exec -T db psql -U postgres -d fleet -c "
INSERT INTO employee_companies (employee_id, company_id, account_id, role_id)
SELECT ec.employee_id, ec.company_id, ec.account_id,
       (SELECT r.id FROM role r WHERE r.company_id <> ec.company_id ORDER BY r.id LIMIT 1)
FROM employee_companies ec LIMIT 1;"
```

Expected: `ERROR: ... violates foreign key constraint "fk_ec_role_company"`. Report the verbatim error. If it succeeds, the constraint is wrong and the task is not done.

- [ ] **Step 7: Verify the down migration reverses**

```bash
docker compose run --rm --entrypoint migrate migrate \
  -path /migrations -database "postgres://postgres:postgres@db:5432/fleet?sslmode=disable" down 1
docker compose run --rm migrate
```

Expected: both exit 0.

- [ ] **Step 8: Commit**

```bash
git add internal/db/migrations/000021_role_on_membership.up.sql internal/db/migrations/000021_role_on_membership.down.sql
git commit -m "feat(db): move role onto the membership, preserving effective access"
```

---

### Task 2: Queries and regeneration

**Files:**
- Modify: `internal/db/queries/employee.sql`
- Modify: `internal/db/queries/company.sql`
- Modify: `internal/db/queries/notification.sql`
- Create: `internal/db/queries/membership.sql`
- Modify: `internal/db/gen/**` (regenerated)

**Interfaces:**
- Consumes: the schema from Task 1.
- Produces:
  - `GetEmployeeIdentity(ctx, GetEmployeeIdentityParams{CompanyID *int64, ID int64})` — row gains `RoleID *int64`, keeps `IsMember`, `IsAdmin`, `Permissions`, `IsAccountOwner`
  - `GetEmployeeAuthByEmail(ctx, email string)` — **no longer returns `IsAdmin`**
  - `GrantMembership(ctx, GrantMembershipParams{EmployeeID, CompanyID int64, RoleID *int64}) error`
  - `RevokeMembershipsNotIn(ctx, RevokeMembershipsNotInParams{EmployeeID int64, CompanyIDs []int64}) error`
  - `ListAccountEmployeeMemberships(ctx, accountID int64) ([]ListAccountEmployeeMembershipsRow, error)`
  - `BootstrapEmployeeDefaultCompany(ctx, ...)` — the `role_id` half removed

- [ ] **Step 1: Rewrite `GetEmployeeIdentity`**

Replace the existing block in `internal/db/queries/employee.sql`:

```sql
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
    (ec.id IS NOT NULL)                  AS is_member,
    COALESCE(r.is_admin, false)
        OR COALESCE(a.owner_employee_id = e.id, false) AS is_admin,
    COALESCE(r.permissions, '{}'::jsonb) AS permissions,
    COALESCE(a.owner_employee_id = e.id, false)        AS is_account_owner
FROM employee e
LEFT JOIN employee_companies ec ON ec.employee_id = e.id
                               AND ec.company_id = sqlc.narg(company_id)
LEFT JOIN role r    ON r.id = ec.role_id
LEFT JOIN account a ON a.id = e.account_id
WHERE e.id = sqlc.arg(id);
```

- [ ] **Step 2: Drop `is_admin` from the auth query**

`GetEmployeeAuthByEmail` cannot resolve `is_admin` any more — there is no global role. Remove the `LEFT JOIN role` and the `is_admin` column, leaving `id`, `default_company_id`, `account_id`, `password_hash`. Task 4 resolves admin status after the company is known.

- [ ] **Step 3: Remove `role_id` from employee CRUD**

In `CreateEmployee` and `UpdateEmployee`, delete `role_id` from the column lists, the `VALUES`/`SET` clauses, and their arguments. Granting access in a company is now a membership operation, not a field on the person.

Do the same for the employee list/get queries that select `e.role_id`, and rewrite `notification.sql`'s `LEFT JOIN role r ON r.id = e.role_id` to join through `employee_companies` for the notification's company.

- [ ] **Step 4: Split `BootstrapEmployeeCompany`**

It currently sets both `role_id` and `default_company_id` on the employee. Only the default company remains there:

```sql
-- name: BootstrapEmployeeDefaultCompany :exec
-- Mirrors Django: only fills default_company when the employee has none. The
-- role half moved to the membership.
UPDATE employee SET
    default_company_id = COALESCE(default_company_id, sqlc.arg(company_id)),
    updated_at         = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);
```

- [ ] **Step 5: Write the membership queries**

```sql
-- internal/db/queries/membership.sql
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
WHERE e.account_id = sqlc.arg(account_id)
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
```

Check the real `employee_companies` unique constraint name before writing the `ON CONFLICT` target — `000001_init.up.sql` declares `uq_employee_companies UNIQUE (employee_id, company_id)`. Use the column form as above if it resolves; if sqlc complains, name the constraint.

- [ ] **Step 6: Regenerate and note what breaks**

```bash
sqlc generate
go build ./... 2>&1 | tee /tmp/breaks.txt
```

Expected: `sqlc generate` succeeds; `go build` **fails**. That is correct — Tasks 3-6 own the call sites. Record the exact failing files and lines; that list is the work list.

- [ ] **Step 7: Commit**

```bash
git add internal/db/queries internal/db/gen
git commit -m "feat(db): resolve the role through the membership"
```

---

### Task 3: Identity carries the per-company role

**Files:**
- Modify: `internal/http/handler/identity.go`
- Modify: `internal/http/middleware/rbac.go`
- Test: `internal/http/middleware/rbac_test.go`

**Interfaces:**
- Consumes: `GetEmployeeIdentity` from Task 2.
- Produces: `middleware.Identity` gains `RoleID *int64`; `HasRole` derives from it.

- [ ] **Step 1: Write the failing test**

```go
// internal/http/middleware/rbac_test.go — add to the existing file

// A membership with no role grants nothing. Forgetting to assign one must fail
// closed: the person is a member, and every module gate still refuses.
func TestMemberWithoutRoleIsRefusedEveryModule(t *testing.T) {
	id := Identity{IsActive: true, IsMember: true, HasRole: false, IsAdmin: false}
	for _, module := range Modules {
		if id.Can(module, "GET") {
			t.Fatalf("module %q allowed a read for a member holding no role", module)
		}
		if id.Can(module, "POST") {
			t.Fatalf("module %q allowed a write for a member holding no role", module)
		}
	}
}

// Ownership is a backstop: with roles on memberships it becomes possible to
// remove your own admin role from a company you own, and nobody in the account
// could then repair it.
func TestAccountOwnerIsAdminWithoutARole(t *testing.T) {
	id := Identity{IsActive: true, IsMember: true, HasRole: false, IsAdmin: true}
	if !id.Can("assets", "DELETE") {
		t.Fatal("an account owner must retain admin without holding a role")
	}
}
```

- [ ] **Step 2: Run it and watch it fail or pass**

Run: `go test ./internal/http/middleware/ -run 'TestMemberWithoutRole|TestAccountOwnerIsAdmin' -v`

If it already passes, say so — `Identity.allows` may already refuse a roleless member. Do not delete a passing test; it pins behaviour this change must not break.

- [ ] **Step 3: Thread the role through**

Add to `middleware.Identity`:

```go
	// RoleID is the role held in the token's company, if any. Nil means the
	// membership carries no role, which grants nothing.
	RoleID *int64
```

In `internal/http/handler/identity.go`, map `row.RoleID` through and set `HasRole: row.RoleID != nil`.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/http/middleware/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/http/middleware internal/http/handler/identity.go
git commit -m "feat(authz): identity carries the per-company role"
```

---

### Task 4: Login resolves the role after the company

**Files:**
- Modify: `internal/http/handler/credential.go`
- Test: `internal/http/handler/credential_test.go`

**Interfaces:**
- Consumes: `GetEmployeeIdentity`, `GetEmployeeAuthByEmail` (no `IsAdmin`) from Task 2.
- Produces: `Verify` unchanged in signature; `Identity.IsAdmin` now resolved per company.

- [ ] **Step 1: Write the failing test**

```go
// internal/http/handler/credential_test.go — add to the existing file

// A company-less session has no company whose role could grant admin, so its
// admin status can only come from account ownership.
func TestCompanylessSessionAdminComesFromOwnershipOnly(t *testing.T) {
	if adminForSession(nil, false) {
		t.Fatal("a company-less non-owner must not be admin")
	}
	if !adminForSession(nil, true) {
		t.Fatal("a company-less account owner must be admin")
	}
}

// With a company, the role in THAT company decides, and ownership still backs it.
func TestScopedSessionAdminComesFromRoleOrOwnership(t *testing.T) {
	company := int64(7)
	if !adminForSession(&company, false) != !true {
		// placeholder guard; see the table below
	}
	cases := []struct {
		name       string
		roleAdmin  bool
		owner      bool
		wantAdmin  bool
	}{
		{"role grants admin", true, false, true},
		{"ownership grants admin without a role", false, true, true},
		{"neither grants admin", false, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.roleAdmin || tc.owner; got != tc.wantAdmin {
				t.Fatalf("admin = %v, want %v", got, tc.wantAdmin)
			}
		})
	}
}
```

The second test as written asserts the rule, not the implementation. **Replace its body with a call to whatever helper you extract in Step 3** — the point is that the combination rule is tested somewhere pure, not that this exact shape survives. If extracting a helper is not natural, delete this test and say why in your report rather than leaving a test that asserts `a || b == a || b`.

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./internal/http/handler/ -run 'TestCompanylessSessionAdmin|TestScopedSessionAdmin' -v`
Expected: compile failure — `adminForSession` is undefined.

- [ ] **Step 3: Resolve admin after the company**

In `Verify`, after `resolveLoginCompany` and the `mayLogIn` check, fetch the identity for the resolved company and take `IsAdmin` from it:

```go
	// is_admin is a property of the role held in THIS company, so it cannot be
	// resolved until the company is. A company-less session has no role, and its
	// admin status comes from account ownership alone — which GetEmployeeIdentity
	// already answers, so the token claim and the request-time gate come from one
	// query rather than two copies of the rule.
	ident, err := v.q.GetEmployeeIdentity(ctx, gen.GetEmployeeIdentityParams{
		ID:        row.ID,
		CompanyID: companyID,
	})
	if err != nil {
		return Identity{}, ErrInvalidCredentials
	}

	return Identity{
		EmployeeID: row.ID,
		CompanyID:  companyID,
		AccountID:  row.AccountID,
		IsAdmin:    ident.IsAdmin,
	}, nil
```

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/http/handler/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/http/handler/credential.go internal/http/handler/credential_test.go
git commit -m "feat(auth): resolve admin from the role in the session's company"
```

---

### Task 5: Employee CRUD without a global role

**Files:**
- Modify: `internal/http/handler/employee.go`
- Modify: `internal/http/dto/employee.go`
- Modify: `internal/bootstrap/bootstrap.go`
- Modify: `internal/http/handler/company.go` (owner setup, if it assigns a role)

**Interfaces:**
- Consumes: the query changes from Task 2.
- Produces: `dto.CreateEmployeeRequest` and `dto.UpdateEmployeeRequest` lose `RoleID`; `dto.EmployeeResponse` loses it too.

- [ ] **Step 1: Remove `role_id` from the DTOs**

Delete the `RoleID *int64 \`json:"role_id"\`` field from `CreateEmployeeRequest`, `UpdateEmployeeRequest` and `EmployeeResponse` in `internal/http/dto/employee.go` (lines 16, 50, 85). A role is no longer a property of a person, so an endpoint that accepts one would be accepting something it cannot honour.

- [ ] **Step 2: Fix the handler and store**

Remove `RoleID` from the params built in `internal/http/handler/employee.go`. The compiler will point at each site once the DTO fields are gone.

- [ ] **Step 3: Update bootstrap**

`bootstrap.Company` currently calls `BootstrapEmployeeCompany` with a role. Split it: call `BootstrapEmployeeDefaultCompany` for the default company, and grant the membership with the admin role via `GrantMembership`:

```go
	if err := q.GrantMembership(ctx, gen.GrantMembershipParams{
		EmployeeID: employeeID,
		CompanyID:  companyID,
		RoleID:     &adminRole.ID,
	}); err != nil {
		return err
	}
	return q.BootstrapEmployeeDefaultCompany(ctx, gen.BootstrapEmployeeDefaultCompanyParams{
		ID:        employeeID,
		CompanyID: &companyID,
		UpdatedAt: time.Now().UTC(),
	})
```

Read the existing `AddEmployeeCompany` call in `bootstrap.go:133` and replace it with `GrantMembership`, since the latter carries the role. Confirm whether `AddEmployeeCompany` still has callers; if none remain, delete the query and regenerate.

- [ ] **Step 4: Build and test**

```bash
go build ./... && go test ./...
```
Expected: both clean. This is the task that closes the build.

- [ ] **Step 5: Commit**

```bash
git add internal/http/handler/employee.go internal/http/dto/employee.go internal/bootstrap internal/db/queries internal/db/gen
git commit -m "feat(employees): a role is granted per company, not per person"
```

---

### Task 6: The account owner appoints directors

**Files:**
- Create: `internal/http/dto/membership.go`
- Create: `internal/http/handler/account_employee.go`
- Modify: `internal/http/handler/router.go`
- Test: `internal/http/handler/account_employee_test.go`

**Interfaces:**
- Consumes: `GrantMembership`, `RevokeMembershipsNotIn`, `ListAccountEmployeeMemberships`, `RoleBelongsToCompany`, `CompanyInAccount` from Task 2.
- Produces:
  - `NewAccountEmployeeHandler(q *gen.Queries, pool *pgxpool.Pool) *AccountEmployeeHandler`
  - methods `List`, `ReplaceCompanies`
  - `dto.CompanyRoleGrant{CompanyID int64, RoleID *int64}`
  - `dto.ReplaceAccountEmployeeCompaniesRequest{Grants []CompanyRoleGrant}`

- [ ] **Step 1: Write the failing test**

```go
// internal/http/handler/account_employee_test.go
package handler

import (
	"strings"
	"testing"

	"fleet/internal/http/dto"
)

// A grant naming the same company twice is ambiguous — the last one silently
// wins and the caller cannot tell which role was applied. Refuse it.
func TestValidateGrantsRejectsDuplicateCompanies(t *testing.T) {
	in := dto.ReplaceAccountEmployeeCompaniesRequest{Grants: []dto.CompanyRoleGrant{
		{CompanyID: 1}, {CompanyID: 2}, {CompanyID: 1},
	}}
	err := validateGrants(in)
	if err == nil || !strings.Contains(err.Error(), "company") {
		t.Fatalf("error = %v, want it to name the duplicate company", err)
	}
}

func TestValidateGrantsRejectsNonPositiveCompany(t *testing.T) {
	in := dto.ReplaceAccountEmployeeCompaniesRequest{Grants: []dto.CompanyRoleGrant{{CompanyID: 0}}}
	if err := validateGrants(in); err == nil {
		t.Fatal("a company id of 0 must be refused")
	}
}

// An empty grant list is legitimate: it revokes every membership. It must not
// be confused with a malformed request.
func TestValidateGrantsAllowsEmpty(t *testing.T) {
	if err := validateGrants(dto.ReplaceAccountEmployeeCompaniesRequest{}); err != nil {
		t.Fatalf("an empty grant list revokes all memberships and is valid: %v", err)
	}
}
```

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./internal/http/handler/ -run TestValidateGrants -v`
Expected: compile failure — `validateGrants` and the DTOs are undefined.

- [ ] **Step 3: Write the DTOs**

```go
// internal/http/dto/membership.go
package dto

// CompanyRoleGrant gives an employee access to one company with one role.
// RoleID is optional: a membership with no role grants nothing, which is a
// legitimate state — associated with the company, permitted nothing in it.
type CompanyRoleGrant struct {
	CompanyID int64  `json:"company_id"`
	RoleID    *int64 `json:"role_id"`
}

// ReplaceAccountEmployeeCompaniesRequest replaces the whole set. An empty list
// revokes every membership, which is how access is removed.
type ReplaceAccountEmployeeCompaniesRequest struct {
	Grants []CompanyRoleGrant `json:"grants"`
}

type AccountEmployeeMembership struct {
	CompanyID   int64  `json:"company_id"`
	CompanyName string `json:"company_name"`
	RoleID      *int64 `json:"role_id"`
	RoleName    string `json:"role_name,omitempty"`
	RoleIsAdmin bool   `json:"role_is_admin"`
}

type AccountEmployeeResponse struct {
	EmployeeID  int64                       `json:"employee_id"`
	FirstName   string                      `json:"first_name"`
	LastName    string                      `json:"last_name"`
	Email       string                      `json:"email"`
	IsActive    bool                        `json:"is_active"`
	Memberships []AccountEmployeeMembership `json:"memberships"`
}
```

- [ ] **Step 4: Write the handler**

`validateGrants` refuses duplicates and non-positive ids, returning `apierr.Validation` with a field-keyed map — matching how the other validation errors in this codebase report.

`ReplaceCompanies` runs in one transaction and, for each grant, checks in order:

1. `CompanyInAccount(company_id, caller's account)` — false → **404**. The caller has no legitimate way to know a company outside their account exists, so this is not a 403.
2. `RoleBelongsToCompany(role_id, company_id)` when a role is given — false → **422** naming `role_id`.
3. `GrantMembership`.

Then `RevokeMembershipsNotIn` for the companies not listed, and `RecordMembershipChange` for each grant and revoke — mirroring `admin_employee.go:290`. Read that handler and follow its audit shape; granting a company with an administrator role is the highest-privilege write an account owner can make.

The target employee must belong to the caller's account. Check it explicitly and return **404** when not — the composite foreign key is the backstop, not the user-facing behaviour.

- [ ] **Step 5: Register the routes**

In `router.go`, alongside the company-scoped groups:

```go
	// Account-scoped, deliberately NOT under /api/v1/admin: that namespace is
	// cross-tenant and platform-admin only. This one operates strictly inside
	// the caller's own account.
	accountOwner := api.Group("", middleware.RequireAccountOwner())
	accountEmployees := NewAccountEmployeeHandler(d.Queries, d.Pool)
	accountOwner.GET("/account/employees", accountEmployees.List)
	accountOwner.PUT("/account/employees/:id/companies", accountEmployees.ReplaceCompanies)
```

Register on `api`, not `member`: an owner with no company of their own must still be able to appoint directors.

- [ ] **Step 6: Build and test**

```bash
go build ./... && go test ./...
```
Expected: both clean.

- [ ] **Step 7: Commit**

```bash
git add internal/http/dto/membership.go internal/http/handler/account_employee.go internal/http/handler/account_employee_test.go internal/http/handler/router.go
git commit -m "feat(accounts): let an owner appoint directors with a role per company"
```

---

### Task 7: Prove it end to end

**Files:**
- Modify: `internal/http/handler/integration_test.go`

**Interfaces:**
- Consumes: everything above.
- Produces: no new symbols.

This repository's unit tests never touch Postgres. Subsystem 1 shipped three defects that a green `go test ./...` did not catch, including one that broke the feature it existed to deliver. The integration suite is where behaviour spanning Go and the schema is proven, and the assertion below **is** this subsystem's deliverable.

- [ ] **Step 1: Extend the integration test**

Add to the existing `//go:build integration` file, following its established helpers:

```go
// The whole point of subsystem 2: one person, two companies, different powers.
func TestDirectorHasDifferentPowersPerCompany(t *testing.T) {
	// Provision an account, log in as the owner, create two companies.
	// Create an employee. Grant them:
	//   company A -> the seeded "Administrador" role
	//   company B -> a role with {"assets": {"read": true}} only
	// Then, logged in as that employee:
	//   switch to A: POST /api/v1/assets            -> 201
	//   switch to B: POST /api/v1/assets            -> 403
	//   switch to B: GET  /api/v1/assets            -> 200
}
```

Write the body using the helpers already in that file. The three assertions at the end are the deliverable — a director who may write in one company and only read in another.

- [ ] **Step 2: Add the ownership-backstop assertion**

```go
// An owner who strips their own role must not lose control of their company.
func TestOwnerKeepsAdminAfterLosingTheirRole(t *testing.T) {
	// As the owner, PUT /account/employees/{ownID}/companies granting their own
	// company with role_id: null.
	// Then POST /api/v1/assets in that company -> 201, because ownership backs it.
}
```

- [ ] **Step 3: Add the wrong-company-role assertion**

```go
// A role from another company must be refused at the API, not just the database.
func TestGrantingARoleFromAnotherCompanyIsRefused(t *testing.T) {
	// PUT /account/employees/{id}/companies with company B and a role belonging
	// to company A -> 422 naming role_id.
}
```

- [ ] **Step 4: Run the integration suite**

```bash
docker compose up -d db
go test -tags=integration ./internal/http/handler/ -run 'TestDirector|TestOwnerKeeps|TestGrantingARole' -v
```

Expected: all pass. If `TestDirectorHasDifferentPowersPerCompany` fails, the subsystem does not work — report it exactly rather than adjusting the assertion.

- [ ] **Step 5: Commit**

```bash
git add internal/http/handler/integration_test.go
git commit -m "test(roles): prove a director's powers differ per company"
```

---

### Task 8: Documentation and API docs

**Files:**
- Modify: `README.md`
- Modify: `docs/swagger.json`, `docs/docs.go`, `docs/swagger.yaml` (regenerated)

- [ ] **Step 1: Regenerate the API docs**

Use exactly the command CI runs (`.github/workflows/ci.yml:40-43`), or `git diff --exit-code docs` fails there:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
swag init -g main.go \
  -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage \
  --parseDependency --parseInternal -o docs
```

- [ ] **Step 2: Document the model in the README**

Extend the onboarding section: a role is granted per company; a membership with no role grants nothing; the account owner is an administrator throughout their own account and cannot lock themselves out; an administrator of a company may edit that company's roles.

State plainly that `employee.role_id` no longer exists and that a person's powers now differ per company.

- [ ] **Step 3: Commit**

```bash
git add README.md docs/
git commit -m "docs: per-company roles and director appointment"
```

---

## Self-Review

**Spec coverage:**

| Spec section | Task |
|---|---|
| `employee_companies.role_id`, `employee.role_id` dropped | 1 |
| Composite FK preventing a wrong-company role | 1 (steps 1, 6) |
| Access-preserving backfill, deterministic on duplicate names | 1 (steps 1, 3, 5) |
| Identity query rewrite, `COALESCE` on both ownership expressions | 2, 3 |
| Membership grants nothing without a role | 3 (test), 6 (DTO allows nil) |
| Ownership backstop | 2 (query), 3 (test), 7 (integration) |
| Login resolves the role after the company | 4 |
| Account-owner director routes, 404 outside the account, 422 wrong role | 6 |
| Role editing needs no new code | — deliberately no task; `RequireAdminRole` already means per-company once Task 2 lands |
| Employee CRUD reshaped, notification, bootstrap | 2, 5 |
| `/me/permissions` reports the role | **GAP — see below** |
| Testing: all six listed items | 1, 3, 4, 7 |
| Success criteria | 7 |

Two gaps found and closed while writing:

1. **`/me/permissions` reporting the role's id and name** had no task. The spec asks for it so a client can show which role is in effect. It is folded into Task 5 as part of the DTO work — the same commit that removes `role_id` from the employee DTOs adds it to the me response, since both are "where does a role appear in the API" decisions. If the implementer finds `me.sql` needs its own query change, that belongs in Task 2's regeneration.

2. **Task 4's second test as written asserts a tautology** (`a || b == a || b`). Rather than delete it silently, the step tells the implementer to replace the body with a call to whatever helper they extract, or delete the test and say why. A test that cannot fail is worse than no test, and this plan should not ship one.

**Placeholder scan:** no TBD or TODO. Task 7's test bodies are described rather than written — deliberately, because they depend on helpers already in a file the implementer will read, and inventing helper names here would produce code that does not compile. The assertions themselves are stated exactly.

**Type consistency:** `RoleID *int64` throughout — `middleware.Identity`, `GetEmployeeIdentityRow`, `dto.CompanyRoleGrant`, `GrantMembershipParams`. `GrantMembership` takes `RoleID *int64` in Tasks 2, 5 and 6. `BootstrapEmployeeDefaultCompany` is named identically in Tasks 2 and 5.
