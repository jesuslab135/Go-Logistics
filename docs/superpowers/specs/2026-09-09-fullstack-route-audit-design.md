# Full-stack route audit — design

**Date:** 2026-09-09
**Status:** approved, pending implementation plan

## Goal

Verify that every route of the deployed Go Logistics frontend actually works
against the deployed Go backend, and produce one documented Markdown file
listing every frontend and backend fix the audit turns up.

"Works" means: the screen loads, the requests it makes are accepted, the
responses match what the OpenAPI spec promises, and the authorization and
multi-tenancy gates hold.

## Context

Established by inspection before this design was written:

| Fact | Value |
|---|---|
| Frontend | https://go-logistics.netlify.app — React + TanStack Router |
| Backend | https://go-logistics.jesuslab135.com — `healthz` 200, `swagger` 200 |
| Backend surface | 186 paths, 374 operations (`docs/swagger.json`) |
| Frontend surface | 58 routes, extracted from the production JS bundle |
| Test identity | employee 1, company 1; admin + account owner + platform admin |
| Test identity reach | all 17 modules, full CRUD, so all 374 operations are reachable |

Two constraints shape everything below.

**The frontend source is not available.** The only local frontend,
`PROJECTS/flotilla-duran`, is a different application (`itsAkzzl/flotilla-duran`)
still pointing at the superseded Django API `api.core.mikecardona076.com`. The
Netlify app is a separate codebase, visible only as minified JavaScript.
Frontend findings are therefore behavioural bug reports carrying request and
response evidence, not file-level patches. Backend findings cite `file:line`,
because that source is in this repository.

**The target is production.** There is no staging environment. The account
holds throwaway data and full CRUD was authorised, so the audit writes to the
live database. The Data safety section states how that is bounded and what
residual risk remains.

The test account belongs to exactly one company, so multi-tenancy has nothing
to test against. Account-owner rights allow `POST /api/v1/companies`, so the
audit creates a second tenant for that purpose and removes it afterwards.

## Architecture

Three components. Two run independently; the third joins them.

### Layer 1 — contract harness

A Node suite (node 25 is present; native `fetch`, no dependencies) driving the
API directly. Operations are enumerated from the local `docs/swagger.json`,
which is authoritative — it is generated from the same handlers the audit would
otherwise have to be read from.

This is deliberately not a generic spec fuzzer. The 374 operations sit on a
foreign-key dependency graph: a work-order line item cannot exist without a
work order, a wheel-position definition cannot exist without an axle definition
under an axle template. The sweep therefore runs in dependency order against a
fixture graph it builds first.

Modules:

- `spec.mjs` — load the spec, enumerate operations, resolve schemas
- `client.mjs` — login, token refresh, and a request wrapper that records every
  request and response pair for later evidence
- `fixtures.mjs` — build the dependency graph of test entities in creation
  order, plus the limited-role identity described below
- `sweep.mjs` — per-module operation suites
- `workflows.mjs` — the stateful suites
- `isolation.mjs` — the second-tenant suite
- `teardown.mjs` — reverse-order deletion
- `report.json` — machine-readable output consumed by Layer 3

### Layer 2 — UI harness

A claude-in-chrome session walking all 58 frontend routes with the real login.
Per route, record:

- every XHR and its status code
- console errors and warnings
- whether the page rendered data or fell into an error or empty state
- on form pages, the result of an actual submit

### Layer 3 — cross-reference

For each request the frontend emits, locate the matching backend operation and
diff the payload against the spec schema.

This is the component the audit exists for. It is what turns "the trailers page
is broken" into "the frontend sends `classification` as a string, the backend
binds `trailer_classification_id` as an integer, so the request 400s."

## Checks per operation

Nine checks, applied where meaningful to the operation:

| Check | Expectation |
|---|---|
| Unauthenticated | 401 |
| Module gate | 403 for a role lacking the module |
| Happy path | the documented 2xx |
| Response conformance | keys match the spec schema; no missing required fields, no undocumented extras |
| Validation | 400 on wrong types or missing required fields |
| Missing row | 404 |
| Tenant isolation | company 2 cannot read company 1's row |
| List contract | pagination, filters, `include_archived` |
| Retired verbs | 405 plus the replacement message (inventory journal PUT and DELETE) |

The module-gate check requires a non-admin identity. The fixture graph
therefore creates a limited role and an employee bound to it. Without that
identity every 403 path in the codebase goes untested, and `RequireModule`,
`RequireAction` and `RequirePlatformAdmin` are precisely where a mistake stops
being a bug and becomes a breach.

## Workflow suites

Stateful behaviour that per-operation checks structurally cannot reach:

- **Purchase order lifecycle** — draft, submit, approve or reject, receive,
  close. Every legal transition, every illegal transition rejected, status-log
  rows appended, `approve` and `reject` refused without the
  `purchase_orders.approve` action, `override-total` audited.
- **Work order** — a status change writes a status log; line items, sub-line
  items and labor entries nest correctly; totals recompute.
- **Inventory journal** — append-only enforced, and `reverse` produces a
  balancing entry.
