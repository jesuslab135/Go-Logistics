# Backend Task List Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close items 1, 2, 3, 4, 5, 9, 13 and 14 of `backend-engineer-task-todo.md` — the tenancy/login security hole, the deployed CORS allowlist, the admin cross-company namespace (employee memberships, company owner, provisioning verification), the warranty `provider_name` denormalization, and the two swagger-annotation hygiene gaps.

**Architecture:** All new cross-tenant capability lands under a single new `/api/v1/admin/...` namespace gated by `middleware.RequireAdminRole()`, registered through one new `registerAdminRoutes(api, member, d)` function modeled on the existing `registerCompanyRoutes`. Every decision that can be expressed as a pure function (login-company resolution, `default_company_id` validation, membership-replace validation) is extracted into a testable Go function so it can be TDD'd without a database — matching the repo's existing `deriveFuel` / `createEmployeeParams` precedent. SQL changes go through sqlc.

**Tech Stack:** Go 1.26.5, gin v1.10.1, pgx/v5 + sqlc v1.29.0, swaggo/swag v1.16.4, stdlib `testing` (no testify).

**Spec:** `C:\Users\jesus.olmos\Downloads\backend-engineer-task-todo.md` (items 1, 2, 3, 4, 5, 9, 13, 14). Items 6, 7, 8, 10, 11, 12 are explicitly out of scope for this plan — they need a product decision first.

## Global Constraints

- **Repo root:** `C:\Users\jesus.olmos\Desktop\PROJECTS\GoLangLogistics\GoLangLogisticsFinal`. Module name is `fleet`.
- **Shell is PowerShell**, but the `Bash` tool (Git Bash) is available and preferred for `go`/`sqlc`/`swag` invocations.
- **TDD is mandatory.** Every task writes a failing test first, runs it to confirm it fails for the stated reason, then writes the minimum implementation, then re-runs.
- **Local gate, must pass before every commit:**
  `go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...`
- **No testify.** Plain stdlib `testing`, table-driven with a `name` field, `t.Helper()` in helpers, `t.Fatalf` for setup failures and `t.Errorf` for continuable assertions, messages formatted `"got %v, want %v"`.
- **No test may require a live Postgres.** Zero existing tests do; keep it that way.
- **sqlc regeneration:** `sqlc generate` from the repo root (config `sqlc.yaml`). Binary is already installed at `C:\Users\jesus.olmos\go\bin\sqlc.exe`. Never hand-edit `internal/db/gen/**`.
- **swag regeneration:**
  ```sh
  swag init -g main.go \
    -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage \
    --parseDependency --parseInternal -o docs
  ```
  Binary already installed at `C:\Users\jesus.olmos\go\bin\swag.exe`. Never hand-edit `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`.
- **Error envelope:** every failure goes through `apierr.Abort(c, err)`. Wire shape is `{"error":{"code":"...","message":"...","details":{...}}}`.
- **Tenant scoping:** stores read the tenant from `middleware.CompanyFromContext(ctx)` and the actor from `middleware.EmployeeFromContext(ctx)`. Never from the URL, except on the new `/admin/companies/{id}/...` routes where the company id *is* the path parameter — that is the whole point of those routes.
- **Every new route must be appended to the expected-routes list in `internal/http/handler/router_test.go`** (currently lines 36-74). `TestAPIRoutesRequireAuth` then covers it automatically.
- **Every new list route needs a concrete `dto.XPage` struct** — swag cannot express generics.
- **Commit style:** conventional commits (`fix(auth): …`, `feat(admin): …`, `docs(swagger): …`), one commit per task. Work on a branch, never commit to `main` directly.

---

## Branch setup (do this once, before Task 1)

```bash
cd "C:/Users/jesus.olmos/Desktop/PROJECTS/GoLangLogistics/GoLangLogisticsFinal"
git checkout -b feat/backend-task-list-2026-08-20
```

---

## File Structure

| File | Responsibility | Tasks |
|---|---|---|
| `internal/db/queries/employee.sql` | + `ListEmployeeCompanyIDs`, `ListAllEmployees`, `CountAllEmployees`, `GetEmployeeByID`, `ReplaceEmployeeCompanies`, `SetEmployeeDefaultCompany` | 1, 2, 3 |
| `internal/db/queries/company.sql` | + `GetCompanyByID`, `CountCompaniesByIDs`, `GetCompanyOwner`, `ClearCompanyAccountOwner`, `SetCompanyAccountOwner` | 3, 4 |
| `internal/db/queries/warranty.sql` | join `vendor` for `provider_name` | 6 |
| `internal/http/handler/credential.go` | login company resolved through `employee_companies` | 1 |
| `internal/http/handler/credential_test.go` | **new** — table tests for `resolveLoginCompany` | 1 |
| `internal/http/handler/auth.go` | `ErrNoCompanyMembership` + 403 mapping in `Login` | 1 |
| `internal/http/handler/employee.go` | `default_company_id` validation on Create/Update | 2 |
| `internal/http/handler/employee_defaults_test.go` | **new** — table tests for `validateDefaultCompany` | 2 |
| `internal/http/handler/admin_employee.go` | **new** — admin employee list + membership handlers | 3 |
| `internal/http/handler/admin_employee_test.go` | **new** — table tests for `normalizeCompanyIDs` / `validateMembershipReplace` | 3 |
| `internal/http/handler/admin_company.go` | **new** — company owner read/set + cross-company roles/statuses | 4, 5 |
| `internal/http/dto/admin.go` | **new** — admin request/response/page DTOs | 3, 4, 5 |
| `internal/http/handler/router.go` | + `registerAdminRoutes(api, member, d)` | 3, 4, 5 |
| `internal/http/handler/router_test.go` | + new routes in the expected list | 3, 4, 5 |
| `internal/http/dto/warranty.go` | + `ProviderName` | 6 |
| `internal/http/handler/warranty.go` | map the joined row; re-read on Create/Update | 6 |
| `internal/http/handler/docs_link.go` | delete 12 duplicated `@Param` lines | 7 |
| `internal/http/handler/docs.go`, `docs_catalog.go` | `page`/`page_size` → `limit`/`offset` | 7 |
| `internal/http/handler/openapi_spec_test.go` | **new** — regression guard on the generated spec | 7 |
| `internal/http/middleware/cors_test.go` | **new** — first coverage of the CORS middleware | 8 |
| `.env.example`, `docs/2026-08-14_golanglogistics-deployment.md` | concrete `CORS_ORIGINS` guidance | 8 |
| `C:\Users\jesus.olmos\Downloads\backend-engineer-task-todo.md` | per-item `**Fixed:**` / `**Frontend:**` annotations | 9 |

---

### Task 1: Login resolves its company through `employee_companies` (spec item 1, part 1)

**Problem:** `EmployeeCredentialVerifier.Verify` takes `row.DefaultCompanyID` straight from the `employee` row with no membership check. If it is nil, `companyID` stays `0` and login "succeeds" with a token scoped to a company that does not exist.

**Files:**
- Modify: `internal/db/queries/employee.sql` (append after `AddEmployeeCompany`, line 72)
- Modify: `internal/http/handler/credential.go`
- Modify: `internal/http/handler/auth.go:56-80` (`Login`)
- Test: `internal/http/handler/credential_test.go` (create)

**Interfaces:**
- Produces: `func resolveLoginCompany(defaultCompanyID *int64, memberships []int64) (int64, bool)` — package `handler`, unexported. Returns the company id to scope the session to and `false` when the employee has no membership at all.
- Produces: `var ErrNoCompanyMembership = errors.New("no company membership")` in `internal/http/handler/auth.go`.
- Produces (sqlc): `func (q *Queries) ListEmployeeCompanyIDs(ctx context.Context, employeeID int64) ([]int64, error)`.

**Design note:** the resolution rule is deliberately implemented as a *pure Go function* over the membership list rather than as clever SQL, because the repo has no database test harness and TDD is mandatory. `deriveFuel` and `createEmployeeParams` set this precedent.

