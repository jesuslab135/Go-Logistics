# BACKEND-REQUIREMENTS implementation record

- **Date:** 2026-08-14
- **Scope:** BR-01 … BR-20 from the consolidated `BACKEND-REQUIREMENTS.md` handoff
- **Base:** `9d201e5` → 19 commits
- **Verification:** `go build`, `go vet`, `gofmt -l`, `go test ./...` green on every
  commit individually (checked out into a throwaway worktree, not just on the final
  tree). `sqlc generate` and `swag init` both reproduce the tree byte-for-byte.

The source corpus this handoff was derived from (`docs/backend-migration/*`) is not
present in this repository, so the per-route READMEs the completion checklist names
could not be updated. This file records the same information instead.

---

## What landed, by requirement

### P0

**BR-01 — Provision employee credentials.** `POST /api/v1/employees/{id}/set-password`
(admin-gated, inside the `employees` module group). There is no mail infrastructure in
this stack, so the spec's `set-password` alternative was taken rather than an invite
link. The target is resolved through the tenant-scoped employee lookup, so an employee
of another company returns 404 rather than leaking existence. `password_hash` is neither
accepted nor returned.

Also adds `bindJSONValidated`, which turns binding failures into 422 with per-field
details keyed by the JSON name the client sent — a weak password now reports against
`password` instead of one opaque 400.

**BR-02 — Switch the active company.** `POST /auth/switch-company` re-issues both tokens
scoped to another company. Membership is re-read from `employee_companies` rather than
trusted from the presented token, and `is_admin` is resolved against the target
company's role, so a stale claim cannot widen access. Non-members get 403. Authenticated
via `middleware.Auth` only — deliberately no company gate, since the caller is leaving
the company the token is scoped to.

**BR-03 — Approve tire-assignment requests transactionally.**
`POST /api/v1/tire-assignment-requests/{id}/approve`. In one transaction: the request
and tire rows are locked, the request is resolved to `APPROVED` with resolver and server
time, the installation is created, an `INSTALL` mount log is appended, and the tire's
status/vehicle/position are updated.

Idempotency comes from the state guard on the UPDATE (`WHERE … AND state = 'PENDING'`):
a retry matches no row and reports 409 instead of creating a second installation.
Conflicts carry stable codes — `request_already_resolved`, `tire_already_mounted`,
`tire_unavailable`, `position_occupied` — instead of letting the unique constraints
surface as a 500.

Gated on `tire_approvals/approve` via the new `middleware.RequireAction`, which lets a
route require a named permission rather than the one its HTTP method implies.

**BR-04 — Replace legacy `auth_user` references.** Migration `000007`.
`fuel_comment.user_id` → `employee_id`; that column plus `fuel_photo.uploaded_by_id`
and `media.uploaded_by_id` now carry real employee foreign keys.
`employee.user_id` preserves the original `auth_user` id and is the mapping key.
Unmappable rows become NULL — explicitly unattributed rather than credited to the wrong
employee — which is why the two NOT NULL columns were relaxed. Authorship and upload
attribution are now stamped from the authenticated context and are not reassignable by a
later edit.

`inventory_journal_entry.user_id` already referenced `employee(id)` and needed no work;
those four columns were the complete set (`grep 'auth_user' 000001_init.up.sql`).

### P1

**BR-05 — Revoke refresh tokens on logout.** Refresh tokens carry a unique `jti`;
`revoked_token` (migration `000006`) holds revoked identifiers until their natural
expiry. `POST /auth/logout` revokes the presented token; refresh parsing rejects revoked
ones, and also rejects pre-BR-05 tokens, which carry no `jti` and could never be revoked.
The `jti` is bound to its own employee, so a caller can only revoke their own session.
Expired rows are pruned opportunistically on logout.

**BR-06 / BR-07 / BR-09 / BR-11 / BR-12 / BR-19 — filtered collections.** See
"The list-query mechanism" below. Endpoints and their filters:

