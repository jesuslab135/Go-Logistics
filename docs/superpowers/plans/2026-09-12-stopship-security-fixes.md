# Stop-Ship Security Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the seven stop-ship findings from the 2026-09-12 code review — six cross-tenant / data-integrity defects and the absence of rate limiting.

**Architecture:** Each fix is applied at the layer that owns the invariant. Tenant checks go into the query or the resolver that all callers share, not into individual handlers. Privilege-bearing fields are removed from request DTOs rather than validated in place. The rate limiter is a new transport middleware with no dependencies outside the standard library.

**Tech Stack:** Go 1.26.5, Gin, pgx/v5, sqlc v1.29.0, golang-migrate, Postgres 16, nginx.

**Spec:** `../../../../code-review-2026-09-12.md` (findings CRIT-3 … CRIT-8, HIGH-3, HIGH-8). That file sits beside the repo, not inside it.

## Global Constraints

- **Build/test only outside OneDrive.** The repo does not compile at `C:\Users\lider\OneDrive\...` — `go build` reports `cannot find package` for `golang.org/x/crypto`, `golang.org/x/image`, `golang-jwt/jwt/v5` and `pgx-shopspring-decimal`, all of which resolve correctly via `go list`. The same tree builds clean elsewhere. **Work in a git worktree at a non-OneDrive path** (see Task 0).
- **Baseline before any change:** `go test ./...` → 14 packages ok, 0 failures.
- **sqlc pin:** `v1.29.0`, matching `.github/workflows/ci.yml:34`. Already on PATH. Any `.sql` query change requires `sqlc generate` and a commit of the regenerated `internal/db/gen`.
- **Generated-code drift is CI-enforced** via `git diff --exit-code internal/db/gen`. On this CRLF checkout, `diff(1)` and `git status` misreport drift — trust `git diff --exit-code` only.
- **Tests are `go test ./...` by default.** Integration tests are behind `//go:build integration` and need Postgres: `docker compose up -d db`, then `go test -tags integration ./internal/http/handler/`. Both are currently running on this machine.
- **Error helpers:** use `apierr.NotFound`, `apierr.Conflict`, `apierr.Validation(map[string]string{...})` — never construct `*apierr.Error` by hand.
- **Tenant accessor:** `middleware.CompanyFromContext(ctx) int64`. Never read the company from a request body.
- **Commit style:** conventional commits, matching existing history (`fix(bulk): ...`, `feat(api): ...`). One commit per task.

---

## File Structure

**Created:**
- `internal/http/middleware/ratelimit.go` — token-bucket limiter, one responsibility: decide allow/deny for a key.
- `internal/http/middleware/ratelimit_test.go`
- `internal/db/migrations/000024_inventory_non_negative.up.sql` / `.down.sql`
- `internal/http/handler/tenant_fileref_test.go`
- `internal/http/handler/crosstenant_integration_test.go` (build tag `integration`)

**Modified:**
- `cmd/api/main.go:83-98` — shutdown drain
- `internal/http/middleware/rbac.go:254-272` — `DecodePermissions` null handling
- `internal/http/middleware/rbac_test.go` — null/`{}` table test
- `internal/db/queries/{work_order,purchase_order,service_entry}_line_item.sql:27`, `work_order_sub_line_item.sql:27` — `:exec` → `:execrows`
- `internal/db/queries/money.sql` — tenant predicates on the recompute family
- `internal/http/handler/money.go` — `recalc*` signatures take `companyID`
- 21 `recalc*` call sites (enumerated in Task 6)
- `internal/http/handler/{work_order,purchase_order,service_entry}_line_item.go`, `work_order_sub_line_item.go` — 404 on zero rows
- `internal/http/handler/inventory_journal_entry.go` — negative-stock guard
- `internal/http/dto/tire_assignment_request.go` — drop state fields from update
- `internal/http/handler/tire_assignment_request.go` — stop forwarding them
- `internal/http/handler/router.go` — tire module gate, rate-limit wiring
- `internal/http/handler/fileref.go` — tenant-validated key resolution
- `internal/http/handler/storage_cleanup.go` — pass the tenant through
- `deploy/nginx/go-logistics.conf` — `limit_req_zone` + `limit_req`

---

## Task 0: Isolated worktree and baseline

**Files:** none modified.

**Interfaces:**
- Produces: `$WT` — the worktree path every later task works in.

- [ ] **Step 1: Create the worktree on a fix branch**

```bash
WT="$TMPDIR/golog-stopship"   # any non-OneDrive path
cd "/c/Users/lider/OneDrive/Desktop/PROJECTS/GoLogistics/Go-Logistics"
git worktree add "$WT" -b fix/stopship-findings main
cd "$WT"
```

- [ ] **Step 2: Confirm the baseline builds and tests green**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: build and vet silent; 14 packages `ok`, zero `FAIL`.

If this does not pass, stop — every later "verify it fails" step depends on a green baseline.

- [ ] **Step 3: Confirm the integration harness runs**

Run: `docker compose -f "<repo>/docker-compose.yml" ps` then
`go test -tags integration ./internal/http/handler/ -run TestGrantingARoleFromAnotherCompanyIsRefused`
Expected: `ok`. This proves throwaway-database creation works before Task 5 relies on it.

---

## Task 1: Shutdown drains before exit (CRIT-8)

**Files:**
- Modify: `cmd/api/main.go:83-98`

**Interfaces:**
- Consumes: nothing.
- Produces: nothing (behavioural fix inside `run`).

`srv.Shutdown` runs in a goroutine nobody waits on. `Shutdown` closes listeners first, so `ListenAndServe` returns `http.ErrServerClosed` immediately, `run` returns, the deferred `pool.Close()` fires, and the process exits while requests are still in flight. The 15-second budget never elapses.

There is no unit test for this — it is a process-lifetime property, and a test would have to bind a port and race the runtime. The change is small and reviewable by inspection; correctness is argued from the `net/http` contract that `Shutdown` returns only once all connections are idle.

- [ ] **Step 1: Apply the fix**

Replace lines 83-98 of `cmd/api/main.go`:

```go
	// Shutdown returns only once in-flight requests have finished, so run()
	// must wait for it: ListenAndServe returns as soon as Shutdown *starts*,
	// and returning there would fire the deferred pool.Close() underneath
	// requests that are still running.
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	logger.Info("http server listening", "addr", cfg.HTTPAddr, "env", cfg.Env)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-shutdownDone
	return nil
}
```

- [ ] **Step 2: Verify it compiles and nothing regressed**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: 14 packages `ok`.

- [ ] **Step 3: Commit**

```bash
git add cmd/api/main.go
git commit -m "fix(api): wait for graceful shutdown to drain before exiting"
```

---