**Rule:**
1. `default_company_id` is used **only** when it appears in the membership list.
2. Otherwise fall back to the lowest-id real membership (deterministic).
3. No memberships at all → refuse login.

- [ ] **Step 1: Write the failing test**

Create `internal/http/handler/credential_test.go`:

```go
package handler

import "testing"

func ptr(v int64) *int64 { return &v }

// A login token is only safe if its company_id claim names a company the
// employee is actually a member of: every downstream query trusts that claim.
func TestResolveLoginCompany(t *testing.T) {
	tests := []struct {
		name        string
		defaultID   *int64
		memberships []int64
		want        int64
		wantOK      bool
	}{
		{
			name:        "default company that is a real membership wins",
			defaultID:   ptr(7),
			memberships: []int64{3, 7, 9},
			want:        7,
			wantOK:      true,
		},
		{
			name:        "nil default falls back to the lowest membership",
			defaultID:   nil,
			memberships: []int64{9, 3, 7},
			want:        3,
			wantOK:      true,
		},
		{
			name:        "default naming a company the employee does not belong to is ignored",
			defaultID:   ptr(42),
			memberships: []int64{9, 3},
			want:        3,
			wantOK:      true,
		},
		{
			name:        "no membership at all refuses the login",
			defaultID:   ptr(7),
			memberships: nil,
			want:        0,
			wantOK:      false,
		},
		{
			name:        "no membership and no default refuses the login",
			defaultID:   nil,
			memberships: []int64{},
			want:        0,
			wantOK:      false,
		},
		{
			name:        "a zero default is never trusted",
			defaultID:   ptr(0),
			memberships: []int64{5},
			want:        5,
			wantOK:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resolveLoginCompany(tt.defaultID, tt.memberships)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("company = %d, want %d", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
cd "C:/Users/jesus.olmos/Desktop/PROJECTS/GoLangLogistics/GoLangLogisticsFinal"
go test ./internal/http/handler/ -run TestResolveLoginCompany -v
```
Expected: compile failure — `undefined: resolveLoginCompany`.

- [ ] **Step 3: Add the SQL query**

Append to `internal/db/queries/employee.sql`:

```sql
-- name: ListEmployeeCompanyIDs :many
-- The companies an employee actually belongs to. Login resolves its company_id
-- claim through this rather than trusting employee.default_company_id, which is
-- a plain writable field and can name a company the employee is not a member of.
SELECT company_id FROM employee_companies
WHERE employee_id = sqlc.arg(employee_id)
ORDER BY company_id;
```

Then regenerate:
```bash
sqlc generate
```

- [ ] **Step 4: Write the minimal implementation**

Replace the body of `Verify` in `internal/http/handler/credential.go` (keep the imports tidy — `slices` is needed):

```go
func (v *EmployeeCredentialVerifier) Verify(ctx context.Context, email, password string) (Identity, error) {
	row, err := v.q.GetEmployeeAuthByEmail(ctx, email)
	if err != nil {
		// Unknown email (ErrNoRows) or lookup failure — do not distinguish, to
		// avoid leaking which emails exist.
		return Identity{}, ErrInvalidCredentials
	}
	if !auth.CheckPassword(row.PasswordHash, password) {
		return Identity{}, ErrInvalidCredentials
	}

	// default_company_id is a plain writable field, so it is a hint, not an
	// authority: the session is scoped to a company only after employee_companies
	// confirms the membership — the same check /auth/switch-company already makes.
	memberships, err := v.q.ListEmployeeCompanyIDs(ctx, row.ID)
	if err != nil {
		return Identity{}, ErrInvalidCredentials
	}
	companyID, ok := resolveLoginCompany(row.DefaultCompanyID, memberships)
	if !ok {
		return Identity{}, ErrNoCompanyMembership
	}

	return Identity{EmployeeID: row.ID, CompanyID: companyID, IsAdmin: row.IsAdmin}, nil
}

// resolveLoginCompany picks the company a session is scoped to. The employee's
// default_company_id wins when it names a real membership; otherwise the lowest
// membership id does, so the choice is deterministic across logins. An employee
// with no membership resolves to nothing and must not be issued a token.
func resolveLoginCompany(defaultCompanyID *int64, memberships []int64) (int64, bool) {
	if len(memberships) == 0 {
		return 0, false
	}
	if defaultCompanyID != nil && slices.Contains(memberships, *defaultCompanyID) {
		return *defaultCompanyID, true
	}
	return slices.Min(memberships), true
}
```

- [ ] **Step 5: Add the login error and its HTTP mapping**

In `internal/http/handler/auth.go`, next to `ErrInvalidCredentials` (line 19):

```go
// ErrNoCompanyMembership is returned when credentials are valid but the employee
// belongs to no company. Issuing a token here would scope the session to a
// company that does not exist, so the login is refused instead.
var ErrNoCompanyMembership = errors.New("no company membership")
```

In `Login`, replace the single error branch after `h.verifier.Verify(...)`:

```go
	id, err := h.verifier.Verify(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrNoCompanyMembership) {
			apierr.Abort(c, apierr.New(http.StatusForbidden, "no_company_membership",
				"this account is not a member of any company"))
			return
		}
		apierr.Abort(c, apierr.Unauthorized("invalid credentials"))
		return
	}
```

Add `@Failure 403 {object} dto.ErrorResponse` to the `Login` swagger annotation block.

- [ ] **Step 6: Run the tests and the gate**

```bash
go test ./internal/http/handler/ -run TestResolveLoginCompany -v
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
```
Expected: `TestResolveLoginCompany` PASS, all other packages still `ok`.

- [ ] **Step 7: Regenerate the OpenAPI spec and commit**

```bash
swag init -g main.go -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage --parseDependency --parseInternal -o docs
git add internal/db/queries/employee.sql internal/db/gen internal/http/handler/credential.go internal/http/handler/credential_test.go internal/http/handler/auth.go docs
git commit -m "fix(auth): resolve the login company through employee_companies

Login took company_id straight from employee.default_company_id, a plain
writable field, with no membership check — a nil value issued a token scoped
to company 0, and a cross-tenant value issued one scoped to a company the
employee does not belong to. The session now resolves through the same
employee_companies lookup /auth/switch-company already uses, and an employee
with no membership is refused with 403 no_company_membership."
```

---

### Task 2: Reject a `default_company_id` with no membership row (spec item 1, part 2)

**Problem:** `default_company_id` is a plain writable field on `CreateEmployeeRequest`/`UpdateEmployeeRequest` and `EmployeeStore.Update` passes it through unvalidated, so an admin of company A can point an employee at company B.

**Files:**
- Modify: `internal/http/handler/employee.go:53-116` (`Create`, `Update`)
- Test: `internal/http/handler/employee_defaults_test.go` (create)

**Interfaces:**
- Consumes: `ListEmployeeCompanyIDs` from Task 1.
- Produces: `func validateDefaultCompany(defaultCompanyID *int64, memberships []int64) error` — package `handler`, unexported. Returns `nil` or a 422 `*apierr.Error` whose `Details` is `map[string]string{"default_company_id": "..."}`.

**Rule:** `default_company_id` must be `null`, or a company id present in the employee's membership set. On **Create** the membership set is exactly `[callerCompany]`, because `Create` inserts precisely that one `employee_companies` row in the same transaction. On **Update** it is whatever `ListEmployeeCompanyIDs` returns.

- [ ] **Step 1: Write the failing test**

Create `internal/http/handler/employee_defaults_test.go`:

```go
package handler

import (
	"errors"
	"testing"

	"fleet/internal/platform/apierr"
)

// default_company_id is what login scopes a session to, so accepting one the
// employee has no employee_companies row for would hand them a token for a
// tenant they are not a member of.
func TestValidateDefaultCompany(t *testing.T) {
	tests := []struct {
		name        string
		defaultID   *int64
		memberships []int64
		wantErr     bool
	}{
		{name: "null is always allowed", defaultID: nil, memberships: []int64{4}, wantErr: false},
		{name: "a real membership is allowed", defaultID: ptr(4), memberships: []int64{4, 9}, wantErr: false},
		{name: "a company the employee does not belong to is rejected", defaultID: ptr(8), memberships: []int64{4, 9}, wantErr: true},
		{name: "any value is rejected when there are no memberships", defaultID: ptr(4), memberships: nil, wantErr: true},
		{name: "zero is rejected", defaultID: ptr(0), memberships: []int64{4}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDefaultCompany(tt.defaultID, tt.memberships)
			if tt.wantErr == (err == nil) {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				return
			}
			var ae *apierr.Error
			if !errors.As(err, &ae) {
				t.Fatalf("error is %T, want *apierr.Error", err)
			}
			if ae.Status != 422 {
				t.Errorf("status = %d, want 422", ae.Status)
			}
			details, ok := ae.Details.(map[string]string)
			if !ok || details["default_company_id"] == "" {
				t.Errorf("details = %#v, want a default_company_id entry", ae.Details)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
go test ./internal/http/handler/ -run TestValidateDefaultCompany -v
```
Expected: compile failure — `undefined: validateDefaultCompany`.

- [ ] **Step 3: Write the minimal implementation**

Add to `internal/http/handler/employee.go` (imports: `slices`, `fleet/internal/platform/apierr`):

```go
// validateDefaultCompany rejects a default_company_id the employee has no
// employee_companies row for. Login scopes a session to this value, so an
// unvalidated one is a cross-tenant token waiting to be issued.
func validateDefaultCompany(defaultCompanyID *int64, memberships []int64) error {
	if defaultCompanyID == nil {
		return nil
	}
	if *defaultCompanyID > 0 && slices.Contains(memberships, *defaultCompanyID) {
		return nil
	}
	return apierr.Validation(map[string]string{
		"default_company_id": "must be a company this employee belongs to",
	})
}
```

- [ ] **Step 4: Wire it into `Create`**

In `EmployeeStore.Create`, before `tx, err := s.pool.Begin(ctx)`:

```go
	company := middleware.CompanyFromContext(ctx)

	// Create grants exactly one membership — the caller's own company — so that
	// is the only default_company_id this request can legitimately set.
	if err := validateDefaultCompany(in.DefaultCompanyID, []int64{company}); err != nil {
		return dto.EmployeeResponse{}, err
	}
```
(The existing `company := middleware.CompanyFromContext(ctx)` line moves up here rather than being duplicated.)

- [ ] **Step 5: Wire it into `Update`**

In `EmployeeStore.Update`, before building `gen.UpdateEmployeeParams`:

```go
	memberships, err := s.q.ListEmployeeCompanyIDs(ctx, id)
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	if err := validateDefaultCompany(in.DefaultCompanyID, memberships); err != nil {
		return dto.EmployeeResponse{}, err
	}
```

- [ ] **Step 6: Run the tests and the gate**

```bash
go test ./internal/http/handler/ -run 'TestValidateDefaultCompany|TestCreateEmployee' -v
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
```
Expected: PASS. `defaults_test.go` exercises `createEmployeeParams` directly and is unaffected.

- [ ] **Step 7: Commit**

```bash
git add internal/http/handler/employee.go internal/http/handler/employee_defaults_test.go
git commit -m "fix(employees): reject a default_company_id with no membership row

An admin of company A could set an employee's default_company_id to company B
and, after the login fix, that value is simply ignored — but writing it at all
misrepresents the employee's tenancy. Create validates against the single
membership it grants; Update validates against the employee's real memberships.
Both return 422 with a default_company_id field detail."
```

---

### Task 3: Admin employee list and company-membership management (spec item 3)

**Files:**
- Modify: `internal/db/queries/employee.sql`, `internal/db/queries/company.sql`
- Create: `internal/http/dto/admin.go`
- Create: `internal/http/handler/admin_employee.go`
- Create: `internal/http/handler/admin_employee_test.go`
- Modify: `internal/http/handler/router.go`, `internal/http/handler/router_test.go`

**Interfaces:**
- Consumes: `ListEmployeeCompanyIDs` (Task 1), `toEmployeeResponse` (`employee.go:159`), `AddEmployeeCompanyParams` (existing).
- Produces:
  - `type AdminEmployeeHandler struct{ q *gen.Queries; pool *pgxpool.Pool }` + `func NewAdminEmployeeHandler(q *gen.Queries, pool *pgxpool.Pool) *AdminEmployeeHandler`
  - methods `List`, `ListCompanies`, `ReplaceCompanies` — all `func(c *gin.Context)`
  - `func normalizeCompanyIDs(ids []int64) []int64` — sorted, de-duplicated
  - `func validateMembershipReplace(companyIDs []int64, defaultCompanyID *int64) error`
- Produces (DTOs, `internal/http/dto/admin.go`):
  ```go
  type ReplaceEmployeeCompaniesRequest struct {
      CompanyIDs       []int64 `json:"company_ids" binding:"required,dive,min=1"`
      DefaultCompanyID *int64  `json:"default_company_id"`
  }
  type EmployeeCompaniesResponse struct {
      EmployeeID       int64   `json:"employee_id"`
      CompanyIDs       []int64 `json:"company_ids"`
      DefaultCompanyID *int64  `json:"default_company_id"`
  }
  ```
  (`dto.EmployeePage` already exists at `internal/http/dto/employee.go:114` and is reused for the list.)

**Routes:**
- `GET /api/v1/admin/employees` → paginated cross-company employee list
- `GET /api/v1/admin/employees/:id/companies` → current memberships
- `PUT /api/v1/admin/employees/:id/companies` → full replace

**Deliberate constraint:** `company_ids` is `binding:"required"`, so an empty array is a 422. Removing an employee's last membership would lock them out of login entirely (Task 1 refuses a memberless login); deactivating an employee is what `is_active: false` is for. Note this in the swagger description.

**Known limitation to record, not to solve here:** `is_admin` is a per-company role flag, so this namespace lets *any* company admin enumerate every tenant's employees. The spec asks for an admin gate and this backend has no platform-superadmin concept; `RequireAdminRole()` is the strictest gate that exists. Flag it in the task-list annotation.

- [ ] **Step 1: Write the failing test**

Create `internal/http/handler/admin_employee_test.go`:

```go
package handler

import (
	"errors"
	"reflect"
	"testing"

	"fleet/internal/platform/apierr"
)

func TestNormalizeCompanyIDs(t *testing.T) {
	tests := []struct {
		name string
		in   []int64
		want []int64
	}{
		{name: "sorts and de-duplicates", in: []int64{9, 3, 9, 1}, want: []int64{1, 3, 9}},
		{name: "already normal is unchanged", in: []int64{1, 2}, want: []int64{1, 2}},
		{name: "nil becomes an empty slice, never nil", in: nil, want: []int64{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCompanyIDs(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// The replace is a full overwrite of employee_companies, so a default that is
// not in the new set would leave the employee pointing at a company the same
// request just removed them from.
func TestValidateMembershipReplace(t *testing.T) {
	tests := []struct {
		name      string
		companies []int64
		defaultID *int64
		wantErr   bool
		wantField string
	}{
		{name: "default inside the set is allowed", companies: []int64{3, 7}, defaultID: ptr(7), wantErr: false},
		{name: "null default is allowed", companies: []int64{3, 7}, defaultID: nil, wantErr: false},
		{name: "default outside the set is rejected", companies: []int64{3, 7}, defaultID: ptr(8), wantErr: true, wantField: "default_company_id"},
		{name: "an empty set is rejected", companies: []int64{}, defaultID: nil, wantErr: true, wantField: "company_ids"},
		{name: "a non-positive company id is rejected", companies: []int64{0, 3}, defaultID: nil, wantErr: true, wantField: "company_ids"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMembershipReplace(tt.companies, tt.defaultID)
			if tt.wantErr == (err == nil) {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				return
			}
			var ae *apierr.Error
			if !errors.As(err, &ae) {
				t.Fatalf("error is %T, want *apierr.Error", err)
			}
			if ae.Status != 422 {
				t.Errorf("status = %d, want 422", ae.Status)
			}
			details, ok := ae.Details.(map[string]string)
			if !ok || details[tt.wantField] == "" {
				t.Errorf("details = %#v, want a %s entry", ae.Details, tt.wantField)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
go test ./internal/http/handler/ -run 'TestNormalizeCompanyIDs|TestValidateMembershipReplace' -v
```
Expected: compile failure — `undefined: normalizeCompanyIDs`, `undefined: validateMembershipReplace`.