| Route | Filters | Ordering |
|---|---|---|
| `GET /work-orders` | `asset_id`, `status_id` | `order=issued_at,id`; default `-issued_at` |
| `GET /issues` | `asset_id`, `state` | `order=created_at,due_date,state,id`; default `-created_at` |
| `GET /inventory-journal-entries` | `part_id`, `created_from`, `created_to` | fixed `created_at DESC` |
| `GET /assets` | `q` (name/VIN/plate), `vehicle_type`, `status_id` | `order=name,vin_sn,vehicle_type,updated_at,id`; default `name` |
| `GET /asset-trailer-assignments` | `trailer_id`, `asset_id`, `is_active` | fixed `assigned_date DESC` |
| `GET /part-inventory` | `part_id` | fixed `id` |
| `GET /purchase-order-line-items` | `purchase_order_id`, `part_id` | fixed `position, id` |
| `GET /fuel-entries` | `asset_id`, `fuel_type`, `date_from`, `date_to` | `order=date,odometer,quantity,total_cost,updated_at,id`; default `-date` |
| `GET /tire-mount-logs` | `event_date_from/to`, `event_type`, `tire_id`, `vehicle_id`, `position_code`, `performed_by_id` | fixed `event_date DESC` |
| `GET /catalog-options` | `category` | fixed `category, value, id` |
| `GET /media` | `asset_id` | fixed `created_at DESC` |
| `GET /purchase-orders` | `vendor_id`, `state` | `order=created_at,state,id`; default `-created_at` |
| `GET /tire-assignment-requests` | `state`, `tire_id`, `vehicle_id`, `requested_from/to` | `order=requested_at,resolved_at,id`; default `-requested_at` |
| `GET /comments` | `content_type`, `object_id` | fixed `created_at DESC` |

**BR-07** additionally denormalizes `trailer_type`, `trailer_classification` and
`trailer_size` onto `AssetResponse` (the spec's preferred option over a separate
`/trailers` route), and **BR-09** adds `operator` the same way. A trailer registry —
including its DRY_VAN / CONTAINER / CHASSIS / DOLLY tabs — and the operator column now
come from one paginated request. Single-asset responses fill the same fields, so list
and detail shapes agree. The authoritative records stay at `/assets/{id}/vehicle` and
`/trailer`.

**BR-08** is the reverse trailer lookup. A partial unique index already allows at most
one active assignment per trailer, so `?trailer_id=N&is_active=true` identifies the
current tow unambiguously. **BR-19** also adds `trailer_name` to
`AssetTrailerAssignmentResponse`, joined from the trailer asset; nested create/update
re-read the row to return it, since an INSERT cannot return a joined column.

**BR-10 — Nested collections pageable and complete.** `GET /part-inventory` and
`GET /purchase-order-line-items` expose rows previously reachable only under a parent, so
inventory-journal and sub-line-item forms can offer them as foreign-key sources instead
of a raw numeric id field. Both are read-only; mutations stay on the nested routes that
own the parent scope.

Every list query now ends its `ORDER BY` with the row id. Without a total order, rows
equal on the sort key — a line item's `position`, a status log's timestamp — can repeat
or disappear between pages, which is what made a truncated child list look like a wrong
total. List operations also now document `limit`/`offset`, the convention the paginator
has always applied; the `page`/`page_size` spelling in the older stubs was the stale
dialect the handoff warned about.

**BR-13 — Upload type and size.** Content-sniffed type checking (415) and the streaming
size limit (413) were already in place. What was missing is the per-resource policy: a
caller now declares `purpose` (`photo`, `document`, `generic`) and the allowlist narrows
accordingly, so a PDF cannot be stored where a gallery expects an image. A purpose can
only remove types from the configured allowlist, never add one; an unrecognised value is
refused with `invalid_upload_purpose` rather than silently treated as generic.

### P2