## Task 2: A JSON null denies a module (HIGH-3)

**Files:**
- Modify: `internal/http/middleware/rbac.go:254-272`
- Test: `internal/http/middleware/rbac_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: `DecodePermissions(raw []byte) map[string]ModulePermissions` — unchanged signature.

`DecodePermissions` treats an empty action map as "grant the whole module" (Django's convention, deliberate for `{}`). `json.Unmarshal` of `null` into a `map[string]bool` succeeds and yields a **nil** map, which is also `len() == 0` — so `null` grants everything. Verified:

```
{"assets":{}}     -> All=true   (intended)
{"assets":null}   -> All=true   (NOT intended)
```

- [ ] **Step 1: Write the failing test**

Append to `internal/http/middleware/rbac_test.go`:

```go
// An explicit null is a client saying "nothing selected for this module". It
// must not take the {} path, which grants the module in full — that inverts
// the caller's intent, and a permissions form that serialises an empty
// selection as null would silently hand out everything.
func TestDecodePermissionsNullDeniesTheModule(t *testing.T) {
	tests := []struct {
		name     string
		doc      string
		wantAll  bool
		wantRead bool
	}{
		{"empty object grants the module", `{"assets":{}}`, true, true},
		{"null denies the module", `{"assets":null}`, false, false},
		{"explicit action", `{"assets":{"read":true}}`, false, true},
		{"explicit false", `{"assets":{"read":false}}`, false, false},
		{"malformed value denies", `{"assets":[1,2]}`, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := DecodePermissions([]byte(tt.doc))
			if got := perms["assets"].All; got != tt.wantAll {
				t.Errorf("All = %v, want %v", got, tt.wantAll)
			}
			id := Identity{IsActive: true, HasRole: true, Permissions: perms}
			if got := id.CanAction("assets", "read"); got != tt.wantRead {
				t.Errorf("CanAction(assets, read) = %v, want %v", got, tt.wantRead)
			}
		})
	}
}
```

- [ ] **Step 2: Run it to confirm it fails**

Run: `go test ./internal/http/middleware/ -run TestDecodePermissionsNullDeniesTheModule -v`
Expected: FAIL on the "null denies the module" subtest — `All = true, want false`.

- [ ] **Step 3: Apply the fix**

In `internal/http/middleware/rbac.go`, add `"bytes"` to the imports and change the decode loop:

```go
	for module, body := range modules {
		// An explicit null is not an empty object. Only {} means "the whole
		// module"; null means the client named the module and selected
		// nothing in it, which denies.
		if bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
			continue
		}
		var actions map[string]bool
		if err := json.Unmarshal(body, &actions); err != nil {
			continue
		}
		out[module] = ModulePermissions{All: len(actions) == 0, Actions: actions}
	}
```

- [ ] **Step 4: Run the test to confirm it passes**

Run: `go test ./internal/http/middleware/ -v`
Expected: all PASS, including the pre-existing rbac tests.

- [ ] **Step 5: Commit**

```bash
git add internal/http/middleware/rbac.go internal/http/middleware/rbac_test.go
git commit -m "fix(rbac): a null permissions value denies a module instead of granting it"
```

---

## Task 3: Stock cannot go negative (CRIT-7)

**Files:**
- Modify: `internal/http/handler/inventory_journal_entry.go` (`applyAdjustment`)
- Create: `internal/db/migrations/000024_inventory_non_negative.up.sql`
- Create: `internal/db/migrations/000024_inventory_non_negative.down.sql`

**Interfaces:**
- Consumes: `adjustment` struct (existing, unexported), `gen.LockPartInventoryRow.AvailableQuantity decimal.Decimal`.
- Produces: no signature change to `applyAdjustment`.

`applyAdjustment` locks the row correctly (`LockPartInventory … FOR UPDATE`, inside a transaction) but never checks the result. `ApplyPartInventoryAdjustment` is `available_quantity = available_quantity + $1` with no floor, and the schema has **zero CHECK constraints anywhere**. `adjustment_quantity: -1000` against a stock of 5 commits `-995`.

Two layers: refuse in the handler (gives a useful 422), and a CHECK constraint so no other path can do it.

- [ ] **Step 1: Write the failing test**

Create `internal/http/handler/inventory_guard_test.go`:

```go
package handler

import (
	"testing"

	"github.com/shopspring/decimal"
)

// A negative adjustment may not take stock below zero. The lock is already
// held when this is checked, so the decision is made against the same row the
// write will touch.
func TestAdjustmentWithinStockIsAllowed(t *testing.T) {
	for _, tt := range []struct {
		name       string
		stock      string
		adjustment string
		wantErr    bool
	}{
		{"decrease within stock", "5", "-3", false},
		{"decrease to exactly zero", "5", "-5", false},
		{"decrease past zero", "5", "-6", true},
		{"decrease far past zero", "5", "-1000", true},
		{"any increase", "0", "1000", false},
		{"zero adjustment", "0", "0", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stock := decimal.RequireFromString(tt.stock)
			adj := decimal.RequireFromString(tt.adjustment)
			err := checkStockFloor(stock, adj)
			if tt.wantErr && err == nil {
				t.Errorf("stock %s adjusted by %s: want refusal, got nil", tt.stock, tt.adjustment)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("stock %s adjusted by %s: want nil, got %v", tt.stock, tt.adjustment, err)
			}
		})
	}
}
```

- [ ] **Step 2: Run it to confirm it fails**

Run: `go test ./internal/http/handler/ -run TestAdjustmentWithinStockIsAllowed`
Expected: FAIL — `undefined: checkStockFloor`.

- [ ] **Step 3: Add the guard and call it**

In `internal/http/handler/inventory_journal_entry.go`, add:

```go
// checkStockFloor refuses an adjustment that would take available stock below
// zero. Inventory that has gone negative is not a smaller number — it is a
// valuation, a reorder point and a low-stock dashboard all reporting on a
// quantity that cannot exist.
func checkStockFloor(available, adjustment decimal.Decimal) error {
	if available.Add(adjustment).Sign() < 0 {
		return apierr.Validation(map[string]string{
			"adjustment_quantity": "would take available stock below zero",
		})
	}
	return nil
}
```

and call it in `applyAdjustment`, immediately after the `stock.PartID != a.partID` check and before `ApplyPartInventoryAdjustment`:

```go
	if err := checkStockFloor(stock.AvailableQuantity, a.quantity); err != nil {
		return gen.InventoryJournalEntry{}, err
	}