- [ ] **Step 3: Add the SQL**

Append to `internal/db/queries/employee.sql`:

```sql
-- name: ListAllEmployees :many
-- Cross-company employee list for the admin namespace. Unlike ListEmployees
-- this is deliberately unscoped: it exists to answer "who exists anywhere",
-- which the company-scoped route cannot.
SELECT * FROM employee ORDER BY id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountAllEmployees :one
SELECT count(*) FROM employee;

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
```

Append to `internal/db/queries/company.sql`:

```sql
-- name: CountCompaniesByIDs :one
-- Pre-flight for a membership replace: a mismatch against the requested count
-- means at least one id names no company, which is a 422 rather than the 409 a
-- foreign-key violation would surface as.
SELECT count(*) FROM company WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: GetCompanyByID :one
-- Unscoped single-company read for the admin namespace.
SELECT * FROM company WHERE id = sqlc.arg(id);
```

Then:
```bash
sqlc generate
```

- [ ] **Step 4: Write the DTOs**

Create `internal/http/dto/admin.go`:

```go
package dto

// ReplaceEmployeeCompaniesRequest is a full overwrite of an employee's
// employee_companies rows. company_ids must be non-empty: an employee with no
// membership cannot log in at all, so deactivation belongs on is_active, not
// here. default_company_id must be null or one of company_ids.
type ReplaceEmployeeCompaniesRequest struct {
	CompanyIDs       []int64 `json:"company_ids" binding:"required,dive,min=1"`
	DefaultCompanyID *int64  `json:"default_company_id"`
}

// EmployeeCompaniesResponse is the employee's membership set after the write,
// normalized (sorted, de-duplicated).
type EmployeeCompaniesResponse struct {
	EmployeeID       int64   `json:"employee_id"`
	CompanyIDs       []int64 `json:"company_ids"`
	DefaultCompanyID *int64  `json:"default_company_id"`
}
```

- [ ] **Step 5: Write the handler**

Create `internal/http/handler/admin_employee.go`. Structure (follow `employee_actions.go` and `tire_actions.go` verbatim for the idioms):

```go
package handler

import (
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
)

// AdminEmployeeHandler serves the cross-company employee routes. Everything
// here is deliberately unscoped by the caller's own company_id — that is the
// gap it exists to close — so every route is gated on RequireAdminRole.
type AdminEmployeeHandler struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewAdminEmployeeHandler(q *gen.Queries, pool *pgxpool.Pool) *AdminEmployeeHandler {
	return &AdminEmployeeHandler{q: q, pool: pool}
}

// normalizeCompanyIDs returns the ids sorted and de-duplicated, never nil, so a
// replace is idempotent and the response is stable regardless of request order.
func normalizeCompanyIDs(ids []int64) []int64 {
	out := slices.Clone(ids)
	if out == nil {
		out = []int64{}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// validateMembershipReplace enforces the two invariants a full replace must
// hold: the employee keeps at least one company, and default_company_id names
// one of the companies the same request is granting.
func validateMembershipReplace(companyIDs []int64, defaultCompanyID *int64) error {
	if len(companyIDs) == 0 {
		return apierr.Validation(map[string]string{
			"company_ids": "at least one company is required; use is_active to deactivate an employee",
		})
	}
	for _, id := range companyIDs {
		if id < 1 {
			return apierr.Validation(map[string]string{"company_ids": "company ids must be positive"})
		}
	}
	if defaultCompanyID != nil && !slices.Contains(companyIDs, *defaultCompanyID) {
		return apierr.Validation(map[string]string{
			"default_company_id": "must be null or one of company_ids",
		})
	}
	return nil
}
```

Then the three gin methods:

**`List`** — `p := paginate.Parse(c)`; `h.q.ListAllEmployees(ctx, gen.ListAllEmployeesParams{Lim: int32(p.Limit), Off: int32(p.Offset)})`; `h.q.CountAllEmployees(ctx)`; map with the existing `toEmployeeResponse`; respond `c.JSON(http.StatusOK, paginate.NewPage(out, total, p))`.

**`ListCompanies`** — parse `:id` with the `strconv.ParseInt` + `id < 1` → `apierr.BadRequest("invalid id")` idiom; `h.q.GetEmployeeByID(ctx, id)` first so a missing employee is a 404 via `apierr.Abort` (`pgx.ErrNoRows` maps to 404 automatically); then `h.q.ListEmployeeCompanyIDs(ctx, id)`; respond with `dto.EmployeeCompaniesResponse{EmployeeID: id, CompanyIDs: normalizeCompanyIDs(ids), DefaultCompanyID: row.DefaultCompanyID}`.

**`ReplaceCompanies`** — in order:
1. parse `:id`
2. `bindJSONValidated(c, &req)`
3. `ids := normalizeCompanyIDs(req.CompanyIDs)`
4. `validateMembershipReplace(ids, req.DefaultCompanyID)` → abort on error
5. `h.q.GetEmployeeByID(ctx, id)` → 404 if absent
6. `n, err := h.q.CountCompaniesByIDs(ctx, ids)`; if `n != int64(len(ids))` → `apierr.Validation(map[string]string{"company_ids": "one or more companies do not exist"})`
7. transaction — `tx, err := h.pool.Begin(ctx)`, `defer tx.Rollback(ctx)`, `qtx := h.q.WithTx(tx)`:
   - `qtx.RemoveEmployeeCompaniesNotIn(ctx, gen.RemoveEmployeeCompaniesNotInParams{EmployeeID: id, CompanyIds: ids})`
   - for each id: `qtx.AddEmployeeCompany(ctx, gen.AddEmployeeCompanyParams{EmployeeID: id, CompanyID: cid})` (already `ON CONFLICT DO NOTHING`)
   - `qtx.SetEmployeeDefaultCompany(ctx, gen.SetEmployeeDefaultCompanyParams{ID: id, DefaultCompanyID: req.DefaultCompanyID, UpdatedAt: time.Now().UTC()})`
   - `tx.Commit(ctx)`
8. respond `c.JSON(http.StatusOK, dto.EmployeeCompaniesResponse{EmployeeID: id, CompanyIDs: ids, DefaultCompanyID: req.DefaultCompanyID})`