**BR-14 — Comment content types.** Migration `000008`. `content_type_id` held a row id
from `django_content_type` — assigned at migrate time, never ported, not reproducible in
a fresh database. It is replaced by a `content_type` string with a declared, server-
validated vocabulary: `asset`, `issue`, `work_order`, `service_entry`. Creating a comment
verifies the parent exists and belongs to the caller's company. The author is stamped
from context; an edit changes only the body. `object_id` is widened to `bigint` — it
points at `bigserial` primary keys, so Django's `PositiveIntegerField` was always too
narrow. Historic rows keep an empty `content_type`: an explicit unknown, not a guess.

**BR-15 — Notifications.** Migration `000009` plus a real table; the endpoint previously
returned a hardcoded empty list. Rows are scoped to the employee *within* a company, so
someone working for two companies does not see one company's notifications while acting
for the other. `POST /notifications/{id}/read` is idempotent (`COALESCE` preserves the
original `read_at`) and the recipient is part of the update predicate, so a caller can
only mark their own.

Two producers, each writing in the same transaction as the event it describes, so a
notification can never announce a rolled-back change: filing a tire assignment request
notifies everyone who could approve it (admins, and roles granting
`tire_approvals/approve`), and approving one notifies the requester.

Creating a request no longer accepts `state`, requester, resolver or timestamps from the
body — it was possible to file a request that was already `APPROVED`, which the approve
action would then be handed as a resolved request.

**BR-16 — Dashboard contract.** `upcoming_days` is returned, so UI copy can name the
window the backend actually counted over. The two misleading counter names are kept —
the generated client already uses them — and their calculations are documented in
OpenAPI and beside the DTO: `pending_inspections` counts submissions with failed items
and no issue raised from them; `low_stock_parts` counts inventory rows (a part at one
location), not distinct parts. Trends are not implemented: real ones need periodic
snapshots, and deriving a previous period from current-state counters would invent it.

**BR-17 — Identity flags.** `is_technician` and `is_vehicle_operator` on
`dto.MeEmployee`. They are operational flags rather than permissions — which capture
flows apply to this person — and a client cannot infer them from the permission map.

**BR-18 — Fuel distance and efficiency.** Derived server-side and no longer accepted
from the request. A client cannot compute them without holding the asset's whole
history, and two clients disagreeing would corrupt the series.

The series is one asset and one fuel type ordered by `(date, id)`, so diesel and DEF
fills for the same truck stay independent. Ported from Django's
`FuelRecalculationService`, including every case where a value cannot be known: no
previous entry or `reset` (a new baseline), a non-positive odometer delta (rollback or
correction), either tank not full, and a zero quantity. Distance is still recorded where
only efficiency is unknowable.

Because each entry depends on the one before it, every write recomputes the entries that
follow, inside the write's own transaction with those rows locked. That covers backdated
inserts, edits that move an entry between dates, and edits that move it to another asset
or fuel type (both the series it left and the one it joined are repaired). **Deletion
recomputes too, which Django never did** — it left stale distances behind every removal.

**BR-20 — Reclaim orphaned upload objects.** Policy: delete the old object once the row
that referenced it has been written. Covers all six URL-bearing columns (asset photo,
company logo, media file + thumbnail, fuel photo, inspection signature, inspection item
photo).

The delete happens *after* the write, not inside it: an object store cannot join a
database transaction, and deleting first would destroy a live file if the write then
failed. The cost is that a crash between the two leaves one orphan — far better than
removing something still in use. Failures are logged and swallowed for the same reason.
Only objects this storage owns are touched; a URL it cannot resolve to a key (an
externally hosted image) is left alone.

---

## The list-query mechanism

sqlc generates one fixed SQL string per query, so a resource whose filters vary per
request cannot be served from it. `internal/http/handler/listquery.go` assembles those
statements instead, under three rules that hold by construction rather than by
convention:

1. **Nothing from the request is concatenated into SQL.** Table and alias names must be
   plain identifiers (`^[a-z_][a-z0-9_]*$`); code-owned join and select fragments cannot
   carry a statement terminator; order columns come from an allow-list; values are always
   bound parameters. A violation panics at the call site, where it is a wiring bug.