```

Ensure `github.com/shopspring/decimal` is imported in that file.

- [ ] **Step 4: Run the test to confirm it passes**

Run: `go test ./internal/http/handler/ -run TestAdjustmentWithinStockIsAllowed -v`
Expected: all subtests PASS.

- [ ] **Step 5: Add the database backstop**

`internal/db/migrations/000024_inventory_non_negative.up.sql`:

```sql
-- 000024_inventory_non_negative.up.sql
-- applyAdjustment locks the stock row and now refuses an adjustment that would
-- take it below zero, but that is one code path's discipline. This is the
-- invariant itself: available stock is a physical count, and no write of any
-- kind may leave it negative.
--
-- Existing negative rows would make this constraint un-addable, so they are
-- floored to zero first. That is a data correction, not a silent one — a row
-- that reached a negative count was already reporting a quantity that could
-- not exist, and there is no record of what the true count was.
UPDATE part_inventory SET available_quantity = 0 WHERE available_quantity < 0;

ALTER TABLE part_inventory
    ADD CONSTRAINT ck_part_inventory_available_non_negative
    CHECK (available_quantity >= 0);
```

`internal/db/migrations/000024_inventory_non_negative.down.sql`:

```sql
-- 000024_inventory_non_negative.down.sql
-- The floored rows are not restored: their pre-migration values were not
-- recorded, and re-deriving them from the journal is not something a down
-- migration should attempt.
ALTER TABLE part_inventory
    DROP CONSTRAINT IF EXISTS ck_part_inventory_available_non_negative;
```

- [ ] **Step 6: Confirm sqlc still generates identically**

The migration adds no column, so `internal/db/gen` must not change.

Run: `sqlc generate && git diff --exit-code internal/db/gen`
Expected: exit 0, no output.

- [ ] **Step 7: Run the full suite**

Run: `go test ./...`
Expected: 14 packages `ok`.

- [ ] **Step 8: Commit**

```bash
git add internal/http/handler/inventory_journal_entry.go \
        internal/http/handler/inventory_guard_test.go \
        internal/db/migrations/000024_inventory_non_negative.up.sql \
        internal/db/migrations/000024_inventory_non_negative.down.sql