- **Tire assignment** — request, then approve behind the `tire_approvals` gate.
- **Archive and restore** — across all five catalogs carrying `archived_at`,
  including list-hiding and the `include_archived` opt-in.
- **Uploads** — `POST /uploads`, presigned URL, media row, and confirmation
  that the `fleet-private` bucket returns 403 to an unsigned request. If that
  bucket is publicly readable it is the most serious finding available, and the
  audit proves it either way.

## Data safety

Every row the harness creates carries a `ZZ-TEST-<runid>-` prefix. Teardown
deletes in reverse dependency order. Rows pinned by foreign keys are archived
instead, which is the system's own answer to that case, and the report states
what remains.

The sweep mutates only rows it created. Existing company-1 data is read, never
written. The one deliberate exception is the second tenant,
`ZZ-TEST Isolation Co`, created and torn down.

Residual risk, stated plainly: this is a live production database receiving
several hundred writes. Nothing in the design is destructive, but a backend bug
discovered during the sweep could still have side effects on shared rows. That
is inherent to testing against production and was accepted when full CRUD was
authorised.

## Deliverable

`docs/2026-09-09_fullstack-route-audit.md`, following the repository's existing
`docs/YYYY-MM-DD_*.md` convention.

1. Executive summary — counts, severity histogram, the leading findings
2. Frontend route table — all 58 routes against a verdict of works, partial,
   broken, or not implemented
3. Backend operation table — all 374 against the same verdicts
4. Findings — each carrying an ID, severity, layer (frontend, backend, or
   contract), affected route, an exact reproduction as a runnable curl or
   numbered UI steps, observed versus expected behaviour, the real response
   body as evidence, and a recommended fix. Backend fixes cite `file:line`;
   frontend fixes specify the required request shape.
5. Untested and blocked appendix — what was not proven, and why. The file
   should admit a gap rather than imply coverage it does not have.
6. Teardown confirmation

Every finding is re-verified once with a fresh token before entering the
report. A stale-token 401 reads exactly like an authorization bug, and the
extra request is cheaper than a false positive.

## Harness location

`qa/route-audit/`, committed. The suite is re-runnable, and a harness that can
be pointed at the API after every deploy outlives the single report it produces
today.

## Phasing

| Phase | Work |
|---|---|
| 0 | Harness scaffold, auth, fixture dependency graph, limited-role identity |
| 1 | Layer 1 sweep across the 17 modules |
| 2 | Workflow suites |
| 3 | Isolation suite against the second tenant |
| 4 | Layer 2 browser walkthrough of the 58 frontend routes |
| 5 | Cross-reference, finding verification, report |
| 6 | Teardown and confirmation |

## Success criteria

- Every one of the 374 backend operations carries a verdict, or an explicit
  reason it could not be tested.
- Every one of the 58 frontend routes carries a verdict.
- Every finding is reproducible from the report alone, without access to the
  harness.
- Teardown leaves no `ZZ-TEST-` rows behind, or the report names those that
  survive and explains why.

## Appendix — canonical frontend route inventory

Extracted from the production bundle `assets/index-CRmilDIu.js`. `$id` denotes a
dynamic segment, resolved during the walkthrough against a fixture row. This is
the definitive list the 58-route success criterion is measured against.

- `/app/admin`
- `/app/admin/companies`
- `/app/admin/companies/$id`
- `/app/admin/users`
- `/app/dashboard`
- `/app/design-system`
- `/app/fuel`
- `/app/inspections`
- `/app/inspections/forms`
- `/app/inventory`
- `/app/inventory/inventory-adjustment-reasons`
- `/app/inventory/inventory-journal-entries`
- `/app/inventory/part-categories`
- `/app/inventory/part-locations`
- `/app/inventory/part-manufacturers`
- `/app/issues`
- `/app/issues/board`
- `/app/issues/faults`
- `/app/issues/priorities`
- `/app/locations`
- `/app/maintenance`
- `/app/maintenance/$id`
- `/app/organization`
- `/app/organization/employees`
- `/app/organization/groups`
- `/app/organization/roles`
- `/app/purchase-orders`
- `/app/purchase-orders/$id`
- `/app/service`
- `/app/service/entries`
- `/app/service/entries/$id`
- `/app/service/reminders`
- `/app/settings`
- `/app/settings/custom-fields`
- `/app/settings/fuel-types`
- `/app/settings/measurement-units`
- `/app/settings/service-tasks`
- `/app/settings/weekly-mileage-goals`
- `/app/tires`
- `/app/tires/$id`
- `/app/tires/assignment-requests`
- `/app/tires/axle-templates`
- `/app/tires/models`
- `/app/tires/movements`
- `/app/trailers`
- `/app/trailers/$id`
- `/app/trailers/$id/edit`
- `/app/trailers/new`
- `/app/vehicle-makes`
- `/app/vehicle-models`
- `/app/vehicles`
- `/app/vehicles/$id`
- `/app/vehicles/$id/edit`
- `/app/vehicles/new`
- `/app/vendors`
- `/login`
- `/onboarding/company`
- `/signup`