2. **The projection names the row struct's columns explicitly, in field order**, derived
   by reversing sqlc's column→field naming. Positional scanning is therefore correct by
   definition rather than relying on `*` happening to line up, and
   `TestTableColumnsMatchSchema` pins the derivation against real tables. A projection
   that does not match its row struct panics at build time rather than scanning a column
   into the wrong field.
3. **The count reuses the same FROM and WHERE**, so a reported total always describes the
   filtered set. Joins that could fan out can opt into `count(DISTINCT …)`.

This also put `internal/platform/filter` to work; it had been written earlier with no
callers. `filter.Raw` was added for shapes `Add` cannot express (the asset search spans
three columns), still numbering its own placeholders.

`registerCrudWithList` wires the standard CRUD shape but serves the collection from a
custom handler — following the existing precedent in `registerCompanyRoutes` and the
`crud` package's own advice to "wire the exported handlers individually" for partial
resources.

---

## Deliberate departures from Django

The handoff's precedence rules put a Go migration document above Django behavior, and
these are the places the port intentionally diverges:

- **Tire approval** accepts installation readings in the body (Django ignored the body
  entirely), returns 409 with stable codes (Django returned 400), and updates
  `tire.current_vehicle_id` / `current_position_code` (Django left them stale).
- **Fuel deletion** recomputes the following entries (Django did not).
- **Comment authorship, fuel comment/photo attribution and media attribution** are
  stamped server-side (Django accepted them from the client).
- **Tire request creation** stamps state, requester and timestamps (Django accepted
  `state` from the body).

Values ported verbatim from Django, since the schema stores them as free varchars with no
CHECK constraint and they are the only written record: tire status
`IN_STOCK|MOUNTED|REPAIR|DISPOSED`, mount events `INSTALL|DISMOUNT|ROTATION`, request
states `PENDING|APPROVED|REJECTED`, issue states `OPEN|RESOLVED|CLOSED`, purchase order
states `DRAFT|PENDING_APPROVAL|REJECTED|APPROVED|PURCHASED|RECEIVED_PARTIAL|RECEIVED_FULL|CLOSED`.
They live in `internal/domain/tire/constants.go` and beside the handlers that validate them.

---

## Not implemented, on purpose

Everything in the handoff's "Product-dependent backend work" table (CB-01 … CB-11) is
untouched, pending the decisions in `PENDING-PRODUCT-DECISIONS.md`. Most visibly, there
is **no reject action** on tire-assignment requests (CB-05) even though approve landed:
its audience and audit level are still open. Also absent: purchase-order state actions,
public signup, signed download URLs, inspection-completion, work-order status logging,
archive/unarchive, primary-photo semantics, cross-company admin APIs.

Dashboard trends were declined on their merits (see BR-16), not deferred.

---

## Migrations added

| Migration | Purpose |
|---|---|
| `000006_revoked_token` | Refresh-token denylist (BR-05) |
| `000007_legacy_user_refs` | `auth_user` ids → employee foreign keys (BR-04) |
| `000008_comment_content_type` | Go-owned comment discriminator, `object_id` widened (BR-14) |
| `000009_notification` | Notification persistence (BR-15) |

`000007` and `000008` are lossy in the down direction and say so in their headers: the
Django id spaces they replace were never reproducible here.

---

## Follow-ups

- **Live verification.** Every change is covered by unit tests and by the route-tree
  tests, but the handoff's completion checklist also asks for a live request or browser
  smoke test per item. That still needs to run against the compose stack (`fleet-pg` on
  5433, `fleet-minio` on 9010) with migrations `000006`–`000009` applied.
- **Client regeneration.** The frontend's generated client should be regenerated against
  the updated OpenAPI. Breaking response/request changes to check: comment
  `content_type`/`object_id`, fuel entry create/update (no `miles_traveled` /
  `fuel_efficiency`), fuel comment `employee_id`, fuel photo and media (no
  `uploaded_by_id` on create), tire assignment request create (four fields only).