git commit -m "fix(inventory): refuse adjustments that drive available stock negative"
```

---

## Task 4: Tire approvals cannot be forged through the generic PUT (CRIT-6)

**Files:**
- Modify: `internal/http/dto/tire_assignment_request.go:19-30`
- Modify: `internal/http/handler/tire_assignment_request.go` (`Update`)
- Modify: `internal/http/handler/router.go:203-204`
- Test: `internal/http/handler/validation_test.go` (append)

**Interfaces:**
- Consumes: `gen.UpdateTireAssignmentRequestParams` (unchanged — the handler supplies the existing row's values for the fields the DTO no longer carries).
- Produces: `dto.UpdateTireAssignmentRequestRequest` without `State`, `ApprovedByID`, `ResolvedAt`, `RejectionReason`.

`POST /tire-assignment-requests/:id/approve` is gated on `RequireAction("tire_approvals","approve")`. The CRUD handler for the same resource is registered on the bare `member` group with **no module gate**, and the update DTO carries every state field — so any company member can `PUT` an approval into existence and attribute it to someone else.

- [ ] **Step 1: Write the failing test**

Append to `internal/http/handler/validation_test.go`:

```go
// State transitions belong to the approve action, which is gated on
// tire_approvals/approve. If the ordinary update DTO carries them, any member
// can forge an approval with a PUT and attribute it to another employee.
func TestTireAssignmentUpdateCannotCarryApprovalFields(t *testing.T) {
	body := `{"tire_id":1,"vehicle_id":2,"state":"APPROVED","approved_by_id":42,` +
		`"resolved_at":"2026-09-12T10:00:00Z","rejection_reason":"none"}`

	var in dto.UpdateTireAssignmentRequestRequest
	if err := json.Unmarshal([]byte(body), &in); err != nil {
		t.Fatal(err)
	}

	// Reflection rather than field access: the point is that the fields are
	// gone from the type, which would not compile if asserted directly.
	forbidden := []string{"State", "ApprovedByID", "ResolvedAt", "RejectionReason"}
	typ := reflect.TypeOf(in)
	for _, name := range forbidden {
		if _, ok := typ.FieldByName(name); ok {
			t.Errorf("UpdateTireAssignmentRequestRequest still carries %s; "+
				"state belongs to the approve action, not an ordinary update", name)
		}
	}
}
```

Add `"reflect"` to that file's imports if absent.

- [ ] **Step 2: Run it to confirm it fails**

Run: `go test ./internal/http/handler/ -run TestTireAssignmentUpdateCannotCarryApprovalFields -v`
Expected: FAIL — four errors, one per field.

- [ ] **Step 3: Remove the fields from the DTO**

In `internal/http/dto/tire_assignment_request.go`, replace the update struct:

```go
// UpdateTireAssignmentRequestRequest is an ordinary edit of a pending request.
// State, approver and resolution are deliberately absent: those are set by
// POST /tire-assignment-requests/{id}/approve, which is gated on
// tire_approvals/approve. A field carried here would be a way around that gate.
type UpdateTireAssignmentRequestRequest struct {
	TireID        int64     `json:"tire_id"`
	VehicleID     int64     `json:"vehicle_id"`
	PositionCode  string    `json:"position_code" binding:"omitempty,max=10"`
	RequestedByID *int64    `json:"requested_by_id"`
	RequestedAt   time.Time `json:"requested_at"`
	Notes         string    `json:"notes"`
}
```

- [ ] **Step 4: Preserve the existing state in the handler**

In `internal/http/handler/tire_assignment_request.go`, `Update` must read the current row and carry its state forward unchanged:

```go
func (s *TireAssignmentRequestStore) Update(ctx context.Context, id int64, in dto.UpdateTireAssignmentRequestRequest) (dto.TireAssignmentRequestResponse, error) {
	company := middleware.CompanyFromContext(ctx)

	// State, approver and resolution are not editable here — they are read
	// back from the stored row so an ordinary edit cannot move the request
	// through its lifecycle. Approving is POST .../approve.
	current, err := s.q.GetTireAssignmentRequest(ctx, gen.GetTireAssignmentRequestParams{
		ID: id, CompanyID: company,
	})
	if err != nil {
		return dto.TireAssignmentRequestResponse{}, err
	}

	r, err := s.q.UpdateTireAssignmentRequest(ctx, gen.UpdateTireAssignmentRequestParams{
		ID:              id,
		CompanyID:       company,
		TireID:          in.TireID,
		VehicleID:       in.VehicleID,
		PositionCode:    in.PositionCode,
		State:           current.State,
		RequestedByID:   in.RequestedByID,
		RequestedAt:     in.RequestedAt,
		ApprovedByID:    current.ApprovedByID,
		ResolvedAt:      current.ResolvedAt,
		RejectionReason: current.RejectionReason,
		Notes:           in.Notes,
	})
	if err != nil {
		return dto.TireAssignmentRequestResponse{}, err
	}
	return toTireAssignmentRequestResponse(r), nil
}
```

Confirm the getter's real name and params before writing this — run
`grep -n "GetTireAssignmentRequest" internal/db/gen/*.go` and match the generated signature exactly.

- [ ] **Step 5: Put the CRUD group behind the tires module gate**

In `internal/http/handler/router.go`, change the registration at line 203 from `member` to `tires`:

```go
	// Django's TireAssignmentRequestViewSet required membership only. It is
	// gated on the tires module here: a request to move a tire is a tire
	// write, and leaving it on the bare member group also left its state
	// fields reachable by anyone who could reach the router.
	registerCrudWithList(tires, "/tire-assignment-requests",
		crud.NewHandler[dto.TireAssignmentRequestResponse, dto.CreateTireAssignmentRequestRequest, dto.UpdateTireAssignmentRequestRequest](NewTireAssignmentRequestStore(d.Queries, d.Pool)), lists.TireAssignmentRequests)
```

Leave the `member.POST(".../approve", ...)` line as it is — it carries its own `RequireAction` gate and must stay reachable by the warehouse role.

- [ ] **Step 6: Run the tests**

Run: `go test ./internal/http/handler/ -v -run 'TireAssignment'` then `go test ./...`
Expected: the new test PASSes; 14 packages `ok`.

- [ ] **Step 7: Regenerate the OpenAPI spec**

The DTO changed, so `docs/` drifts. Run:

```bash
swag init -g main.go \
  -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage \
  --parseDependency --parseInternal -o docs
git diff --stat docs
```

Expected: `docs/` updates to drop the four fields. `openapi_spec_test.go` must still pass.

- [ ] **Step 8: Commit**

```bash
git add internal/http/dto/tire_assignment_request.go \
        internal/http/handler/tire_assignment_request.go \
        internal/http/handler/router.go \
        internal/http/handler/validation_test.go docs
git commit -m "fix(tires): approval state is not editable through the generic update"
```

---

## Task 5: Deleting a foreign line item is a 404, not a silent recompute (CRIT-3, phase 1)

**Files:**
- Modify: `internal/db/queries/work_order_line_item.sql:27`
- Modify: `internal/db/queries/work_order_sub_line_item.sql:27`
- Modify: `internal/db/queries/purchase_order_line_item.sql:27`
- Modify: `internal/db/queries/service_entry_line_item.sql:27`
- Modify: the four matching `Delete` methods in `internal/http/handler/`
- Create: `internal/http/handler/crosstenant_integration_test.go`

**Interfaces:**
- Consumes: `middleware.CompanyFromContext`, `apierr.NotFound`.
- Produces: `qtx.Delete*LineItem(...) (int64, error)` — the regenerated sqlc signature returns rows affected.

The DELETE statements **are** correctly tenant-scoped. The bug is that they are `:exec`, so a foreign `parent_id` matches zero rows and returns `nil` — and the handler then calls the unscoped `recalc*` with the attacker's `parentID`, rewriting another tenant's totals and clearing their audited override, behind a 204.

This task closes the reachable path. Task 6 removes the class.

- [ ] **Step 1: Write the failing integration test**

Create `internal/http/handler/crosstenant_integration_test.go`:

```go
//go:build integration

package handler

import (
	"context"
	"net/http"
	"testing"
)

// Deleting a line item under a work order belonging to another company must be
// a 404. Before this was fixed the scoped DELETE matched zero rows, returned
// nil, and the handler went on to recompute the *victim's* document — so the
// attacker's 204 rewrote another tenant's totals and cleared their audited
// total override.
func TestDeletingAnotherCompanysLineItemIs404(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptestServer(t, newIntegrationRouter(pool))
	base := srv.URL

	platformToken := seedPlatformAdmin(t, ctx, pool, base)

	// Two independent tenants.
	_, _, _, tokenA := provisionAccountOwner(t, base, platformToken)
	_, _, _, tokenB := provisionAccountOwner(t, base, platformToken)
	companyA := createCompany(t, base, tokenA, "Company A", "AAA010101AAA")
	companyB := createCompany(t, base, tokenB, "Company B", "BBB010101BBB")
	tokenA = switchCompany(t, base, tokenA, companyA)
	tokenB = switchCompany(t, base, tokenB, companyB)

	// B owns a work order with a line item and a total.
	victimWO := createWorkOrder(t, base, tokenB)
	beforeTotal := workOrderTotal(t, base, tokenB, victimWO)

	// A aims a delete at B's work order.
	doJSON(t, http.MethodDelete,
		base+"/api/v1/work-orders/"+itoa(victimWO)+"/line-items/1",
		tokenA, nil, http.StatusNotFound, nil)

	if after := workOrderTotal(t, base, tokenB, victimWO); !after.Equal(beforeTotal) {
		t.Errorf("company B's total changed from %s to %s after company A's delete",
			beforeTotal, after)
	}
}
```

Before running this, read `integration_test.go` and reuse its real helpers. `httptestServer`, `createWorkOrder`, `workOrderTotal` and `itoa` do **not** exist yet — either add them to this file or replace them with the equivalents already in `integration_test.go` (`newIntegrationRouter`, `postJSON`, `getJSON`, `doJSON`, `provisionAccountOwner`, `createCompany`, `switchCompany` all exist and are listed at `integration_test.go:167-527`).

- [ ] **Step 2: Run it to confirm it fails**

Run: `go test -tags integration ./internal/http/handler/ -run TestDeletingAnotherCompanysLineItemIs404 -v`
Expected: FAIL — the delete returns 204, not 404.

- [ ] **Step 3: Change the four queries to `:execrows`**

In each of the four files, change line 27's magic comment only — the SQL body stays exactly as it is:

```sql
-- name: DeleteWorkOrderLineItem :execrows
```

```sql
-- name: DeleteWorkOrderSubLineItem :execrows
```

```sql
-- name: DeletePurchaseOrderLineItem :execrows
```

```sql
-- name: DeleteServiceEntryLineItem :execrows
```

- [ ] **Step 4: Regenerate**

Run: `sqlc generate`
Expected: the four `Delete*` methods in `internal/db/gen` now return `(int64, error)`.

- [ ] **Step 5: Update the four handlers**

Each `Delete` follows the identical shape. For `internal/http/handler/work_order_line_item.go`:

```go
func (s *WorkOrderLineItemStore) Delete(ctx context.Context, parentID, id int64) error {
	now := time.Now().UTC()
	return inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		// The DELETE is tenant-scoped, so zero rows means the line item, its
		// parent, or the tenancy between them did not match. Recomputing on
		// that parentID anyway would recompute a document the caller was
		// never allowed to touch.
		n, err := qtx.DeleteWorkOrderLineItem(ctx, gen.DeleteWorkOrderLineItemParams{
			ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx),
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return apierr.NotFound("line item not found")
		}
		return recalcWorkOrder(ctx, qtx, parentID, now)
	})
}
```

Apply the same shape to:
- `purchase_order_line_item.go` — `DeletePurchaseOrderLineItem`, then `recalcPurchaseOrder`
- `service_entry_line_item.go` — `DeleteServiceEntryLineItem`, then `recalcServiceEntry`
- `work_order_sub_line_item.go` — `DeleteWorkOrderSubLineItem`, then `recalcWorkOrderLineItem`

Add `"fleet/internal/platform/apierr"` to the imports of any of those files that lacks it.

- [ ] **Step 6: Run the tests**

Run: `go test ./...` then
`go test -tags integration ./internal/http/handler/ -run TestDeletingAnotherCompanysLineItemIs404 -v`
Expected: 14 packages `ok`; the integration test PASSes.

- [ ] **Step 7: Confirm no generated drift beyond the intended change**

Run: `sqlc generate && git diff --stat internal/db/gen`
Expected: only the four `Delete*` functions changed.

- [ ] **Step 8: Commit**

```bash
git add internal/db/queries internal/db/gen internal/http/handler
git commit -m "fix(money): a delete that matched no rows is a 404, not a silent recompute"
```

---

## Task 6: The recompute carries its tenant (CRIT-3, phase 2)

**Files:**
- Modify: `internal/db/queries/money.sql` — tenant predicates
- Modify: `internal/http/handler/money.go:26,60,92,137,164,186` — signatures
- Modify: 21 call sites, enumerated below

**Interfaces:**
- Consumes: `middleware.CompanyFromContext(ctx) int64`.
- Produces:
  - `recalcWorkOrder(ctx, qtx, workOrderID, companyID int64, at time.Time) error`
  - `recalcServiceEntry(ctx, qtx, entryID, companyID int64, at time.Time) error`
  - `recalcPurchaseOrder(ctx, qtx, orderID, companyID int64, at time.Time) error`
  - `recalcWorkOrderLineItem(ctx, qtx, lineItemID, companyID int64, at time.Time) error`
  - `recalcPurchaseOrderLineItem(ctx, qtx, lineItemID, companyID int64, at time.Time) error`
  - `recalcServiceEntryLineItem(ctx, qtx, lineItemID, companyID int64, at time.Time) error`

Task 5 closed the path that is reachable today. The recompute family is still unscoped by design — `money.sql:112-115` says the scope "is already established by the write that triggered it". That invariant is invisible, unenforced, and was already violated once. This task makes it structural.

- [ ] **Step 1: Add tenant predicates to the recompute queries**

In `internal/db/queries/money.sql`, add a `company_id` argument to each of these and delete the comment block at lines 112-115 that justified their absence:

```sql
-- name: GetWorkOrderByID :one
SELECT * FROM work_order WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id);

-- name: GetServiceEntryByID :one
SELECT * FROM service_entry WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id);

-- name: GetPurchaseOrderByID :one
SELECT * FROM purchase_order WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id);

-- name: StoreWorkOrderTotals :exec
UPDATE work_order SET
    parts_subtotal = sqlc.arg(parts_subtotal),
    labor_subtotal = sqlc.arg(labor_subtotal),
    subtotal       = sqlc.arg(subtotal),
    total_amount   = sqlc.arg(total_amount),
    updated_at     = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id);

-- name: StoreServiceEntryTotals :exec
UPDATE service_entry SET
    parts_subtotal = sqlc.arg(parts_subtotal),
    labor_subtotal = sqlc.arg(labor_subtotal),
    subtotal       = sqlc.arg(subtotal),
    total_amount   = sqlc.arg(total_amount),
    updated_at     = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id);

-- name: StorePurchaseOrderTotals :exec
UPDATE purchase_order SET
    subtotal     = sqlc.arg(subtotal),
    total_amount = sqlc.arg(total_amount),
    updated_at   = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id);

-- name: ClearWorkOrderTotalOverride :exec
UPDATE work_order SET
    total_override = NULL, total_override_reason = '', total_override_by_id = NULL, total_override_at = NULL
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id) AND total_override IS NOT NULL;

-- name: ClearPurchaseOrderTotalOverride :exec
UPDATE purchase_order SET
    total_override = NULL, total_override_reason = '', total_override_by_id = NULL, total_override_at = NULL
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id) AND total_override IS NOT NULL;

-- name: ClearServiceEntryTotalOverride :exec
UPDATE service_entry SET
    total_override = NULL, total_override_reason = '', total_override_by_id = NULL, total_override_at = NULL
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id) AND total_override IS NOT NULL;
```

The line-sum queries (`SumWorkOrderLineCosts`, `SumWorkOrderSubLineCosts`, `SumPurchaseOrderLineCosts`, `SumServiceEntryLineCosts`) and the line-total writes (`StoreWorkOrderLineItemTotals`, `StorePurchaseOrderLineItemSubtotal`, `StoreServiceEntryLineItemSubtotal`) reach rows only through a parent id that the caller has now proven. Leave them, and replace the deleted comment with one that says exactly that.

- [ ] **Step 2: Regenerate and confirm the build breaks at every call site**

Run: `sqlc generate && go build ./... 2>&1 | head -40`
Expected: compile errors at each `recalc*` call — this is the checklist for Step 4.

- [ ] **Step 3: Thread the tenant through `money.go`**

Each of the six functions takes `companyID int64` after its id parameter and passes it to every query. For `recalcWorkOrder`:

```go
func recalcWorkOrder(ctx context.Context, qtx *gen.Queries, workOrderID, companyID int64, at time.Time) error {
	order, err := qtx.GetWorkOrderByID(ctx, gen.GetWorkOrderByIDParams{
		ID: workOrderID, CompanyID: companyID,
	})
	if err != nil {
		return err
	}
	// ... unchanged Compute(...) ...
	return qtx.StoreWorkOrderTotals(ctx, gen.StoreWorkOrderTotalsParams{
		ID:        workOrderID,
		CompanyID: companyID,
		// ... unchanged ...
	})
}
```

Apply the same to `recalcServiceEntry`, `recalcPurchaseOrder`, and the three `recalc*LineItem` functions. Where a `*LineItem` function looks its parent up, pass `companyID` on to the parent recompute.

- [ ] **Step 4: Update all 21 call sites**

Each already has the tenant in scope, either as a local or via `middleware.CompanyFromContext(ctx)`. Work through:

| File | Lines |
|---|---|
| `purchase_order.go` | 88, 135 |
| `purchase_order_line_item.go` | 71, 105, 126 |
| `service_entry.go` | 92, 144 |
| `service_entry_line_item.go` | 82, 122, 143 |
| `total_override.go` | 77, 128, 179 |
| `work_order.go` | 150, 250 |
| `work_order_line_item.go` | 75, 108, 131 |
| `work_order_sub_line_item.go` | 76, 104, 117 |

Prefer hoisting `company := middleware.CompanyFromContext(ctx)` once per method over calling it inline repeatedly.

- [ ] **Step 5: Build and test**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: clean; 14 packages `ok`.

- [ ] **Step 6: Extend the cross-tenant integration test**

Add to `crosstenant_integration_test.go` a case proving the recompute itself is now scoped — that a direct recompute trigger against a foreign document 404s rather than rewriting it. Mirror the structure of `TestDeletingAnotherCompanysLineItemIs404`, aiming a total-override clear (`total_override.go`) at the other tenant's work order.

Run: `go test -tags integration ./internal/http/handler/ -run 'CrossTenant|AnotherCompany' -v`
Expected: PASS.

- [ ] **Step 7: Confirm drift checks pass**

Run: `sqlc generate && git diff --exit-code internal/db/gen`
Expected: exit 0.

- [ ] **Step 8: Commit**

```bash
git add internal/db/queries/money.sql internal/db/gen internal/http/handler
git commit -m "fix(money): scope the recompute family to the caller's tenant"
```

---

## Task 7: Stored file references are validated against the caller's tenant (CRIT-4)

**Files:**
- Modify: `internal/http/handler/fileref.go`
- Modify: `internal/http/handler/storage_cleanup.go`
- Modify: `internal/http/handler/fuel_photo.go:145`, `media.go:188,190`
- Test: `internal/http/handler/fileref_test.go` (append)

**Interfaces:**
- Consumes: `storage.VisibilityPublic/VisibilityPrivate`, `middleware.CompanyFromContext`.
- Produces: `fileKeyForCompany(files storage.Storage, stored string, companyID int64) (string, bool)`.

`file` is a free-form client string on fuel photos, asset photos, company logos, inspection signatures and submission-item photos. `fileKey` accepts **any bare relative path** as a key with no tenant check, so a client can store a foreign key and then have the server either presign it (read another tenant's private receipt) or delete it (`reclaimObjects`).

`objectKey` (`upload.go:334`) always writes `uploads/<visibility>/<companyID>/<token>-<name>`. So the rule is: **anything under `uploads/` must carry the caller's own company segment.** Values not under `uploads/` are pre-convention legacy keys — public by definition, since private visibility was introduced together with the prefix — and stay allowed.

- [ ] **Step 1: Write the failing test**

Append to `internal/http/handler/fileref_test.go`:

```go
// A stored file reference is client-supplied on several resources. A key under
// uploads/ names a company in its path, and that company must be the caller's
// — otherwise a client can store another tenant's key and have the server
// presign it (reading their private receipts) or reclaim it (deleting them).
func TestFileKeyForCompanyRejectsForeignKeys(t *testing.T) {
	const me = int64(7)
	for _, tt := range []struct {
		name   string
		stored string
		wantOK bool
	}{
		{"own public key", "uploads/public/7/ab12-photo.jpg", true},
		{"own private key", "uploads/private/7/ab12-receipt.pdf", true},
		{"another tenant's private key", "uploads/private/9/ab12-receipt.pdf", false},
		{"another tenant's public key", "uploads/public/9/ab12-photo.jpg", false},
		{"uploads path with no company", "uploads/private/secret.pdf", false},
		{"uploads path with junk company", "uploads/private/x/secret.pdf", false},
		{"traversal into another tenant", "uploads/private/7/../9/secret.pdf", false},
		{"legacy pre-convention key", "media/2026/07/photo.jpg", true},
		{"empty", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := fileKeyForCompany(nil, tt.stored, me)
			if ok != tt.wantOK {
				t.Errorf("fileKeyForCompany(%q, %d) ok = %v, want %v",
					tt.stored, me, ok, tt.wantOK)
			}
		})
	}
}
```

- [ ] **Step 2: Run it to confirm it fails**

Run: `go test ./internal/http/handler/ -run TestFileKeyForCompanyRejectsForeignKeys`
Expected: FAIL — `undefined: fileKeyForCompany`.

- [ ] **Step 3: Implement the tenant-aware resolver**

Add to `internal/http/handler/fileref.go`:

```go
// fileKeyForCompany resolves a stored reference the way fileKey does, then
// refuses anything that belongs to another tenant.
//
// Every key this API writes comes from objectKey, which namespaces it as
// uploads/<visibility>/<companyID>/<token>-<name>. So a key under uploads/
// carries its owner in the path and is checkable. Anything else is a
// pre-convention key — necessarily public, since private visibility arrived
// with the prefix — and is left alone rather than orphaned.
//
// The check matters because the stored value is client-supplied on fuel
// photos, asset photos, company logos and inspection evidence: without it, a
// client can name another tenant's object and have the server presign it or
// delete it.
func fileKeyForCompany(files storage.Storage, stored string, companyID int64) (string, bool) {
	key, ok := fileKey(files, stored)
	if !ok {
		return "", false
	}
	if !strings.HasPrefix(key, uploadKeyRoot) {
		return key, true // pre-convention key: not ours to adjudicate
	}
	rest := strings.TrimPrefix(key, uploadKeyRoot)
	visibility, rest, found := strings.Cut(rest, "/")
	if !found || (visibility != "public" && visibility != "private") {
		return "", false
	}
	owner, _, found := strings.Cut(rest, "/")
	if !found {
		return "", false
	}
	id, err := strconv.ParseInt(owner, 10, 64)
	if err != nil || id != companyID {
		return "", false
	}
	return key, true
}

const uploadKeyRoot = "uploads/"
```

Add `"strconv"` to the imports. `fileKey` already routes through `storage.cleanKey`, which resolves `..` — confirm with the traversal case in the test; if it does not, normalise before the prefix check.

- [ ] **Step 4: Run the test to confirm it passes**

Run: `go test ./internal/http/handler/ -run TestFileKeyForCompanyRejectsForeignKeys -v`
Expected: all subtests PASS.

- [ ] **Step 5: Route the read and reclaim paths through it**

`fileReadURL` and `reclaimObjects` both need the tenant. Give each a `companyID int64` parameter and use `fileKeyForCompany`. Update the callers:

- `fuel_photo.go:145` and `media.go:188,190` — pass `middleware.CompanyFromContext(ctx)`.
- `storage_cleanup.go` — `reclaimObjects` and `fileOwner.reclaim` take the tenant; every `o.reclaim(...)` call site passes `middleware.CompanyFromContext(ctx)`.
- `upload.go:147,157` — these keys come straight from `objectKey` with the caller's own company, so pass the same `companyID` that built them.

Find every call site with:

```bash
grep -rn "reclaim(\|reclaimObjects(\|fileReadURL(\|fileReadURLPtr(" --include=*.go internal/
```

- [ ] **Step 6: Build and test**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: clean; 14 packages `ok`.

- [ ] **Step 7: Commit**

```bash
git add internal/http/handler
git commit -m "fix(storage): validate stored file references against the caller's tenant"
```

---

## Task 8: Rate limiting middleware (HIGH-8, app side)

**Files:**
- Create: `internal/http/middleware/ratelimit.go`
- Create: `internal/http/middleware/ratelimit_test.go`
- Modify: `internal/http/handler/router.go:49-54, 67-72, 74-75`

**Interfaces:**
- Consumes: `middleware.ClaimsOf`, `apierr.Abort`.
- Produces:
  - `func NewLimiter(rate int, per time.Duration) *Limiter`
  - `func (l *Limiter) Allow(key string) (ok bool, retryAfter time.Duration)`
  - `func RateLimit(l *Limiter, keyFn func(*gin.Context) string) gin.HandlerFunc`

There is no rate limiter and no `limit_req` in nginx. `/auth/login` is an unthrottled guessing oracle with no lockout, and each attempt costs ~80ms of bcrypt, so the same flood is a CPU-exhaustion attack.

In-process is the right scope for the current single-container deploy. `Allow` is kept behind a small interface boundary so a Redis-backed store can replace it without touching the middleware.

- [ ] **Step 1: Write the failing test**

Create `internal/http/middleware/ratelimit_test.go`:

```go
package middleware

import (
	"testing"
	"time"
)

func TestLimiterAllowsUpToTheRateThenRefuses(t *testing.T) {
	l := NewLimiter(3, time.Minute)
	for i := range 3 {
		if ok, _ := l.Allow("a"); !ok {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}
	ok, retryAfter := l.Allow("a")
	if ok {
		t.Error("the 4th request should have been refused")
	}
	if retryAfter <= 0 || retryAfter > time.Minute {
		t.Errorf("retryAfter = %v, want a positive duration within the window", retryAfter)
	}
}

func TestLimiterKeysAreIndependent(t *testing.T) {
	l := NewLimiter(1, time.Minute)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("first request for a should be allowed")
	}
	if ok, _ := l.Allow("b"); !ok {
		t.Error("b must not be refused because a used its budget")
	}
}

func TestLimiterRefillsOverTime(t *testing.T) {
	l := NewLimiter(1, 50*time.Millisecond)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("first request should be allowed")
	}
	if ok, _ := l.Allow("a"); ok {
		t.Fatal("second request in the same window should be refused")
	}
	time.Sleep(60 * time.Millisecond)
	if ok, _ := l.Allow("a"); !ok {
		t.Error("the budget should have refilled after the window elapsed")
	}
}
```

- [ ] **Step 2: Run to confirm it fails**

Run: `go test ./internal/http/middleware/ -run TestLimiter -v`
Expected: FAIL — `undefined: NewLimiter`.

- [ ] **Step 3: Implement the limiter**

Create `internal/http/middleware/ratelimit.go`:

```go
package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
)

// Limiter is a token bucket per key, refilling continuously at rate/per.
//
// In-process and therefore per-replica: with N replicas the effective limit is
// N x rate. That is the right trade for a single-container deployment, and the
// Allow boundary is where a shared store would slot in if that changes.
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64       // tokens per second
	burst   float64       // bucket capacity
	window  time.Duration // the period `rate` was expressed over
	now     func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

// NewLimiter allows `rate` requests per `per` for each key, bursting to `rate`.
func NewLimiter(rate int, per time.Duration) *Limiter {
	if rate < 1 {
		rate = 1
	}
	if per <= 0 {
		per = time.Minute
	}
	return &Limiter{
		buckets: make(map[string]*bucket),
		rate:    float64(rate) / per.Seconds(),
		burst:   float64(rate),
		window:  per,
		now:     time.Now,
	}
}

// Allow consumes a token for key, reporting whether it was available and, when
// it was not, how long until one is.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}

	b.tokens = min(l.burst, b.tokens+now.Sub(b.last).Seconds()*l.rate)
	b.last = now

	if b.tokens < 1 {
		need := (1 - b.tokens) / l.rate
		return false, time.Duration(need * float64(time.Second))
	}
	b.tokens--
	return true, 0
}

// reap drops buckets that have been full and idle for a whole window, so a
// long-lived process does not accumulate one entry per IP ever seen.
func (l *Limiter) reap() {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := l.now().Add(-2 * l.window)
	for k, b := range l.buckets {
		if b.last.Before(cutoff) {
			delete(l.buckets, k)
		}
	}
}

// StartReaper runs reap on a ticker until ctx-less shutdown; call once at wiring.
func (l *Limiter) StartReaper(every time.Duration) {
	go func() {
		for range time.Tick(every) {
			l.reap()
		}
	}()
}

// RateLimit refuses a request whose key has exhausted its budget.
func RateLimit(l *Limiter, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFn(c)
		if key == "" {
			c.Next()
			return
		}
		ok, retryAfter := l.Allow(key)
		if !ok {
			secs := int(retryAfter.Seconds()) + 1
			c.Header("Retry-After", fmt.Sprintf("%d", secs))
			apierr.Abort(c, apierr.New(http.StatusTooManyRequests, "rate_limited",
				"too many requests; retry shortly"))
			return
		}
		c.Next()
	}
}

// ByIP keys on the client address, for routes reached before authentication.
func ByIP(c *gin.Context) string { return "ip:" + c.ClientIP() }

// ByEmployee keys on the authenticated employee, falling back to the client
// address so an unauthenticated request is still bounded.
func ByEmployee(c *gin.Context) string {
	if claims, ok := ClaimsOf(c); ok {
		return fmt.Sprintf("emp:%d", claims.EmployeeID())
	}
	return ByIP(c)
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/http/middleware/ -v`
Expected: all PASS.

- [ ] **Step 5: Wire it into the router**

In `internal/http/handler/router.go`, after the existing middleware block:

```go
	// Rate limits. /auth is strict: it is unauthenticated, it is the only
	// route that can be guessed at, and each attempt costs a bcrypt compare.
	authLimiter := middleware.NewLimiter(20, time.Minute)
	apiLimiter := middleware.NewLimiter(600, time.Minute)
	authLimiter.StartReaper(5 * time.Minute)
	apiLimiter.StartReaper(5 * time.Minute)
```

Apply `middleware.RateLimit(authLimiter, middleware.ByIP)` to the four `/auth/*` routes, and `middleware.RateLimit(apiLimiter, middleware.ByEmployee)` to the `api` group — placed **after** `middleware.Auth` so `ByEmployee` can see the claims:

```go
	api := r.Group("/api/v1")
	api.Use(middleware.Auth(d.Tokens),
		middleware.RateLimit(apiLimiter, middleware.ByEmployee),
		middleware.RequireIdentity(NewIdentityLoader(d.Queries)))
```

- [ ] **Step 6: Add a router-level test**

Append to `internal/http/handler/router_test.go` a test that fires 21 requests at `/auth/login` through the assembled router and asserts the last one is `429` with a `Retry-After` header. Follow the existing construction style in that file.

Run: `go test ./internal/http/handler/ -run 'RateLimit|Router' -v`
Expected: PASS.

- [ ] **Step 7: Full suite**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: 14 packages `ok`.

- [ ] **Step 8: Commit**

```bash
git add internal/http/middleware/ratelimit.go \
        internal/http/middleware/ratelimit_test.go \
        internal/http/handler/router.go internal/http/handler/router_test.go
git commit -m "feat(api): rate-limit authentication and the API surface"
```

---

## Task 9: Rate limiting at the edge (HIGH-8, nginx)

**Files:**
- Modify: `deploy/nginx/go-logistics.conf`

**Interfaces:**
- Consumes: nothing.
- Produces: nothing consumed by Go code.

`limit_req_zone` is only valid in nginx's `http` context. This file's top level contains `server { }` blocks, which are themselves only valid inside `http { }` — so the file is included at http level and the zone declaration can go at the top of the managed block, above the first `server`.

The block is copied by hand into `/opt/sga/nginx/app.conf` between markers, so the zone must live inside those markers or it will be lost on the next copy.

- [ ] **Step 1: Add the zones and apply them**

At the top of `deploy/nginx/go-logistics.conf`, after the header comment and before the first `server {`:

```nginx
# Rate-limit zones. These are http-context directives; this file's contents sit
# inside the sga stack's http block (that is what makes the server blocks below
# valid), so they belong here — and inside the managed markers, or the next
# copy onto the server drops them.
#
# Keyed on $binary_remote_addr rather than a header: X-Forwarded-For is
# client-appendable, so keying on it would let a caller mint a fresh budget per
# request. This is the second layer; the app limits per employee.
limit_req_zone $binary_remote_addr zone=golog_auth:10m rate=20r/m;
limit_req_zone $binary_remote_addr zone=golog_api:10m  rate=600r/m;
limit_req_status 429;
```

Inside the `443` server block, alongside the existing API `location`:

```nginx
    # Unauthenticated and guessable; burst kept small deliberately.
    location /auth/ {
        limit_req zone=golog_auth burst=10 nodelay;
        # ... the same proxy_pass / proxy_set_header lines the API location uses
    }
```

and on the `/api/` location: `limit_req zone=golog_api burst=100 nodelay;`

Read the existing `location` blocks first and copy their `resolver`/`set`/`proxy_pass` shape exactly — the variable-in-`proxy_pass` trick is load-bearing (it forces per-request DNS re-resolution, without which every deploy 502s).

- [ ] **Step 2: Validate the config**

Run: `docker run --rm -v "$PWD/deploy/nginx:/etc/nginx/conf.d:ro" nginx:alpine nginx -t 2>&1 | tail -5`

This will complain that `server` blocks need an `http` context unless wrapped. If so, validate by wrapping a copy:

```bash
{ echo 'events {} http {'; cat deploy/nginx/go-logistics.conf; echo '}'; } > /tmp/nginx-check.conf
docker run --rm -v /tmp/nginx-check.conf:/etc/nginx/nginx.conf:ro nginx:alpine nginx -t
```

Expected: `syntax is ok` / `test is successful`.

- [ ] **Step 3: Document the deploy step**

The file's header already documents the copy-and-reload procedure. Add one line noting that the `limit_req_zone` directives must land inside the markers.

- [ ] **Step 4: Commit**

```bash
git add deploy/nginx/go-logistics.conf
git commit -m "fix(deploy): rate-limit auth and api at the edge"
```

---

## Task 10: Verify the whole branch and update the review report

**Files:**
- Modify: `../../code-review-2026-09-12.md` (beside the repo)

- [ ] **Step 1: Full verification**

```bash
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
sqlc generate && git diff --exit-code internal/db/gen
swag init -g main.go -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage --parseDependency --parseInternal -o docs
git diff --exit-code docs
```

Expected: every command exits 0.

- [ ] **Step 2: Integration suite**

Run: `go test -tags integration ./internal/http/handler/`
Expected: `ok`, including both new cross-tenant tests.

- [ ] **Step 3: Race detector on the new concurrent code**

`Limiter` is the first shared mutable state added by this branch.

Run: `go test -race ./internal/http/middleware/`
Expected: `ok`, no race reports. If cgo is unavailable, say so rather than claiming it passed.

- [ ] **Step 4: Update the review report**

Mark CRIT-3, CRIT-4, CRIT-6, CRIT-7, CRIT-8, HIGH-3 and HIGH-8 as fixed in the §0 status table and remove them from the §9 stop-ship list, leaving the "This sprint" tier intact.

- [ ] **Step 5: Merge back**

```bash
cd "/c/Users/lider/OneDrive/Desktop/PROJECTS/GoLogistics/Go-Logistics"
git merge fix/stopship-findings
git worktree remove "$WT"
```

Do not push — that is the user's call.

---

## Self-Review

**Spec coverage:** CRIT-3 → Tasks 5 and 6. CRIT-4 → Task 7. CRIT-6 → Task 4. CRIT-7 → Task 3. CRIT-8 → Task 1. HIGH-3 → Task 2. HIGH-8 → Tasks 8 and 9. All seven covered.

**Type consistency:** `recalc*` signatures in Task 6's Interfaces block match the call-site table and the `money.go` edits. `fileKeyForCompany` in Task 7 matches its test. `NewLimiter`/`Allow`/`RateLimit`/`ByIP`/`ByEmployee` in Task 8 match the test and the router wiring.

**Known gaps, deliberately left for the executor to resolve against the code:**
- Task 4 Step 4 assumes a `GetTireAssignmentRequest` query exists; the step says to verify the generated name and params first.
- Task 5 Step 1's integration test names three helpers that do not exist yet; the step says to reuse `integration_test.go`'s real helpers instead.
- Task 7 Step 3 assumes `cleanKey` normalises `..`; the step says to confirm via the traversal test case.

These are flagged rather than guessed because getting them wrong silently would produce a passing test that proves nothing.