> Confirm the exact generated field name for the array param after `sqlc generate` (it will be `CompanyIds` or `CompanyIDs` depending on sqlc's initialism handling) and use whatever was generated.

- [ ] **Step 6: Register the routes**

In `internal/http/handler/router.go`, add a call next to the existing `registerCompanyRoutes(api, member, d)` (line 77):

```go
	registerAdminRoutes(member, d)
```

and add the function next to `registerCompanyRoutes` (after line 287):

```go
// registerAdminRoutes wires the cross-company namespace. Every route here reads
// or writes another tenant's data on purpose — the company-scoped routes cannot
// answer "who belongs to company X" or "did company X get seeded" — so the whole
// group sits behind RequireAdminRole rather than a module permission.
func registerAdminRoutes(member *gin.RouterGroup, d Deps) {
	admin := member.Group("", middleware.RequireAdminRole())

	employees := NewAdminEmployeeHandler(d.Queries, d.Pool)
	admin.GET("/admin/employees", employees.List)
	admin.GET("/admin/employees/:id/companies", employees.ListCompanies)
	admin.PUT("/admin/employees/:id/companies", employees.ReplaceCompanies)
}
```

- [ ] **Step 7: Add the routes to the router test**

In `internal/http/handler/router_test.go`, append to the `for _, want := range []string{...}` list:

```go
		"GET /api/v1/admin/employees",
		"GET /api/v1/admin/employees/:id/companies",
		"PUT /api/v1/admin/employees/:id/companies",
```

- [ ] **Step 8: Add the swagger stubs**

All three admin routes are real methods, not generic `crud.Handler` routes, so there are no `docs_*.go` stubs to write — put the annotation blocks **directly above the methods** in `admin_employee.go`, matching `employee_actions.go`. Use tag `admin`, `@Security BearerAuth`, and the standard failure ladder (`400, 401, 403, 404, 422`). Example for the replace:

```go
// ReplaceCompanies godoc
//
//	@Summary		Replace an employee's company memberships
//	@Description	Full overwrite of employee_companies. company_ids must be non-empty — an employee with no membership cannot log in, so use is_active to deactivate instead. default_company_id must be null or one of company_ids.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		int										true	"Employee id"
//	@Param			companies	body		dto.ReplaceEmployeeCompaniesRequest		true	"Membership set"
//	@Success		200			{object}	dto.EmployeeCompaniesResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Failure		422			{object}	dto.ErrorResponse
//	@Router			/api/v1/admin/employees/{id}/companies [put]
```

- [ ] **Step 9: Run the tests and the gate**

```bash
go test ./internal/http/handler/ -run 'TestNormalizeCompanyIDs|TestValidateMembershipReplace|TestRouter|TestAPIRoutes' -v
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
```
Expected: all PASS. If `NewRouter` panics on a route conflict, the new paths collide — check for an existing `/api/v1/:something` wildcard (there is none today).

- [ ] **Step 10: Regenerate the spec and commit**

```bash
swag init -g main.go -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage --parseDependency --parseInternal -o docs
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
git add -A
git commit -m "feat(admin): manage an employee's company memberships

AddEmployeeCompany had exactly one caller — EmployeeStore.Create, always with
the caller's own company — so nothing could add, remove or list memberships for
another company. Adds an admin-gated namespace: a cross-company employee list,
a membership read, and a transactional full-replace that validates
default_company_id against the set it is writing."
```

---

### Task 4: Read and set a company's owner (spec item 4)

**Files:**
- Modify: `internal/db/queries/company.sql`
- Modify: `internal/http/dto/admin.go`
- Create: `internal/http/handler/admin_company.go`
- Modify: `internal/http/handler/router.go`, `internal/http/handler/router_test.go`

**Interfaces:**
- Consumes: `GetCompanyByID` (Task 3).
- Produces:
  - `type AdminCompanyHandler struct{ q *gen.Queries; pool *pgxpool.Pool }` + `func NewAdminCompanyHandler(q *gen.Queries, pool *pgxpool.Pool) *AdminCompanyHandler`
  - methods `Owner`, `SetOwner`, and (Task 5) `Roles`, `WorkOrderStatuses`
- Produces (DTOs, appended to `internal/http/dto/admin.go`):
  ```go
  type CompanyOwnerResponse struct {
      EmployeeID int64  `json:"employee_id"`
      FirstName  string `json:"first_name"`
      LastName   string `json:"last_name"`
      Email      string `json:"email"`
      JobTitle   string `json:"job_title"`
  }
  // CompanyOwnerEnvelope wraps a nullable owner so "this company has no owner"
  // is a 200 with owner:null rather than a 404 the frontend has to special-case.
  type CompanyOwnerEnvelope struct {
      Owner *CompanyOwnerResponse `json:"owner"`
  }
  type SetCompanyOwnerRequest struct {
      EmployeeID int64 `json:"employee_id" binding:"required,min=1"`
  }
  ```

**Routes:**
- `GET /api/v1/admin/companies/:id/owner` → `200 {"owner": {...}}` or `200 {"owner": null}`
- `POST /api/v1/admin/companies/:id/set-owner` → `200 {"owner": {...}}`

**Note on the data model:** `employee.is_account_owner` is a single global boolean on the employee row, not a per-company flag. "Owner of company X" therefore means "a member of X whose `is_account_owner` is true". An employee who belongs to two companies and is flagged owner reads as owner of both. That is the pre-existing model; this task makes the *uniqueness within a company* guarantee real (clear-then-set in one transaction) without redesigning the column. Record this in the annotation.

- [ ] **Step 1: Write the failing test**

This task's logic is all SQL and transaction sequencing; the testable Go surface is the route wiring. Add to `internal/http/handler/router_test.go`'s expected list **first**, so the test fails before the routes exist:

```go
		"GET /api/v1/admin/companies/:id/owner",
		"POST /api/v1/admin/companies/:id/set-owner",
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
go test ./internal/http/handler/ -run TestRouterRegistersRoutes -v
```
Expected: FAIL with `route "GET /api/v1/admin/companies/:id/owner" not registered`.

- [ ] **Step 3: Add the SQL**

Append to `internal/db/queries/company.sql`:

```sql
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
WHERE id = sqlc.arg(id)
  AND EXISTS (
      SELECT 1 FROM employee_companies ec
      WHERE ec.employee_id = employee.id AND ec.company_id = sqlc.arg(company_id)
  )
RETURNING id, first_name, last_name, email, job_title;
```

```bash
sqlc generate
```

- [ ] **Step 4: Write the handler**

Create `internal/http/handler/admin_company.go` with `AdminCompanyHandler` and:

First add the shared id-parse helper that all four `AdminCompanyHandler` methods use (Task 5 adds the other two):

```go
func adminCompanyParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, apierr.BadRequest("invalid id")
	}
	return id, nil
}
```

**`Owner`** — `id, err := adminCompanyParam(c)`; `h.q.GetCompanyByID(ctx, id)` → 404 if absent; `h.q.GetCompanyOwner(ctx, id)`; on `pgx.ErrNoRows` respond `c.JSON(http.StatusOK, dto.CompanyOwnerEnvelope{})` (owner null), on any other error `apierr.Abort`; otherwise respond with the populated envelope.

```go
	owner, err := h.q.GetCompanyOwner(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		// A company with no owner is a real, expected state — not a 404 the
		// frontend has to distinguish from "no such company".
		c.JSON(http.StatusOK, dto.CompanyOwnerEnvelope{})
		return
	}
	if err != nil {
		apierr.Abort(c, err)
		return
	}
```

**`SetOwner`** — `id, err := adminCompanyParam(c)`; `bindJSONValidated(c, &req)`; `h.q.GetCompanyByID(ctx, id)` → 404; then a transaction:
```go
	now := time.Now().UTC()
	tx, err := h.pool.Begin(ctx)
	if err != nil { apierr.Abort(c, err); return }
	defer tx.Rollback(ctx)
	qtx := h.q.WithTx(tx)

	if err := qtx.ClearCompanyAccountOwner(ctx, gen.ClearCompanyAccountOwnerParams{CompanyID: id, UpdatedAt: now}); err != nil {
		apierr.Abort(c, err)
		return
	}
	owner, err := qtx.SetCompanyAccountOwner(ctx, gen.SetCompanyAccountOwnerParams{ID: req.EmployeeID, CompanyID: id, UpdatedAt: now})
	if errors.Is(err, pgx.ErrNoRows) {
		apierr.Abort(c, apierr.Validation(map[string]string{
			"employee_id": "must be an employee who belongs to this company",
		}))
		return
	}
	if err != nil { apierr.Abort(c, err); return }
	if err := tx.Commit(ctx); err != nil { apierr.Abort(c, err); return }
	c.JSON(http.StatusOK, dto.CompanyOwnerEnvelope{Owner: &dto.CompanyOwnerResponse{...}})
```

Annotation blocks inline, tag `admin`, standard failure ladder.

- [ ] **Step 5: Register the routes**

In `registerAdminRoutes`:
```go
	companies := NewAdminCompanyHandler(d.Queries, d.Pool)
	admin.GET("/admin/companies/:id/owner", companies.Owner)
	admin.POST("/admin/companies/:id/set-owner", companies.SetOwner)
```

- [ ] **Step 6: Run the tests and the gate**

```bash
go test ./internal/http/handler/ -run 'TestRouter|TestAPIRoutes' -v
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
```
Expected: PASS.

- [ ] **Step 7: Regenerate the spec and commit**

```bash
swag init -g main.go -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage --parseDependency --parseInternal -o docs
git add -A
git commit -m "feat(admin): read and set a company's account owner

No route returned an owner for a company and none set one, and plain CRUD on
is_account_owner let a caller flag two employees at once. Adds an admin-gated
read (200 with owner:null when there is none) and a set-owner action that
clears the current owner and sets the new one in one transaction, guarded on
employee_companies membership."
```

---

### Task 5: Verify another tenant's provisioning (spec item 5)

**Files:**
- Modify: `internal/http/handler/admin_company.go`
- Modify: `internal/http/handler/router.go`, `internal/http/handler/router_test.go`

**Interfaces:**
- Consumes: existing `ListRoles`/`CountRoles` (`internal/db/queries/role.sql`), `ListWorkOrderStatuses`/`CountWorkOrderStatuses` (`internal/db/queries/work_order_status.sql`), `toRoleResponse`, `toWorkOrderStatusResponse`, `dto.RolePage`, `dto.WorkOrderStatusPage`. **No new SQL.**
- Produces: `AdminCompanyHandler.Roles`, `AdminCompanyHandler.WorkOrderStatuses` — `func(c *gin.Context)`

**Routes:**
- `GET /api/v1/admin/companies/:id/roles`
- `GET /api/v1/admin/companies/:id/work-order-statuses`

**Design choice:** the spec offers either admin-scoped routes or a `company_id` override query param on the existing routes. Take the routes. An override param silently changes the meaning of an existing endpoint depending on who calls it — exactly the "admin's roles mislabeled as another company's" failure the spec is trying to prevent — whereas a distinct path makes the cross-tenant read explicit at the call site.

- [ ] **Step 1: Write the failing test**

Append to `internal/http/handler/router_test.go`'s expected list:
```go
		"GET /api/v1/admin/companies/:id/roles",
		"GET /api/v1/admin/companies/:id/work-order-statuses",
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
go test ./internal/http/handler/ -run TestRouterRegistersRoutes -v
```
Expected: FAIL — `route "GET /api/v1/admin/companies/:id/roles" not registered`.

- [ ] **Step 3: Write the implementation**

Add to `admin_company.go`:

```go
// Roles lists another company's roles so a provisioning check can answer "did
// company X get seeded" — the plain /roles route reads company_id off the
// caller's own JWT, which would show the admin's roles mislabeled as X's.
func (h *AdminCompanyHandler) Roles(c *gin.Context) {
	id, err := adminCompanyParam(c)
	if err != nil { apierr.Abort(c, err); return }
	ctx := c.Request.Context()
	if _, err := h.q.GetCompanyByID(ctx, id); err != nil { apierr.Abort(c, err); return }

	p := paginate.Parse(c)
	rows, err := h.q.ListRoles(ctx, gen.ListRolesParams{CompanyID: id, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil { apierr.Abort(c, err); return }
	total, err := h.q.CountRoles(ctx, id)
	if err != nil { apierr.Abort(c, err); return }

	out := make([]dto.RoleResponse, len(rows))
	for i, r := range rows { out[i] = toRoleResponse(r) }
	c.JSON(http.StatusOK, paginate.NewPage(out, total, p))
}
```

`WorkOrderStatuses` is the same shape over `ListWorkOrderStatuses`/`CountWorkOrderStatuses`/`toWorkOrderStatusResponse`/`dto.WorkOrderStatusResponse`.

`adminCompanyParam` already exists — Task 4 added it. Reuse it; do not redefine it.

> Verify the exact generated param field names (`Limit`/`Offset` vs `Lim`/`Off`) and the exact `toRoleResponse` / `toWorkOrderStatusResponse` signatures in `internal/http/handler/role.go` and `internal/http/handler/work_order_status.go` before writing this — sqlc names them from the SQL, and `role.sql` uses positional `$2`/`$3` so they will be `Limit`/`Offset`.

- [ ] **Step 4: Register the routes**

In `registerAdminRoutes`:
```go
	admin.GET("/admin/companies/:id/roles", companies.Roles)
	admin.GET("/admin/companies/:id/work-order-statuses", companies.WorkOrderStatuses)
```

- [ ] **Step 5: Add the swagger annotations**

Inline blocks, tag `admin`, `@Param id path int true "Company id"`, `@Param limit query int false "Page size"`, `@Param offset query int false "Offset"`, `@Success 200 {object} dto.RolePage` (resp. `dto.WorkOrderStatusPage`), failures `400, 401, 403, 404`.

- [ ] **Step 6: Run the tests and the gate**

```bash
go test ./internal/http/handler/ -run 'TestRouter|TestAPIRoutes' -v
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
```
Expected: PASS.

- [ ] **Step 7: Regenerate the spec and commit**

```bash
swag init -g main.go -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage --parseDependency --parseInternal -o docs
git add -A
git commit -m "feat(admin): read another company's roles and work-order statuses

/roles and /work-order-statuses scope off the caller's own JWT company_id, so
an admin could not verify that another tenant was seeded correctly — calling
them for company X silently returned the admin's own rows. Adds admin-gated
per-company reads that take the company from the path."
```

---

### Task 6: Denormalize `provider_name` onto `WarrantyResponse` (spec item 9)

**Files:**
- Modify: `internal/db/queries/warranty.sql`
- Modify: `internal/http/dto/warranty.go:27-37`
- Modify: `internal/http/handler/warranty.go`

**Interfaces:**
- Produces: `WarrantyResponse.ProviderName string \`json:"provider_name"\``
- Produces (sqlc): `gen.ListWarrantiesRow`, `gen.GetWarrantyRow` (both carry `ProviderName string`)

**Pattern to copy:** `internal/db/queries/asset_trailer_assignment.sql` — join the related table, alias the column `AS provider_name`, and re-read on Create/Update because `INSERT ... RETURNING` cannot return the joined name.

- [ ] **Step 1: Write the failing test**

The mapper is the testable seam. Add to a new `internal/http/handler/warranty_test.go`:

```go
package handler

import (
	"testing"
	"time"

	"fleet/internal/db/gen"
)

// VehicleDossier renders the provider as "#<id>" without this: the response
// carried provider_id but never the resolved name.
func TestToWarrantyResponseCarriesProviderName(t *testing.T) {
	now := time.Now().UTC()
	row := gen.ListWarrantiesRow{
		ID:           3,
		CompanyID:    1,
		ProviderID:   9,
		StartDate:    now,
		EndDate:      now,
		Terms:        "12 months",
		IsActive:     true,
		ProviderName: "Michelin",
	}

	got := toWarrantyResponse(row)
	if got.ProviderID != 9 {
		t.Errorf("provider_id = %d, want 9", got.ProviderID)
	}
	if got.ProviderName != "Michelin" {
		t.Errorf("provider_name = %q, want %q", got.ProviderName, "Michelin")
	}
}
```

> The **flat** shape above is what sqlc emits for `SELECT w.*, v.name AS provider_name` — verified against the existing `gen.ListAssetTrailerAssignmentsRow`, which flattens every column and appends `TrailerName string`. Run `sqlc generate` first and confirm the generated field set before finalizing the literal.

- [ ] **Step 2: Run the test to verify it fails**

```bash
go test ./internal/http/handler/ -run TestToWarrantyResponseCarriesProviderName -v
```
Expected: compile failure — `gen.ListWarrantiesRow` undefined, or `ProviderName` not a field of `dto.WarrantyResponse`.

- [ ] **Step 3: Change the SQL**

In `internal/db/queries/warranty.sql`, replace `GetWarranty` and `ListWarranties`:

```sql
-- The provider's name is joined in so a warranty row renders without a second
-- request per row just to resolve the vendor.

-- name: GetWarranty :one
SELECT w.*, v.name AS provider_name FROM warranty w
JOIN vendor v ON v.id = w.provider_id
WHERE w.id = sqlc.arg(id) AND w.company_id = sqlc.arg(company_id);

-- name: ListWarranties :many
SELECT w.*, v.name AS provider_name FROM warranty w
JOIN vendor v ON v.id = w.provider_id
WHERE w.company_id = sqlc.arg(company_id)
ORDER BY w.end_date DESC, w.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);
```

Leave `CreateWarranty`, `UpdateWarranty`, `CountWarranties` and `DeleteWarranty` unchanged.

```bash
sqlc generate
```

- [ ] **Step 4: Update the DTO**

In `internal/http/dto/warranty.go`, add to `WarrantyResponse` after `ProviderID`:

```go
	// Denormalized from the provider vendor so a row renders without a second
	// request per warranty.
	ProviderName string `json:"provider_name"`
```

- [ ] **Step 5: Update the handler**

- Change `toWarrantyResponse` to take `gen.ListWarrantiesRow` and set `ProviderName`.
- `GetWarranty` returns a *different* generated type (`gen.GetWarrantyRow`) with an identical field set. Convert rather than duplicating the mapper — the asset-trailer-assignment precedent copy-pastes the whole literal twice, which this plan does not follow:

```go
	// GetWarrantyRow and ListWarrantiesRow project the same columns in the same
	// order, so the conversion is exact and one mapper serves both.
	return toWarrantyResponse(gen.ListWarrantiesRow(r)), nil
```

  If that conversion does not compile (field order or type mismatch), fall back to a second small mapper rather than reordering the SQL.
- `List`/`Get`: unchanged call sites, but note the params struct is now `Lim`/`Off` (named args) instead of `Limit`/`Offset`. Fix accordingly.
- `Create`/`Update`: `INSERT ... RETURNING` cannot return the joined name, so re-read — mirroring `asset_trailer_assignment.go:80-83`:

```go
	// Re-read so the response carries the joined provider name, which the
	// INSERT itself cannot return.
	return s.Get(ctx, r.ID)
```

- [ ] **Step 6: Run the tests and the gate**

```bash
go test ./internal/http/handler/ -run TestToWarrantyResponse -v
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
```
Expected: PASS.

- [ ] **Step 7: Regenerate the spec and commit**

```bash
swag init -g main.go -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage --parseDependency --parseInternal -o docs
git add -A
git commit -m "feat(warranties): denormalize provider_name onto WarrantyResponse

The response carried provider_id but not the resolved name, so consumers had to
render '#<id>'. Joins vendor the same way asset trailer assignments join the
trailer's name, and re-reads on create/update since RETURNING cannot produce a
joined column."
```

---

### Task 7: Swagger annotation hygiene (spec items 13 and 14)

**Files:**
- Modify: `internal/http/handler/docs_link.go` (delete 12 lines)
- Modify: `internal/http/handler/docs.go:13-14`
- Modify: `internal/http/handler/docs_catalog.go:14-15, 80-81, 195-196, 260-261`
- Create: `internal/http/handler/openapi_spec_test.go`

**Interfaces:**
- Produces: `TestGeneratedSpecHasNoDuplicateParams`, `TestGeneratedSpecUsesLimitOffset` — regression guards that read `docs/swagger.json` from disk.

**Correction to the spec:** item 14 asks for the "standard `limit`/`offset`/`order` shape". The `order` param belongs only to the 14 filtered-list handlers in `list_filters.go`; these five routes are served by the generic `crud.Handler.List`, which calls `paginate.Parse` and has no ordering support at all. Documenting `order` on them would describe a parameter the server ignores. Use the plain `limit`/`offset` pair that the other 51 generic-CRUD collection stubs already use. Record this in the annotation.

- [ ] **Step 1: Write the failing test**

Create `internal/http/handler/openapi_spec_test.go`:

```go
package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// specPath is the checked-in swag output. These tests read the generated file
// rather than the annotations so a stale regeneration is caught too.
func loadSpec(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "swagger.json"))
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	var spec map[string]any
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	return spec
}

func eachOperation(t *testing.T, fn func(path, method string, params []any)) {
	t.Helper()
	paths, ok := loadSpec(t)["paths"].(map[string]any)
	if !ok {
		t.Fatal("spec has no paths object")
	}
	for path, item := range paths {
		methods, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for method, op := range methods {
			body, ok := op.(map[string]any)
			if !ok {
				continue
			}
			params, _ := body["parameters"].([]any)
			fn(path, method, params)
		}
	}
}

// OpenAPI requires parameters to be unique by (name, in). Orval rejects the
// whole spec on a violation, not just the offending operation.
func TestGeneratedSpecHasNoDuplicateParams(t *testing.T) {
	eachOperation(t, func(path, method string, params []any) {
		seen := make(map[string]bool, len(params))
		for _, raw := range params {
			p, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			key, _ := p["name"].(string)
			in, _ := p["in"].(string)
			key += " in:" + in
			if seen[key] {
				t.Errorf("%s %s declares %s twice", method, path, key)
			}
			seen[key] = true
		}
	})
}

// Every collection route documents the limit/offset convention; page/page_size
// still works at runtime but the generated param types must not offer only it.
func TestGeneratedSpecUsesLimitOffset(t *testing.T) {
	eachOperation(t, func(path, method string, params []any) {
		for _, raw := range params {
			p, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if name, _ := p["name"].(string); name == "page_size" || name == "page" {
				t.Errorf("%s %s documents %q; use limit/offset", method, path, name)
			}
		}
	})
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
go test ./internal/http/handler/ -run TestGeneratedSpec -v
```
Expected: FAIL — 6 duplicate-parameter errors and 10 `page`/`page_size` errors.

- [ ] **Step 3: Delete the duplicated `@Param` lines**

In `internal/http/handler/docs_link.go`, delete the **second** `limit`/`offset` pair in each of the six stubs. Delete lines in descending order so earlier deletions do not shift later line numbers: **232-233, 189-190, 146-147, 103-104, 60-61, 17-18**.

Verify afterwards — every stub should have exactly one pair:
```bash
grep -c "@Param.*limit" internal/http/handler/docs_link.go   # expect 6
grep -c "@Param.*offset" internal/http/handler/docs_link.go  # expect 6
```

- [ ] **Step 4: Convert the five `page`/`page_size` blocks**

In `internal/http/handler/docs.go` replace lines 13-14, and in `internal/http/handler/docs_catalog.go` replace lines 14-15, 80-81, 195-196, 260-261 — each pair becomes:

```go
//	@Param		limit		query		int	false	"Page size"
//	@Param		offset		query		int	false	"Offset"
```

Then let gofmt re-align the comment columns:
```bash
gofmt -w internal/http/handler/docs.go internal/http/handler/docs_catalog.go internal/http/handler/docs_link.go
```

Verify nothing is left:
```bash
grep -rn "page_size" internal/http/handler/   # expect no matches
```

- [ ] **Step 5: Regenerate the spec**

```bash
swag init -g main.go -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage --parseDependency --parseInternal -o docs
```

- [ ] **Step 6: Run the tests and the gate**

```bash
go test ./internal/http/handler/ -run TestGeneratedSpec -v
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
```
Expected: both spec tests PASS.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "docs(swagger): drop duplicated params and standardize pagination

Six m2m list stubs repeated their @Param limit/offset pair, so swag emitted
[limit, offset, limit, offset] and the spec failed OpenAPI's uniqueItems
constraint — Orval rejected the whole document over it. Five collection routes
still documented page/page_size while the other 61 document limit/offset; both
spellings work at runtime, but the generated param types only offered the old
one. Adds a test over the generated spec so neither can regress."
```

---

### Task 8: CORS allowlist for real browser clients (spec item 2)

**Problem:** the deployed `CORS_ORIGINS` contains only the backend's own origin, which no browser ever sends, so no `Access-Control-Allow-Origin` header comes back. This is a deployment-value change, not a code change — the middleware already echoes any listed origin.

**What can and cannot be fixed from this repo:** the *live* value lives in `/opt/fleet/.env` on the VPS, which is gitignored and never written by a deploy — that still needs a human with server access. But the committed template the deployment is provisioned from **carries the defect**: `deploy/.env.production.example:29` reads

```sh
CORS_ORIGINS=https://go-logistics.jesuslab135.com
```

— the backend's own origin, which is precisely the value no browser ever sends. That is an in-repo fix, and it is the source the next deployment copies from.

**Files:**
- Create: `internal/http/middleware/cors_test.go`
- Modify: `deploy/.env.production.example` (line ~28-29 — the real target)
- Modify: `.env.example` (QA block ~line 54, PROD block ~line 81)
- Modify: `docs/2026-08-14_golanglogistics-deployment.md` (the doc-embedded copy of the same template, ~line 274)

- [ ] **Step 1: Write the failing test**

Create `internal/http/middleware/cors_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func corsResponse(t *testing.T, origins []string, method, origin string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS(origins))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(method, "/x", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if method == http.MethodOptions {
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// The deployed allowlist held only the backend's own origin, which no browser
// ever sends, so every client got a response with no Allow-Origin header at all.
func TestCORSAllowOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
		method  string
		origin  string
		want    string
	}{
		{
			name:    "an allowlisted origin is echoed",
			origins: []string{"https://app.example.com", "http://localhost:5173"},
			method:  http.MethodGet,
			origin:  "http://localhost:5173",
			want:    "http://localhost:5173",
		},
		{
			name:    "a preflight from an allowlisted origin is echoed",
			origins: []string{"http://localhost:5173"},
			method:  http.MethodOptions,
			origin:  "http://localhost:5173",
			want:    "http://localhost:5173",
		},
		{
			name:    "an origin outside the allowlist gets no header",
			origins: []string{"https://app.example.com"},
			method:  http.MethodGet,
			origin:  "https://evil.example.com",
			want:    "",
		},
		{
			name:    "a wildcard allowlist allows any origin",
			origins: []string{"*"},
			method:  http.MethodGet,
			origin:  "https://anything.example.com",
			want:    "*",
		},
		{
			name:    "an allowlist holding only the backend's own origin serves no browser",
			origins: []string{"https://go-logistics.jesuslab135.com"},
			method:  http.MethodOptions,
			origin:  "http://localhost:5173",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := corsResponse(t, tt.origins, tt.method, tt.origin)
			if got := w.Header().Get("Access-Control-Allow-Origin"); got != tt.want {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCORSPreflightIsNoContent(t *testing.T) {
	w := corsResponse(t, []string{"http://localhost:5173"}, http.MethodOptions, "http://localhost:5173")
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("preflight returned no Access-Control-Allow-Methods")
	}
}
```

- [ ] **Step 2: Run the test**

```bash
go test ./internal/http/middleware/ -run TestCORS -v
```
Expected: **PASS immediately.** This is the one place in this plan where the test does not start red — the middleware is already correct and the defect is a deployed value. The test is a characterization guard proving the middleware behaves as the fix assumes, and pinning the exact failure mode (the last case) that is live today. State this explicitly rather than pretending it was red.

- [ ] **Step 3: Fix the deployment template that carries the defect**

In `deploy/.env.production.example`, replace the comment and value at lines ~28-29:

```sh
# The frontend's exact origin(s), comma-separated. NOT "*" in production, and
# NOT this backend's own origin — a browser never sends that, so allowlisting
# only it means no Access-Control-Allow-Origin header comes back and every
# browser client is blocked. Add http://localhost:5173 while browser-testing
# against this deployment.
CORS_ORIGINS=https://app.example.com,http://localhost:5173
```

Keep the `example.com` placeholder convention the surrounding file already uses — line 22 of the deployment doc names `example.com` as a value the executor substitutes.

- [ ] **Step 3b: Make the local example config concrete**

In `.env.example`, replace the QA block line 54 and PROD block line 81 comments with a value that names both the app origin and the local dev origin, and explain why:

```sh
# Comma-separated browser origins. This must list every origin a browser client
# is served from — the backend's own origin is never sent by a browser and
# allowlisting only it means no Access-Control-Allow-Origin comes back at all.
# Keep http://localhost:5173 while browser-testing against this deployment.
CORS_ORIGINS=https://app.your-domain.com,http://localhost:5173
```

- [ ] **Step 4: Fix the deployment plan template**

In `docs/2026-08-14_golanglogistics-deployment.md`, update line ~274 in the `deploy/.env.production.example` block to the same shape, keeping the `example.com` placeholder convention the doc already uses:

```sh
# The frontend's exact origin(s), comma-separated — NOT "*" in production, and
# NOT the backend's own origin, which no browser sends. Add http://localhost:5173
# while browser-testing against this deployment.
CORS_ORIGINS=https://app.example.com,http://localhost:5173
```

- [ ] **Step 5: Run the gate and commit**

```bash
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
git add -A
git commit -m "fix(cors): allowlist the origins a browser client actually sends

deploy/.env.production.example set CORS_ORIGINS to the backend's own origin,
which no browser ever sends, so the middleware echoed no
Access-Control-Allow-Origin and every browser client was blocked. The
middleware is correct; the provisioned value was not. Fixes the deployment
template, makes the .env examples concrete, and adds the middleware's first
test coverage, pinning the exact failure mode."
```

- [ ] **Step 6: Report the manual step**

The server value still has to be set by hand. Surface this in the final report and in the task-list annotation:
```sh
# on the VPS, in /opt/fleet/.env
CORS_ORIGINS=https://<frontend-origin>,http://localhost:5173
# then
docker compose up -d api
```
Re-verify with the same curl the spec used, expecting an `Access-Control-Allow-Origin` line this time.

---

### Task 9: Annotate the task list (the user's explicit ask)

**Files:**
- Modify: `C:\Users\jesus.olmos\Downloads\backend-engineer-task-todo.md`

For each of items 1, 2, 3, 4, 5, 9, 13, 14, insert immediately after the item's existing text — **without altering the original text** — an extremely concise pair:

```markdown
> **Fixed:** <one sentence: what changed, naming the file or route.>
> **Frontend:** <one or two sentences: the exact route/shape/field to consume, and any behavior change to handle.>
```

For items 6, 7, 8, 10, 11, 12, insert instead:

```markdown
> **Not done:** blocked on <the specific product decision>. No code changed.
```

Rules: English, no restating the problem, no hedging, name real routes/fields. Where a concern was flagged during implementation (the `is_admin` cross-tenant reach on the admin namespace, the non-empty `company_ids` constraint, the `order` param correction, the manual server-side CORS step), state it in one clause inside the `Frontend:` or `Fixed:` line rather than adding a paragraph.

- [ ] **Step 1: Write the annotations** — one edit per item, in file order.
- [ ] **Step 2: Verify** — `grep -c "^> \*\*Fixed:\*\*" "C:/Users/jesus.olmos/Downloads/backend-engineer-task-todo.md"` should return 8; `grep -c "^> \*\*Not done:\*\*"` should return 6.
- [ ] **Step 3: No commit** — the file lives outside the repo.

---

## Final verification (after all tasks)

```bash
cd "C:/Users/jesus.olmos/Desktop/PROJECTS/GoLangLogistics/GoLangLogisticsFinal"
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./... && echo LOCAL-GATE-OK
sqlc generate && git diff --exit-code internal/db/gen && echo SQLC-CLEAN
swag init -g main.go -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage --parseDependency --parseInternal -o docs && git diff --exit-code docs && echo SWAG-CLEAN
git log --oneline main..HEAD
```

All three gates must print their OK line. The two drift gates mirror the CI jobs the deployment plan specifies.
