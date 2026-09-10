# Full-stack route audit

**Run:** `0910c`  
**Target:** https://go-logistics.jesuslab135.com  
**Generated:** 2026-09-10T08:21:17.742Z

> **How to read this file.** The tables below the "Findings" section are generated
> by `qa/route-audit/run.mjs` from a live run. The Findings section itself, and the
> frontend analysis, are authored. Re-running `run.mjs report` regenerates the
> tables and would discard the authored sections — copy them out first.

## Executive summary

**The backend passed every security and behavioural check.** All 22 authorization-gate
probes, all 12 cross-tenant isolation probes, and all 38 stateful workflow checks
returned the expected result. There are **zero high-severity findings**.

The 109 failed checks are contract-level disagreements between the OpenAPI spec and
the implementation, concentrated in three patterns (F1–F3 below) that together account
for 90 of them.

| Metric | Value |
|---|---|
| Backend operations in the spec | 374 |
| Checks run | 1095 |
| Checks failed | 109 |
| Findings — high | **0** |
| Findings — medium | 99 |
| Findings — low | 10 |
| Module-gate probes failed | **0 of 22** |
| Tenant-isolation probes failed | **0 of 12** |
| Workflow checks failed | **0 of 38** |
| Operations exercised by the sweep | 176 |
| Operations accounted for by exclusion | 198 |
| Frontend routes walked in a browser | **0 — blocked, see below** |
| Frontend endpoints calling a nonexistent backend path | **0 of 167** |
| Backend paths with no frontend consumer | 19 of 186 |
| Fixture rows created / removed | 49 / 49 — **zero residue** |

### What was NOT tested

**The live browser walkthrough of the 58 frontend routes did not run.** Browser
automation was unavailable in the session that produced this report. No screen was
loaded, so nothing here states whether a page renders, whether its console is clean,
or whether a form submits. Every frontend claim below comes from static analysis of
the deployed JavaScript bundle, and is labelled as such.

To complete it: enable browser tools, then follow `qa/route-audit/ui/walkthrough.md`
and re-run `run.mjs report`.

Generic request-body validation was also not probed. Only 13 of 59 create schemas
declare required fields and the Go binding tags are overwhelmingly `omitempty`, so a
spec-driven validation sweep would mostly assert the absence of validation the API
never promised. Where validation matters it is covered by the workflow suites, and
finding F4 records the consequence of that looseness.

## Frontend analysis (static — no browser)

Method: fetched `https://go-logistics.netlify.app` plus all **205** lazy-loaded chunks,
extracted every `/api/v1/…` and `/auth/…` string, normalised interpolations to path
parameters, and matched against the 186 documented spec paths by shape.

**Result: the frontend calls 167 distinct endpoints, and all 167 exist in the backend.**
There are no calls to undefined endpoints — the most common integration failure mode
is absent here.

Nineteen backend paths have no frontend consumer:

| Path | Note |
|---|---|
| `/api/v1/warranties` | An entire module with no UI |
| `/api/v1/warranties/{id}` | " |
| `/api/v1/purchase-orders/{id}/status-logs` | Transition history is written automatically but never shown |
| `/api/v1/purchase-orders/{id}/status-logs/{child_id}` | " |
| `/api/v1/work-orders/{id}/status-logs/{child_id}` | " |
| `/api/v1/work-orders/{id}/faults` | m2m link, no UI |
| `/api/v1/work-orders/{id}/faults/{fault_id}` | " |
| `/api/v1/assets/{id}/trailer-assignments` | Reverse view is used; the nested mutation route is not |
| `/api/v1/assets/{id}/trailer-assignments/{child_id}` | " |
| `/api/v1/tires/{id}/installations/{child_id}` | Detail route unused |
| `/api/v1/tires/{id}/mount-logs/{child_id}` | " |
| `/api/v1/asset-types`, `/api/v1/asset-types/{id}` | No settings screen |
| `/api/v1/asset-statuses/{id}` | List is used, detail is not |
| `/api/v1/work-order-statuses/{id}` | " |
| `/api/v1/trailer-classifications/{id}` | " |
| `/api/v1/catalog-options/{id}` | " |
| `/api/v1/inventory-journal-entries/{id}` | " |
| `/healthz` | Infrastructure, not a UI concern |

Whether these are gaps or deliberate is a product question. The warranties module and
the two status-log histories are the ones most likely to be unintended: the backend
writes a status-log row on every purchase-order and work-order transition (verified —
12 rows across one PO lifecycle) and no screen ever reads them.

## Frontend route verdicts

Not available — the browser walkthrough did not run. All 58 routes are listed in
`qa/route-audit/ui/routes.json` and remain **unverified**.

## Backend operation verdicts

| Operation | Tested | Failed checks |
|---|---|---|
| `GET /api/v1/admin/companies/{id}/owner` | no — no fixture row satisfies this path | — |
| `GET /api/v1/admin/companies/{id}/roles` | no — no fixture row satisfies this path | list-contract |
| `POST /api/v1/admin/companies/{id}/set-owner` | no — no fixture row satisfies this path | — |
| `GET /api/v1/admin/companies/{id}/work-order-statuses` | no — no fixture row satisfies this path | list-contract |
| `GET /api/v1/admin/employees` | yes | conformance |
| `GET /api/v1/admin/employees/{id}/companies` | yes | — |
| `PUT /api/v1/admin/employees/{id}/companies` | yes | not-found |
| `GET /api/v1/asset-statuses` | yes | — |
| `POST /api/v1/asset-statuses` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/asset-statuses/{id}` | yes | — |
| `PUT /api/v1/asset-statuses/{id}` | yes | not-found |
| `DELETE /api/v1/asset-statuses/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/asset-trailer-assignments` | yes | — |
| `GET /api/v1/asset-types` | yes | — |
| `POST /api/v1/asset-types` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/asset-types/{id}` | yes | — |
| `PUT /api/v1/asset-types/{id}` | yes | not-found |
| `DELETE /api/v1/asset-types/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/assets` | yes | — |
| `POST /api/v1/assets` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/assets/facets` | yes | — |
| `GET /api/v1/assets/{id}` | yes | — |
| `PUT /api/v1/assets/{id}` | yes | — |
| `DELETE /api/v1/assets/{id}` | no — covered by teardown | — |
| `POST /api/v1/assets/{id}/archive` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/assets/{id}/axle-config` | yes | — |
| `PUT /api/v1/assets/{id}/axle-config` | yes | — |
| `DELETE /api/v1/assets/{id}/axle-config` | no — covered by teardown | not-found |
| `GET /api/v1/assets/{id}/fuel-entries` | yes | not-found |
| `POST /api/v1/assets/{id}/fuel-entries` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/assets/{id}/fuel-entries/{child_id}` | yes | — |
| `PUT /api/v1/assets/{id}/fuel-entries/{child_id}` | yes | — |
| `DELETE /api/v1/assets/{id}/fuel-entries/{child_id}` | no — covered by teardown | — |
| `POST /api/v1/assets/{id}/restore` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/assets/{id}/trailer` | yes | — |
| `PUT /api/v1/assets/{id}/trailer` | yes | — |
| `DELETE /api/v1/assets/{id}/trailer` | no — covered by teardown | not-found |
| `GET /api/v1/assets/{id}/trailer-assignments` | yes | not-found |
| `POST /api/v1/assets/{id}/trailer-assignments` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/assets/{id}/trailer-assignments/{child_id}` | yes | — |
| `PUT /api/v1/assets/{id}/trailer-assignments/{child_id}` | yes | — |
| `DELETE /api/v1/assets/{id}/trailer-assignments/{child_id}` | no — covered by teardown | not-found |
| `GET /api/v1/assets/{id}/vehicle` | yes | — |
| `PUT /api/v1/assets/{id}/vehicle` | yes | — |
| `DELETE /api/v1/assets/{id}/vehicle` | no — covered by teardown | not-found |
| `GET /api/v1/axle-definitions/{id}/wheel-positions` | yes | not-found |
| `POST /api/v1/axle-definitions/{id}/wheel-positions` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/axle-definitions/{id}/wheel-positions/{child_id}` | yes | — |
| `PUT /api/v1/axle-definitions/{id}/wheel-positions/{child_id}` | yes | — |
| `DELETE /api/v1/axle-definitions/{id}/wheel-positions/{child_id}` | no — covered by teardown | not-found |
| `GET /api/v1/axle-templates` | yes | — |
| `POST /api/v1/axle-templates` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/axle-templates/{id}` | yes | — |
| `PUT /api/v1/axle-templates/{id}` | yes | — |
| `DELETE /api/v1/axle-templates/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/axle-templates/{id}/definitions` | yes | not-found |
| `POST /api/v1/axle-templates/{id}/definitions` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/axle-templates/{id}/definitions/{child_id}` | yes | — |
| `PUT /api/v1/axle-templates/{id}/definitions/{child_id}` | yes | — |
| `DELETE /api/v1/axle-templates/{id}/definitions/{child_id}` | no — covered by teardown | not-found |
| `GET /api/v1/catalog-options` | yes | — |
| `POST /api/v1/catalog-options` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/catalog-options/{id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/catalog-options/{id}` | no — no fixture row satisfies this path | not-found |
| `DELETE /api/v1/catalog-options/{id}` | no — no fixture row satisfies this path | not-found |
| `GET /api/v1/comments` | yes | — |
| `POST /api/v1/comments` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/comments/{id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/comments/{id}` | no — no fixture row satisfies this path | not-found |
| `DELETE /api/v1/comments/{id}` | no — no fixture row satisfies this path | not-found |
| `GET /api/v1/companies` | yes | — |
| `POST /api/v1/companies` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/companies/{id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/companies/{id}` | no — no fixture row satisfies this path | not-found |
| `DELETE /api/v1/companies/{id}` | no — no fixture row satisfies this path | — |
| `GET /api/v1/custom-field-definitions` | yes | — |
| `POST /api/v1/custom-field-definitions` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/custom-field-definitions/{id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/custom-field-definitions/{id}` | no — no fixture row satisfies this path | not-found |
| `DELETE /api/v1/custom-field-definitions/{id}` | no — no fixture row satisfies this path | not-found |
| `GET /api/v1/dashboard/stats` | yes | — |
| `GET /api/v1/employees` | yes | conformance |
| `POST /api/v1/employees` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/employees/{id}` | yes | conformance |
| `PUT /api/v1/employees/{id}` | yes | conformance |
| `DELETE /api/v1/employees/{id}` | no — covered by teardown | not-found |
| `POST /api/v1/employees/{id}/set-password` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/faults` | yes | — |
| `POST /api/v1/faults` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/faults/{id}` | yes | — |
| `PUT /api/v1/faults/{id}` | yes | — |
| `DELETE /api/v1/faults/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/fuel-entries` | yes | — |
| `GET /api/v1/fuel-entries/{id}/comments` | yes | not-found |
| `POST /api/v1/fuel-entries/{id}/comments` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/fuel-entries/{id}/comments/{child_id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/fuel-entries/{id}/comments/{child_id}` | no — no fixture row satisfies this path | not-found |
| `DELETE /api/v1/fuel-entries/{id}/comments/{child_id}` | no — no fixture row satisfies this path | not-found |
| `GET /api/v1/fuel-entries/{id}/photos` | yes | not-found |
| `POST /api/v1/fuel-entries/{id}/photos` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/fuel-entries/{id}/photos/{child_id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/fuel-entries/{id}/photos/{child_id}` | no — no fixture row satisfies this path | — |
| `DELETE /api/v1/fuel-entries/{id}/photos/{child_id}` | no — no fixture row satisfies this path | — |
| `POST /api/v1/fuel-entries/{id}/photos/{child_id}/set-primary` | no — no fixture row satisfies this path | — |
| `GET /api/v1/fuel-types` | yes | — |
| `POST /api/v1/fuel-types` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/fuel-types/{id}` | yes | — |
| `PUT /api/v1/fuel-types/{id}` | yes | — |
| `DELETE /api/v1/fuel-types/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/groups` | yes | — |
| `POST /api/v1/groups` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/groups/{id}` | yes | — |
| `PUT /api/v1/groups/{id}` | yes | not-found |
| `DELETE /api/v1/groups/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/inspection-forms` | yes | — |
| `POST /api/v1/inspection-forms` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/inspection-forms/{id}` | yes | — |
| `PUT /api/v1/inspection-forms/{id}` | yes | — |
| `DELETE /api/v1/inspection-forms/{id}` | no — covered by teardown | not-found |
| `POST /api/v1/inspection-forms/{id}/archive` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/inspection-forms/{id}/items` | yes | not-found, conformance |
| `POST /api/v1/inspection-forms/{id}/items` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/inspection-forms/{id}/items/{child_id}` | yes | conformance |
| `PUT /api/v1/inspection-forms/{id}/items/{child_id}` | yes | conformance |
| `DELETE /api/v1/inspection-forms/{id}/items/{child_id}` | no — covered by teardown | not-found |
| `POST /api/v1/inspection-forms/{id}/restore` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/inspection-submissions` | yes | — |
| `POST /api/v1/inspection-submissions` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/inspection-submissions/{id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/inspection-submissions/{id}` | no — no fixture row satisfies this path | — |
| `DELETE /api/v1/inspection-submissions/{id}` | no — no fixture row satisfies this path | — |
| `GET /api/v1/inspection-submissions/{id}/items` | no — no fixture row satisfies this path | not-found |
| `POST /api/v1/inspection-submissions/{id}/items` | no — no fixture row satisfies this path | — |
| `GET /api/v1/inspection-submissions/{id}/items/{child_id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/inspection-submissions/{id}/items/{child_id}` | no — no fixture row satisfies this path | — |
| `DELETE /api/v1/inspection-submissions/{id}/items/{child_id}` | no — no fixture row satisfies this path | — |
| `GET /api/v1/inventory-adjustment-reasons` | yes | — |
| `POST /api/v1/inventory-adjustment-reasons` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/inventory-adjustment-reasons/{id}` | yes | — |
| `PUT /api/v1/inventory-adjustment-reasons/{id}` | yes | — |
| `DELETE /api/v1/inventory-adjustment-reasons/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/inventory-journal-entries` | yes | — |
| `POST /api/v1/inventory-journal-entries` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/inventory-journal-entries/{id}` | yes | — |
| `POST /api/v1/inventory-journal-entries/{id}/reverse` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/issue-priorities` | yes | — |
| `POST /api/v1/issue-priorities` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/issue-priorities/{id}` | yes | — |
| `PUT /api/v1/issue-priorities/{id}` | yes | — |
| `DELETE /api/v1/issue-priorities/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/issues` | yes | — |
| `POST /api/v1/issues` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/issues/facets` | yes | — |
| `GET /api/v1/issues/{id}` | yes | — |
| `PUT /api/v1/issues/{id}` | yes | — |
| `DELETE /api/v1/issues/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/issues/{id}/assigned-to` | yes | not-found |
| `POST /api/v1/issues/{id}/assigned-to` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `DELETE /api/v1/issues/{id}/assigned-to/{employee_id}` | no — covered by teardown | not-found |
| `GET /api/v1/issues/{id}/watchers` | yes | not-found |
| `POST /api/v1/issues/{id}/watchers` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `DELETE /api/v1/issues/{id}/watchers/{employee_id}` | no — covered by teardown | not-found |
| `GET /api/v1/locations` | yes | — |
| `POST /api/v1/locations` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/locations/{id}` | yes | — |
| `PUT /api/v1/locations/{id}` | yes | — |
| `DELETE /api/v1/locations/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/me/permissions` | yes | — |
| `GET /api/v1/measurement-units` | yes | — |
| `POST /api/v1/measurement-units` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/measurement-units/{id}` | yes | — |
| `PUT /api/v1/measurement-units/{id}` | yes | — |
| `DELETE /api/v1/measurement-units/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/media` | yes | — |
| `POST /api/v1/media` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/media/{id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/media/{id}` | no — no fixture row satisfies this path | not-found |
| `DELETE /api/v1/media/{id}` | no — no fixture row satisfies this path | — |
| `GET /api/v1/notifications` | yes | list-contract |
| `POST /api/v1/notifications/{id}/read` | no — no fixture row satisfies this path | — |
| `GET /api/v1/part-categories` | yes | — |
| `POST /api/v1/part-categories` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/part-categories/{id}` | yes | — |
| `PUT /api/v1/part-categories/{id}` | yes | — |
| `DELETE /api/v1/part-categories/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/part-inventory` | yes | — |
| `GET /api/v1/part-locations` | yes | — |
| `POST /api/v1/part-locations` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/part-locations/{id}` | yes | — |
| `PUT /api/v1/part-locations/{id}` | yes | — |
| `DELETE /api/v1/part-locations/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/part-manufacturers` | yes | — |
| `POST /api/v1/part-manufacturers` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/part-manufacturers/{id}` | yes | — |
| `PUT /api/v1/part-manufacturers/{id}` | yes | — |
| `DELETE /api/v1/part-manufacturers/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/parts` | yes | — |
| `POST /api/v1/parts` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/parts/{id}` | yes | — |
| `PUT /api/v1/parts/{id}` | yes | — |
| `DELETE /api/v1/parts/{id}` | no — covered by teardown | not-found |
| `POST /api/v1/parts/{id}/archive` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/parts/{id}/inventory` | yes | not-found |
| `POST /api/v1/parts/{id}/inventory` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/parts/{id}/inventory/{child_id}` | yes | — |
| `PUT /api/v1/parts/{id}/inventory/{child_id}` | yes | — |
| `DELETE /api/v1/parts/{id}/inventory/{child_id}` | no — covered by teardown | not-found |
| `POST /api/v1/parts/{id}/restore` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/purchase-order-line-items` | yes | — |
| `GET /api/v1/purchase-orders` | yes | — |
| `POST /api/v1/purchase-orders` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/purchase-orders/{id}` | yes | — |
| `PUT /api/v1/purchase-orders/{id}` | yes | — |
| `DELETE /api/v1/purchase-orders/{id}` | no — covered by teardown | not-found |
| `POST /api/v1/purchase-orders/{id}/approve` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `POST /api/v1/purchase-orders/{id}/close` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/purchase-orders/{id}/line-items` | yes | not-found |
| `POST /api/v1/purchase-orders/{id}/line-items` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/purchase-orders/{id}/line-items/{child_id}` | yes | — |
| `PUT /api/v1/purchase-orders/{id}/line-items/{child_id}` | yes | — |
| `DELETE /api/v1/purchase-orders/{id}/line-items/{child_id}` | no — covered by teardown | — |
| `POST /api/v1/purchase-orders/{id}/override-total` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `POST /api/v1/purchase-orders/{id}/purchase` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `POST /api/v1/purchase-orders/{id}/receive-full` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `POST /api/v1/purchase-orders/{id}/receive-partial` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `POST /api/v1/purchase-orders/{id}/reject` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `POST /api/v1/purchase-orders/{id}/revise` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/purchase-orders/{id}/status-logs` | yes | not-found |
| `GET /api/v1/purchase-orders/{id}/status-logs/{child_id}` | no — no fixture row satisfies this path | — |
| `POST /api/v1/purchase-orders/{id}/submit` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/roles` | yes | conformance |
| `POST /api/v1/roles` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/roles/{id}` | yes | conformance |
| `PUT /api/v1/roles/{id}` | yes | conformance |
| `DELETE /api/v1/roles/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/service-entries` | yes | — |
| `POST /api/v1/service-entries` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/service-entries/{id}` | yes | — |
| `PUT /api/v1/service-entries/{id}` | yes | — |
| `DELETE /api/v1/service-entries/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/service-entries/{id}/line-items` | yes | not-found |
| `POST /api/v1/service-entries/{id}/line-items` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/service-entries/{id}/line-items/{child_id}` | yes | — |
| `PUT /api/v1/service-entries/{id}/line-items/{child_id}` | yes | — |
| `DELETE /api/v1/service-entries/{id}/line-items/{child_id}` | no — covered by teardown | — |
| `POST /api/v1/service-entries/{id}/override-total` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/service-entry-line-items/{id}/issues` | yes | not-found |
| `POST /api/v1/service-entry-line-items/{id}/issues` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `DELETE /api/v1/service-entry-line-items/{id}/issues/{issue_id}` | no — covered by teardown | not-found |
| `GET /api/v1/service-reminders` | yes | — |
| `POST /api/v1/service-reminders` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/service-reminders/{id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/service-reminders/{id}` | no — no fixture row satisfies this path | — |
| `DELETE /api/v1/service-reminders/{id}` | no — no fixture row satisfies this path | not-found |
| `GET /api/v1/service-tasks` | yes | — |
| `POST /api/v1/service-tasks` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/service-tasks/{id}` | yes | — |
| `PUT /api/v1/service-tasks/{id}` | yes | — |
| `DELETE /api/v1/service-tasks/{id}` | no — covered by teardown | not-found |
| `POST /api/v1/service-tasks/{id}/archive` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/service-tasks/{id}/parts` | yes | not-found |
| `POST /api/v1/service-tasks/{id}/parts` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/service-tasks/{id}/parts/{child_id}` | yes | — |
| `PUT /api/v1/service-tasks/{id}/parts/{child_id}` | yes | — |
| `DELETE /api/v1/service-tasks/{id}/parts/{child_id}` | no — covered by teardown | not-found |
| `POST /api/v1/service-tasks/{id}/restore` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/tire-assignment-requests` | yes | — |
| `POST /api/v1/tire-assignment-requests` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/tire-assignment-requests/{id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/tire-assignment-requests/{id}` | no — no fixture row satisfies this path | — |
| `DELETE /api/v1/tire-assignment-requests/{id}` | no — no fixture row satisfies this path | not-found |
| `POST /api/v1/tire-assignment-requests/{id}/approve` | no — no fixture row satisfies this path | — |
| `GET /api/v1/tire-models` | yes | — |
| `POST /api/v1/tire-models` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/tire-models/{id}` | yes | — |
| `PUT /api/v1/tire-models/{id}` | yes | — |
| `DELETE /api/v1/tire-models/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/tire-mount-logs` | yes | — |
| `GET /api/v1/tires` | yes | — |
| `POST /api/v1/tires` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/tires/{id}` | yes | — |
| `PUT /api/v1/tires/{id}` | yes | — |
| `DELETE /api/v1/tires/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/tires/{id}/inspections` | yes | not-found |
| `POST /api/v1/tires/{id}/inspections` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/tires/{id}/inspections/{child_id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/tires/{id}/inspections/{child_id}` | no — no fixture row satisfies this path | — |
| `DELETE /api/v1/tires/{id}/inspections/{child_id}` | no — no fixture row satisfies this path | not-found |
| `GET /api/v1/tires/{id}/installations` | yes | not-found |
| `POST /api/v1/tires/{id}/installations` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/tires/{id}/installations/{child_id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/tires/{id}/installations/{child_id}` | no — no fixture row satisfies this path | — |
| `DELETE /api/v1/tires/{id}/installations/{child_id}` | no — no fixture row satisfies this path | not-found |
| `GET /api/v1/tires/{id}/mount-logs` | yes | not-found |
| `POST /api/v1/tires/{id}/mount-logs` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/tires/{id}/mount-logs/{child_id}` | no — no fixture row satisfies this path | — |
| `PUT /api/v1/tires/{id}/mount-logs/{child_id}` | no — no fixture row satisfies this path | — |
| `DELETE /api/v1/tires/{id}/mount-logs/{child_id}` | no — no fixture row satisfies this path | not-found |
| `GET /api/v1/trailer-classifications` | yes | — |
| `POST /api/v1/trailer-classifications` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/trailer-classifications/{id}` | yes | — |
| `PUT /api/v1/trailer-classifications/{id}` | yes | not-found |
| `DELETE /api/v1/trailer-classifications/{id}` | no — covered by teardown | not-found |
| `POST /api/v1/uploads` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/vehicle-makes` | yes | — |
| `POST /api/v1/vehicle-makes` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/vehicle-makes/{id}` | yes | — |
| `PUT /api/v1/vehicle-makes/{id}` | yes | not-found |
| `DELETE /api/v1/vehicle-makes/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/vehicle-models` | yes | — |
| `POST /api/v1/vehicle-models` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/vehicle-models/{id}` | yes | — |
| `PUT /api/v1/vehicle-models/{id}` | yes | not-found |
| `DELETE /api/v1/vehicle-models/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/vendors` | yes | — |
| `POST /api/v1/vendors` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/vendors/{id}` | yes | — |
| `PUT /api/v1/vendors/{id}` | yes | — |
| `DELETE /api/v1/vendors/{id}` | no — covered by teardown | not-found |
| `POST /api/v1/vendors/{id}/archive` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `POST /api/v1/vendors/{id}/restore` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/warranties` | yes | — |
| `POST /api/v1/warranties` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/warranties/{id}` | yes | — |
| `PUT /api/v1/warranties/{id}` | yes | not-found |
| `DELETE /api/v1/warranties/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/weekly-mileage-goals` | yes | — |
| `POST /api/v1/weekly-mileage-goals` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/weekly-mileage-goals/{id}` | yes | — |
| `PUT /api/v1/weekly-mileage-goals/{id}` | yes | — |
| `DELETE /api/v1/weekly-mileage-goals/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/work-order-line-items/{id}/issues` | yes | not-found |
| `POST /api/v1/work-order-line-items/{id}/issues` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `DELETE /api/v1/work-order-line-items/{id}/issues/{issue_id}` | no — covered by teardown | not-found |
| `GET /api/v1/work-order-line-items/{id}/sub-line-items` | yes | not-found |
| `POST /api/v1/work-order-line-items/{id}/sub-line-items` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/work-order-line-items/{id}/sub-line-items/{child_id}` | yes | — |
| `PUT /api/v1/work-order-line-items/{id}/sub-line-items/{child_id}` | yes | — |
| `DELETE /api/v1/work-order-line-items/{id}/sub-line-items/{child_id}` | no — covered by teardown | — |
| `GET /api/v1/work-order-statuses` | yes | — |
| `POST /api/v1/work-order-statuses` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/work-order-statuses/{id}` | yes | — |
| `PUT /api/v1/work-order-statuses/{id}` | yes | — |
| `DELETE /api/v1/work-order-statuses/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/work-order-sub-line-items/{id}/labor-entries` | yes | not-found |
| `POST /api/v1/work-order-sub-line-items/{id}/labor-entries` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id}` | yes | — |
| `PUT /api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id}` | yes | — |
| `DELETE /api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id}` | no — covered by teardown | not-found |
| `GET /api/v1/work-orders` | yes | — |
| `POST /api/v1/work-orders` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/work-orders/facets` | yes | — |
| `GET /api/v1/work-orders/{id}` | yes | — |
| `PUT /api/v1/work-orders/{id}` | yes | — |
| `DELETE /api/v1/work-orders/{id}` | no — covered by teardown | not-found |
| `GET /api/v1/work-orders/{id}/faults` | yes | not-found |
| `POST /api/v1/work-orders/{id}/faults` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `DELETE /api/v1/work-orders/{id}/faults/{fault_id}` | no — covered by teardown | not-found |
| `GET /api/v1/work-orders/{id}/issues` | yes | not-found |
| `POST /api/v1/work-orders/{id}/issues` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `DELETE /api/v1/work-orders/{id}/issues/{issue_id}` | no — covered by teardown | not-found |
| `GET /api/v1/work-orders/{id}/line-items` | yes | not-found |
| `POST /api/v1/work-orders/{id}/line-items` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/work-orders/{id}/line-items/{child_id}` | yes | — |
| `PUT /api/v1/work-orders/{id}/line-items/{child_id}` | yes | — |
| `DELETE /api/v1/work-orders/{id}/line-items/{child_id}` | no — covered by teardown | — |
| `POST /api/v1/work-orders/{id}/override-total` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /api/v1/work-orders/{id}/status-logs` | yes | not-found |
| `GET /api/v1/work-orders/{id}/status-logs/{child_id}` | no — no fixture row satisfies this path | — |
| `POST /auth/login` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `POST /auth/logout` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `POST /auth/refresh` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `POST /auth/switch-company` | no — collection POST is proved by the fixture build; action POST is owned by the workflow suites | — |
| `GET /healthz` | yes | — |

## Findings

Every finding below was re-verified against production with a freshly issued token
after the run completed. Each carries a runnable reproduction. Backend fixes cite
`file:line`; frontend findings describe required behaviour, because the frontend
source was not available.

Severity reflects consequence, not effort: nothing here is a security defect, and the
audit found no way for one tenant to read another's data or for an unprivileged role
to reach a gated module.

---

### F1 — 50 DELETE endpoints return `204` for a row that does not exist (medium)

**Expected** `404`, per the OpenAPI spec. **Observed** `204 No Content`.

```bash
curl -s -o /dev/null -w '%{http_code}\n' -X DELETE \
  -H "Authorization: Bearer $TOKEN" \
  https://go-logistics.jesuslab135.com/api/v1/asset-types/999999999
# 204
```

A client cannot distinguish "I deleted it" from "it was never there". This matters for
any caller doing reconciliation or retry logic: a delete that silently succeeds against
a missing row hides a broken reference rather than surfacing it.

The API *can* detect absence — the same id returns `404` on `GET`:

```bash
curl -s -o /dev/null -w '%{http_code}\n' \
  -H "Authorization: Bearer $TOKEN" \
  https://go-logistics.jesuslab135.com/api/v1/asset-types/999999999
# 404
```

So this is not a limitation, it is an inconsistency: `GET` checks existence first,
`DELETE` does not.

**Fix — pick one and make it uniform.** Idempotent `DELETE` returning `204` for a
missing row is a defensible REST choice; if that is the intent, the spec is wrong and
should document `204`. If `404` is the intent, the generic delete handler needs an
existence check. The generic path is `internal/platform/crud/` — `Handler.Delete` — and
the change belongs there rather than in 50 individual stores.

Affected: every `DELETE /api/v1/<collection>/{id}` operation. Full list in the
"Failed contract checks" table below, filtered on `check = not-found`, `actual = 204`.

---

### F2 — 26 nested list endpoints return an empty page when the PARENT does not exist (medium)

**Expected** `404`. **Observed** `200` with an empty page.

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  'https://go-logistics.jesuslab135.com/api/v1/assets/999999999/fuel-entries'
# {"data":[],"total":0,"limit":25,"offset":0,"has_next":false}
```

A client asking for a nonexistent asset's fuel entries is told "this asset has no fuel
entries" rather than "there is no such asset". A typo in an id, or a stale reference
after a delete, renders as a legitimately empty screen. This is the finding most likely
to produce a confusing bug report from a real user: the UI will show an empty list and
nothing anywhere indicates the parent is missing.

**Fix:** the nested list handlers should verify the parent exists and scope to the
caller's company before returning a page. The nested read path is registered through
`crud.NewNestedHandler` (`internal/http/handler/router.go`, Phase 8 block) and the
parent check belongs in the shared nested handler, not per resource.

---

### F3 — 14 PUT endpoints validate the body before checking the row exists (low/medium)

**Expected** `404`. **Observed** `422 validation_failed`.

```bash
curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{}' \
  https://go-logistics.jesuslab135.com/api/v1/asset-types/999999999
# {"error":{"code":"validation_failed","details":{"name":"this field is required"}}}
```

Ordering, not a defect in itself — but it means a caller cannot learn that a row is
missing without first constructing a payload valid enough to pass validation. Combined
with F1 it makes "does this row exist?" answerable only via `GET`.

**Fix:** if uniformity with `GET` is wanted, check existence before binding. Lower
priority than F1 and F2.

---

### F4 — The API silently accepts and discards unknown JSON fields (medium — highest practical impact)

**Observed:** a create carrying a field the DTO does not declare returns `201`, and the
field is silently dropped.

```bash
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"ZZ-TEST-demo","totally_not_a_field":"xyz"}' \
  https://go-logistics.jesuslab135.com/api/v1/vendors
# 201 — and the response does not contain totally_not_a_field
```

This is Gin's default binding behaviour, so it applies to every write endpoint.

**Why this is the most consequential finding in the report despite not being a
security issue:** it makes integration errors invisible. During construction of this
audit, **9 of 45 fixture payloads** hit it — written from the OpenAPI spec, which does
not reflect the Go structs. Two were particularly instructive:

- `part` sent `name`, `category_id` and `manufacturer_id`. The real fields are
  `part_number`, `description`, `part_category_id`, `part_manufacturer_id`. The row was
  created with a valid `part_number`, so it *looked* healthy — while **both foreign keys
  were silently dropped**.
- `workOrderLineItem` sent `service_task_id`, which is not a foreign key in that DTO.
  The row was created, looked correct, and had no service-task link.

Neither produced any error. Both were found only by reading the Go structs field by
field. A third-party client integrating against this API would make the same mistakes
and discover them months later as missing data.

**Fix:** enable strict field rejection on request binding, so an unknown field returns
`400` naming it. In Gin this means decoding with `DisallowUnknownFields`. The shared
bind path is `internal/platform/reqbind/`. This is a breaking change for any existing
client currently sending extra fields, so it deserves a deliberate rollout — but the
current behaviour converts client bugs into silent data loss.

A cheaper partial mitigation: regenerate the OpenAPI spec so it faithfully describes the
Go structs. Much of the harm here came from the spec and the implementation disagreeing.

---

### F5 — `409 foreign_key_violation` does not name the offending field (low)

```json
{"error":{"code":"foreign_key_violation","message":"referenced resource constraint violated"}}
```

Compare a `422` from the same API:

```json
{"error":{"code":"validation_failed","details":{"provider_id":"must name a vendor of your company"}}}
```

The `422` responses are diagnosable from the response alone. The `409`s required
reading Go source to determine which of several foreign keys was at fault — during this
audit that cost several debugging cycles on `issue`, `purchaseOrder`, `fuelEntry` and
`laborEntry`, each of which carries multiple FKs.

**Fix:** include the constraint or column name in the error details, as the validation
path already does. The Postgres error carries the constraint name; the mapping lives in
`internal/platform/apierr/`.

---

### F6 — The append-only ledger guarantee is bypassable via cascade (low)

The HTTP layer enforces immutability:

```bash
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" \
  https://go-logistics.jesuslab135.com/api/v1/inventory-journal-entries/1
# 405 "the inventory ledger is append-only: correct an entry with POST .../reverse"
```

That is good behaviour, and the message correctly names its replacement. But the
database undoes it:

```
internal/db/migrations/000001_init.up.sql:1281
ALTER TABLE inventory_journal_entry
  ADD CONSTRAINT fk_inventory_journal_entry_part_location_detail
  FOREIGN KEY (part_location_detail_id) REFERENCES part_inventory(id)
  ON DELETE CASCADE;
```

Deleting a `part_inventory` row silently removes every ledger entry referencing it,
including reversals. An immutability guarantee asserted at the API layer is defeated by
an ordinary delete one level up. Observed directly: this audit's teardown relied on it.

**Fix:** if the ledger is genuinely append-only, that FK should be `RESTRICT`, forcing
callers to reverse entries before retiring a stock location. If cascade is intended,
the 405's promise is overstated.

---

### F7 — A required-looking field is satisfied by omitting it (low)

`CreateAssetTrailerAssignmentRequest.AssignedDate` is a non-pointer `time.Time`, which
reads as required in the Go source. A create omitting it succeeds and stores the zero
value. Same family as F4: the struct's apparent contract is not the enforced one.

**Fix:** make it a pointer if optional, or add `binding:"required"` if not.

---

### F8 — Inconsistent path-parameter naming across link routes (low, spec-only)

Many-to-many link routes use named parameters — `/issues/{id}/assigned-to/{employee_id}`,
`/work-orders/{id}/faults/{fault_id}` — while ordinary nested child routes use
`{child_id}`. A generated client produces inconsistent parameter names for
structurally identical routes. Cosmetic, but it is the kind of thing that makes a
generated SDK feel unreliable.

---

### F9 — `/tire-assignment-requests` is readable regardless of module grants (informational)

A role granting no modules can list tire assignment requests. This is **intentional** —
`internal/http/handler/router.go:177-185` registers the list on `member` (company
membership only) for Django parity, and puts the `tire_approvals` gate on the approve
action, which correctly returned `403` to the same role.

Recorded because a reader auditing module gating would flag it, and the reasoning is
worth having written down. Not a defect.

---

### Failed contract checks

| Operation | Check | Expected | Observed | Severity |
|---|---|---|---|---|
| `GET /api/v1/admin/companies/{id}/roles` | list-contract | 200 with a page envelope | 404 | medium |
| `GET /api/v1/admin/companies/{id}/work-order-statuses` | list-contract | 200 with a page envelope | 404 | medium |
| `GET /api/v1/admin/employees` | conformance | response matches the documented schema | type: data[0].dashboard_preferences expected array got object; data[0].table_preferences expected array got object; data[1].dashboard_preferences expected array | low |
| `PUT /api/v1/admin/employees/{id}/companies` | not-found | 404 | 422 | medium |
| `PUT /api/v1/asset-statuses/{id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/asset-statuses/{id}` | not-found | 404 | 204 | medium |
| `PUT /api/v1/asset-types/{id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/asset-types/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/assets/{id}/axle-config` | not-found | 404 | 204 | medium |
| `GET /api/v1/assets/{id}/fuel-entries` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/assets/{id}/trailer` | not-found | 404 | 204 | medium |
| `GET /api/v1/assets/{id}/trailer-assignments` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/assets/{id}/trailer-assignments/{child_id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/assets/{id}/vehicle` | not-found | 404 | 204 | medium |
| `GET /api/v1/axle-definitions/{id}/wheel-positions` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/axle-definitions/{id}/wheel-positions/{child_id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/axle-templates/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/axle-templates/{id}/definitions` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/axle-templates/{id}/definitions/{child_id}` | not-found | 404 | 204 | medium |
| `PUT /api/v1/catalog-options/{id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/catalog-options/{id}` | not-found | 404 | 204 | medium |
| `PUT /api/v1/comments/{id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/comments/{id}` | not-found | 404 | 204 | medium |
| `PUT /api/v1/companies/{id}` | not-found | 404 | 422 | medium |
| `PUT /api/v1/custom-field-definitions/{id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/custom-field-definitions/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/employees` | conformance | response matches the documented schema | type: data[0].dashboard_preferences expected array got object; data[0].table_preferences expected array got object; data[1].dashboard_preferences expected array | low |
| `GET /api/v1/employees/{id}` | conformance | response matches the documented schema | type: dashboard_preferences expected array got object; table_preferences expected array got object | low |
| `PUT /api/v1/employees/{id}` | conformance | response matches the documented schema | type: dashboard_preferences expected array got object; table_preferences expected array got object | low |
| `DELETE /api/v1/employees/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/faults/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/fuel-entries/{id}/comments` | not-found | 404 | 200 | medium |
| `PUT /api/v1/fuel-entries/{id}/comments/{child_id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/fuel-entries/{id}/comments/{child_id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/fuel-entries/{id}/photos` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/fuel-types/{id}` | not-found | 404 | 204 | medium |
| `PUT /api/v1/groups/{id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/groups/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/inspection-forms/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/inspection-forms/{id}/items` | not-found | 404 | 200 | medium |
| `GET /api/v1/inspection-forms/{id}/items` | conformance | response matches the documented schema | type: data[0].type_config expected array got object | low |
| `GET /api/v1/inspection-forms/{id}/items/{child_id}` | conformance | response matches the documented schema | type: type_config expected array got object | low |
| `PUT /api/v1/inspection-forms/{id}/items/{child_id}` | conformance | response matches the documented schema | type: type_config expected array got object | low |
| `DELETE /api/v1/inspection-forms/{id}/items/{child_id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/inspection-submissions/{id}/items` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/inventory-adjustment-reasons/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/issue-priorities/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/issues/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/issues/{id}/assigned-to` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/issues/{id}/assigned-to/{employee_id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/issues/{id}/watchers` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/issues/{id}/watchers/{employee_id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/locations/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/measurement-units/{id}` | not-found | 404 | 204 | medium |
| `PUT /api/v1/media/{id}` | not-found | 404 | 422 | medium |
| `GET /api/v1/notifications` | list-contract | limit honoured, envelope complete | total is not a number; has_next is not a boolean | medium |
| `DELETE /api/v1/part-categories/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/part-locations/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/part-manufacturers/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/parts/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/parts/{id}/inventory` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/parts/{id}/inventory/{child_id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/purchase-orders/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/purchase-orders/{id}/line-items` | not-found | 404 | 200 | medium |
| `GET /api/v1/purchase-orders/{id}/status-logs` | not-found | 404 | 200 | medium |
| `GET /api/v1/roles` | conformance | response matches the documented schema | type: data[0].permissions expected array got object; data[1].permissions expected array got object; data[2].permissions expected array got object | low |
| `GET /api/v1/roles/{id}` | conformance | response matches the documented schema | type: permissions expected array got object | low |
| `PUT /api/v1/roles/{id}` | conformance | response matches the documented schema | type: permissions expected array got object | low |
| `DELETE /api/v1/roles/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/service-entries/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/service-entries/{id}/line-items` | not-found | 404 | 200 | medium |
| `GET /api/v1/service-entry-line-items/{id}/issues` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/service-entry-line-items/{id}/issues/{issue_id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/service-reminders/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/service-tasks/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/service-tasks/{id}/parts` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/service-tasks/{id}/parts/{child_id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/tire-assignment-requests/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/tire-models/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/tires/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/tires/{id}/inspections` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/tires/{id}/inspections/{child_id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/tires/{id}/installations` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/tires/{id}/installations/{child_id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/tires/{id}/mount-logs` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/tires/{id}/mount-logs/{child_id}` | not-found | 404 | 204 | medium |
| `PUT /api/v1/trailer-classifications/{id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/trailer-classifications/{id}` | not-found | 404 | 204 | medium |
| `PUT /api/v1/vehicle-makes/{id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/vehicle-makes/{id}` | not-found | 404 | 204 | medium |
| `PUT /api/v1/vehicle-models/{id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/vehicle-models/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/vendors/{id}` | not-found | 404 | 204 | medium |
| `PUT /api/v1/warranties/{id}` | not-found | 404 | 422 | medium |
| `DELETE /api/v1/warranties/{id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/weekly-mileage-goals/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/work-order-line-items/{id}/issues` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/work-order-line-items/{id}/issues/{issue_id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/work-order-line-items/{id}/sub-line-items` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/work-order-statuses/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/work-order-sub-line-items/{id}/labor-entries` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id}` | not-found | 404 | 204 | medium |
| `DELETE /api/v1/work-orders/{id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/work-orders/{id}/faults` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/work-orders/{id}/faults/{fault_id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/work-orders/{id}/issues` | not-found | 404 | 200 | medium |
| `DELETE /api/v1/work-orders/{id}/issues/{issue_id}` | not-found | 404 | 204 | medium |
| `GET /api/v1/work-orders/{id}/line-items` | not-found | 404 | 200 | medium |
| `GET /api/v1/work-orders/{id}/status-logs` | not-found | 404 | 200 | medium |

## Untested and blocked

Operations that carry no verdict, and why. This section exists so the
report admits its gaps rather than implying coverage it does not have.

| Operation | Reason |
|---|---|
| `GET /api/v1/admin/companies/{id}/owner` | no fixture row satisfies this path |
| `GET /api/v1/admin/companies/{id}/roles` | no fixture row satisfies this path |
| `POST /api/v1/admin/companies/{id}/set-owner` | no fixture row satisfies this path |
| `GET /api/v1/admin/companies/{id}/work-order-statuses` | no fixture row satisfies this path |
| `POST /api/v1/asset-statuses` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/asset-statuses/{id}` | covered by teardown |
| `POST /api/v1/asset-types` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/asset-types/{id}` | covered by teardown |
| `POST /api/v1/assets` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/assets/{id}` | covered by teardown |
| `POST /api/v1/assets/{id}/archive` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/assets/{id}/axle-config` | covered by teardown |
| `POST /api/v1/assets/{id}/fuel-entries` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/assets/{id}/fuel-entries/{child_id}` | covered by teardown |
| `POST /api/v1/assets/{id}/restore` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/assets/{id}/trailer` | covered by teardown |
| `POST /api/v1/assets/{id}/trailer-assignments` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/assets/{id}/trailer-assignments/{child_id}` | covered by teardown |
| `DELETE /api/v1/assets/{id}/vehicle` | covered by teardown |
| `POST /api/v1/axle-definitions/{id}/wheel-positions` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/axle-definitions/{id}/wheel-positions/{child_id}` | covered by teardown |
| `POST /api/v1/axle-templates` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/axle-templates/{id}` | covered by teardown |
| `POST /api/v1/axle-templates/{id}/definitions` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/axle-templates/{id}/definitions/{child_id}` | covered by teardown |
| `POST /api/v1/catalog-options` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/catalog-options/{id}` | no fixture row satisfies this path |
| `PUT /api/v1/catalog-options/{id}` | no fixture row satisfies this path |
| `DELETE /api/v1/catalog-options/{id}` | no fixture row satisfies this path |
| `POST /api/v1/comments` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/comments/{id}` | no fixture row satisfies this path |
| `PUT /api/v1/comments/{id}` | no fixture row satisfies this path |
| `DELETE /api/v1/comments/{id}` | no fixture row satisfies this path |
| `POST /api/v1/companies` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/companies/{id}` | no fixture row satisfies this path |
| `PUT /api/v1/companies/{id}` | no fixture row satisfies this path |
| `DELETE /api/v1/companies/{id}` | no fixture row satisfies this path |
| `POST /api/v1/custom-field-definitions` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/custom-field-definitions/{id}` | no fixture row satisfies this path |
| `PUT /api/v1/custom-field-definitions/{id}` | no fixture row satisfies this path |
| `DELETE /api/v1/custom-field-definitions/{id}` | no fixture row satisfies this path |
| `POST /api/v1/employees` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/employees/{id}` | covered by teardown |
| `POST /api/v1/employees/{id}/set-password` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/faults` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/faults/{id}` | covered by teardown |
| `POST /api/v1/fuel-entries/{id}/comments` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/fuel-entries/{id}/comments/{child_id}` | no fixture row satisfies this path |
| `PUT /api/v1/fuel-entries/{id}/comments/{child_id}` | no fixture row satisfies this path |
| `DELETE /api/v1/fuel-entries/{id}/comments/{child_id}` | no fixture row satisfies this path |
| `POST /api/v1/fuel-entries/{id}/photos` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/fuel-entries/{id}/photos/{child_id}` | no fixture row satisfies this path |
| `PUT /api/v1/fuel-entries/{id}/photos/{child_id}` | no fixture row satisfies this path |
| `DELETE /api/v1/fuel-entries/{id}/photos/{child_id}` | no fixture row satisfies this path |
| `POST /api/v1/fuel-entries/{id}/photos/{child_id}/set-primary` | no fixture row satisfies this path |
| `POST /api/v1/fuel-types` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/fuel-types/{id}` | covered by teardown |
| `POST /api/v1/groups` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/groups/{id}` | covered by teardown |
| `POST /api/v1/inspection-forms` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/inspection-forms/{id}` | covered by teardown |
| `POST /api/v1/inspection-forms/{id}/archive` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/inspection-forms/{id}/items` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/inspection-forms/{id}/items/{child_id}` | covered by teardown |
| `POST /api/v1/inspection-forms/{id}/restore` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/inspection-submissions` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/inspection-submissions/{id}` | no fixture row satisfies this path |
| `PUT /api/v1/inspection-submissions/{id}` | no fixture row satisfies this path |
| `DELETE /api/v1/inspection-submissions/{id}` | no fixture row satisfies this path |
| `GET /api/v1/inspection-submissions/{id}/items` | no fixture row satisfies this path |
| `POST /api/v1/inspection-submissions/{id}/items` | no fixture row satisfies this path |
| `GET /api/v1/inspection-submissions/{id}/items/{child_id}` | no fixture row satisfies this path |
| `PUT /api/v1/inspection-submissions/{id}/items/{child_id}` | no fixture row satisfies this path |
| `DELETE /api/v1/inspection-submissions/{id}/items/{child_id}` | no fixture row satisfies this path |
| `POST /api/v1/inventory-adjustment-reasons` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/inventory-adjustment-reasons/{id}` | covered by teardown |
| `POST /api/v1/inventory-journal-entries` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/inventory-journal-entries/{id}/reverse` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/issue-priorities` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/issue-priorities/{id}` | covered by teardown |
| `POST /api/v1/issues` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/issues/{id}` | covered by teardown |
| `POST /api/v1/issues/{id}/assigned-to` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/issues/{id}/assigned-to/{employee_id}` | covered by teardown |
| `POST /api/v1/issues/{id}/watchers` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/issues/{id}/watchers/{employee_id}` | covered by teardown |
| `POST /api/v1/locations` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/locations/{id}` | covered by teardown |
| `POST /api/v1/measurement-units` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/measurement-units/{id}` | covered by teardown |
| `POST /api/v1/media` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/media/{id}` | no fixture row satisfies this path |
| `PUT /api/v1/media/{id}` | no fixture row satisfies this path |
| `DELETE /api/v1/media/{id}` | no fixture row satisfies this path |
| `POST /api/v1/notifications/{id}/read` | no fixture row satisfies this path |
| `POST /api/v1/part-categories` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/part-categories/{id}` | covered by teardown |
| `POST /api/v1/part-locations` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/part-locations/{id}` | covered by teardown |
| `POST /api/v1/part-manufacturers` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/part-manufacturers/{id}` | covered by teardown |
| `POST /api/v1/parts` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/parts/{id}` | covered by teardown |
| `POST /api/v1/parts/{id}/archive` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/parts/{id}/inventory` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/parts/{id}/inventory/{child_id}` | covered by teardown |
| `POST /api/v1/parts/{id}/restore` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/purchase-orders` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/purchase-orders/{id}` | covered by teardown |
| `POST /api/v1/purchase-orders/{id}/approve` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/purchase-orders/{id}/close` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/purchase-orders/{id}/line-items` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/purchase-orders/{id}/line-items/{child_id}` | covered by teardown |
| `POST /api/v1/purchase-orders/{id}/override-total` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/purchase-orders/{id}/purchase` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/purchase-orders/{id}/receive-full` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/purchase-orders/{id}/receive-partial` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/purchase-orders/{id}/reject` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/purchase-orders/{id}/revise` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/purchase-orders/{id}/status-logs/{child_id}` | no fixture row satisfies this path |
| `POST /api/v1/purchase-orders/{id}/submit` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/roles` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/roles/{id}` | covered by teardown |
| `POST /api/v1/service-entries` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/service-entries/{id}` | covered by teardown |
| `POST /api/v1/service-entries/{id}/line-items` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/service-entries/{id}/line-items/{child_id}` | covered by teardown |
| `POST /api/v1/service-entries/{id}/override-total` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/service-entry-line-items/{id}/issues` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/service-entry-line-items/{id}/issues/{issue_id}` | covered by teardown |
| `POST /api/v1/service-reminders` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/service-reminders/{id}` | no fixture row satisfies this path |
| `PUT /api/v1/service-reminders/{id}` | no fixture row satisfies this path |
| `DELETE /api/v1/service-reminders/{id}` | no fixture row satisfies this path |
| `POST /api/v1/service-tasks` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/service-tasks/{id}` | covered by teardown |
| `POST /api/v1/service-tasks/{id}/archive` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/service-tasks/{id}/parts` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/service-tasks/{id}/parts/{child_id}` | covered by teardown |
| `POST /api/v1/service-tasks/{id}/restore` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/tire-assignment-requests` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/tire-assignment-requests/{id}` | no fixture row satisfies this path |
| `PUT /api/v1/tire-assignment-requests/{id}` | no fixture row satisfies this path |
| `DELETE /api/v1/tire-assignment-requests/{id}` | no fixture row satisfies this path |
| `POST /api/v1/tire-assignment-requests/{id}/approve` | no fixture row satisfies this path |
| `POST /api/v1/tire-models` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/tire-models/{id}` | covered by teardown |
| `POST /api/v1/tires` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/tires/{id}` | covered by teardown |
| `POST /api/v1/tires/{id}/inspections` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/tires/{id}/inspections/{child_id}` | no fixture row satisfies this path |
| `PUT /api/v1/tires/{id}/inspections/{child_id}` | no fixture row satisfies this path |
| `DELETE /api/v1/tires/{id}/inspections/{child_id}` | no fixture row satisfies this path |
| `POST /api/v1/tires/{id}/installations` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/tires/{id}/installations/{child_id}` | no fixture row satisfies this path |
| `PUT /api/v1/tires/{id}/installations/{child_id}` | no fixture row satisfies this path |
| `DELETE /api/v1/tires/{id}/installations/{child_id}` | no fixture row satisfies this path |
| `POST /api/v1/tires/{id}/mount-logs` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/tires/{id}/mount-logs/{child_id}` | no fixture row satisfies this path |
| `PUT /api/v1/tires/{id}/mount-logs/{child_id}` | no fixture row satisfies this path |
| `DELETE /api/v1/tires/{id}/mount-logs/{child_id}` | no fixture row satisfies this path |
| `POST /api/v1/trailer-classifications` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/trailer-classifications/{id}` | covered by teardown |
| `POST /api/v1/uploads` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/vehicle-makes` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/vehicle-makes/{id}` | covered by teardown |
| `POST /api/v1/vehicle-models` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/vehicle-models/{id}` | covered by teardown |
| `POST /api/v1/vendors` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/vendors/{id}` | covered by teardown |
| `POST /api/v1/vendors/{id}/archive` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/vendors/{id}/restore` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /api/v1/warranties` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/warranties/{id}` | covered by teardown |
| `POST /api/v1/weekly-mileage-goals` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/weekly-mileage-goals/{id}` | covered by teardown |
| `POST /api/v1/work-order-line-items/{id}/issues` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/work-order-line-items/{id}/issues/{issue_id}` | covered by teardown |
| `POST /api/v1/work-order-line-items/{id}/sub-line-items` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/work-order-line-items/{id}/sub-line-items/{child_id}` | covered by teardown |
| `POST /api/v1/work-order-statuses` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/work-order-statuses/{id}` | covered by teardown |
| `POST /api/v1/work-order-sub-line-items/{id}/labor-entries` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id}` | covered by teardown |
| `POST /api/v1/work-orders` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/work-orders/{id}` | covered by teardown |
| `POST /api/v1/work-orders/{id}/faults` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/work-orders/{id}/faults/{fault_id}` | covered by teardown |
| `POST /api/v1/work-orders/{id}/issues` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/work-orders/{id}/issues/{issue_id}` | covered by teardown |
| `POST /api/v1/work-orders/{id}/line-items` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `DELETE /api/v1/work-orders/{id}/line-items/{child_id}` | covered by teardown |
| `POST /api/v1/work-orders/{id}/override-total` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `GET /api/v1/work-orders/{id}/status-logs/{child_id}` | no fixture row satisfies this path |
| `POST /auth/login` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /auth/logout` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /auth/refresh` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |
| `POST /auth/switch-company` | collection POST is proved by the fixture build; action POST is owned by the workflow suites |

### Backend operations the frontend never calls

These are proven by Layer 1 but unreachable through the UI. That is
not automatically a defect — it may be unbuilt frontend, or a route
with no screen behind it.

- `GET /api/v1/admin/companies/{id}/owner`
- `GET /api/v1/admin/companies/{id}/roles`
- `POST /api/v1/admin/companies/{id}/set-owner`
- `GET /api/v1/admin/companies/{id}/work-order-statuses`
- `GET /api/v1/admin/employees`
- `GET /api/v1/admin/employees/{id}/companies`
- `PUT /api/v1/admin/employees/{id}/companies`
- `GET /api/v1/asset-statuses`
- `POST /api/v1/asset-statuses`
- `GET /api/v1/asset-statuses/{id}`
- `PUT /api/v1/asset-statuses/{id}`
- `DELETE /api/v1/asset-statuses/{id}`
- `GET /api/v1/asset-trailer-assignments`
- `GET /api/v1/asset-types`
- `POST /api/v1/asset-types`
- `GET /api/v1/asset-types/{id}`
- `PUT /api/v1/asset-types/{id}`
- `DELETE /api/v1/asset-types/{id}`
- `GET /api/v1/assets`
- `POST /api/v1/assets`
- `GET /api/v1/assets/facets`
- `GET /api/v1/assets/{id}`
- `PUT /api/v1/assets/{id}`
- `DELETE /api/v1/assets/{id}`
- `POST /api/v1/assets/{id}/archive`
- `GET /api/v1/assets/{id}/axle-config`
- `PUT /api/v1/assets/{id}/axle-config`
- `DELETE /api/v1/assets/{id}/axle-config`
- `GET /api/v1/assets/{id}/fuel-entries`
- `POST /api/v1/assets/{id}/fuel-entries`
- `GET /api/v1/assets/{id}/fuel-entries/{child_id}`
- `PUT /api/v1/assets/{id}/fuel-entries/{child_id}`
- `DELETE /api/v1/assets/{id}/fuel-entries/{child_id}`
- `POST /api/v1/assets/{id}/restore`
- `GET /api/v1/assets/{id}/trailer`
- `PUT /api/v1/assets/{id}/trailer`
- `DELETE /api/v1/assets/{id}/trailer`
- `GET /api/v1/assets/{id}/trailer-assignments`
- `POST /api/v1/assets/{id}/trailer-assignments`
- `GET /api/v1/assets/{id}/trailer-assignments/{child_id}`
- `PUT /api/v1/assets/{id}/trailer-assignments/{child_id}`
- `DELETE /api/v1/assets/{id}/trailer-assignments/{child_id}`
- `GET /api/v1/assets/{id}/vehicle`
- `PUT /api/v1/assets/{id}/vehicle`
- `DELETE /api/v1/assets/{id}/vehicle`
- `GET /api/v1/axle-definitions/{id}/wheel-positions`
- `POST /api/v1/axle-definitions/{id}/wheel-positions`
- `GET /api/v1/axle-definitions/{id}/wheel-positions/{child_id}`
- `PUT /api/v1/axle-definitions/{id}/wheel-positions/{child_id}`
- `DELETE /api/v1/axle-definitions/{id}/wheel-positions/{child_id}`
- `GET /api/v1/axle-templates`
- `POST /api/v1/axle-templates`
- `GET /api/v1/axle-templates/{id}`
- `PUT /api/v1/axle-templates/{id}`
- `DELETE /api/v1/axle-templates/{id}`
- `GET /api/v1/axle-templates/{id}/definitions`
- `POST /api/v1/axle-templates/{id}/definitions`
- `GET /api/v1/axle-templates/{id}/definitions/{child_id}`
- `PUT /api/v1/axle-templates/{id}/definitions/{child_id}`
- `DELETE /api/v1/axle-templates/{id}/definitions/{child_id}`
- `GET /api/v1/catalog-options`
- `POST /api/v1/catalog-options`
- `GET /api/v1/catalog-options/{id}`
- `PUT /api/v1/catalog-options/{id}`
- `DELETE /api/v1/catalog-options/{id}`
- `GET /api/v1/comments`
- `POST /api/v1/comments`
- `GET /api/v1/comments/{id}`
- `PUT /api/v1/comments/{id}`
- `DELETE /api/v1/comments/{id}`
- `GET /api/v1/companies`
- `POST /api/v1/companies`
- `GET /api/v1/companies/{id}`
- `PUT /api/v1/companies/{id}`
- `DELETE /api/v1/companies/{id}`
- `GET /api/v1/custom-field-definitions`
- `POST /api/v1/custom-field-definitions`
- `GET /api/v1/custom-field-definitions/{id}`
- `PUT /api/v1/custom-field-definitions/{id}`
- `DELETE /api/v1/custom-field-definitions/{id}`
- `GET /api/v1/dashboard/stats`
- `GET /api/v1/employees`
- `POST /api/v1/employees`
- `GET /api/v1/employees/{id}`
- `PUT /api/v1/employees/{id}`
- `DELETE /api/v1/employees/{id}`
- `POST /api/v1/employees/{id}/set-password`
- `GET /api/v1/faults`
- `POST /api/v1/faults`
- `GET /api/v1/faults/{id}`
- `PUT /api/v1/faults/{id}`
- `DELETE /api/v1/faults/{id}`
- `GET /api/v1/fuel-entries`
- `GET /api/v1/fuel-entries/{id}/comments`
- `POST /api/v1/fuel-entries/{id}/comments`
- `GET /api/v1/fuel-entries/{id}/comments/{child_id}`
- `PUT /api/v1/fuel-entries/{id}/comments/{child_id}`
- `DELETE /api/v1/fuel-entries/{id}/comments/{child_id}`
- `GET /api/v1/fuel-entries/{id}/photos`
- `POST /api/v1/fuel-entries/{id}/photos`
- `GET /api/v1/fuel-entries/{id}/photos/{child_id}`
- `PUT /api/v1/fuel-entries/{id}/photos/{child_id}`
- `DELETE /api/v1/fuel-entries/{id}/photos/{child_id}`
- `POST /api/v1/fuel-entries/{id}/photos/{child_id}/set-primary`
- `GET /api/v1/fuel-types`
- `POST /api/v1/fuel-types`
- `GET /api/v1/fuel-types/{id}`
- `PUT /api/v1/fuel-types/{id}`
- `DELETE /api/v1/fuel-types/{id}`
- `GET /api/v1/groups`
- `POST /api/v1/groups`
- `GET /api/v1/groups/{id}`
- `PUT /api/v1/groups/{id}`
- `DELETE /api/v1/groups/{id}`
- `GET /api/v1/inspection-forms`
- `POST /api/v1/inspection-forms`
- `GET /api/v1/inspection-forms/{id}`
- `PUT /api/v1/inspection-forms/{id}`
- `DELETE /api/v1/inspection-forms/{id}`
- `POST /api/v1/inspection-forms/{id}/archive`
- `GET /api/v1/inspection-forms/{id}/items`
- `POST /api/v1/inspection-forms/{id}/items`
- `GET /api/v1/inspection-forms/{id}/items/{child_id}`
- `PUT /api/v1/inspection-forms/{id}/items/{child_id}`
- `DELETE /api/v1/inspection-forms/{id}/items/{child_id}`
- `POST /api/v1/inspection-forms/{id}/restore`
- `GET /api/v1/inspection-submissions`
- `POST /api/v1/inspection-submissions`
- `GET /api/v1/inspection-submissions/{id}`
- `PUT /api/v1/inspection-submissions/{id}`
- `DELETE /api/v1/inspection-submissions/{id}`
- `GET /api/v1/inspection-submissions/{id}/items`
- `POST /api/v1/inspection-submissions/{id}/items`
- `GET /api/v1/inspection-submissions/{id}/items/{child_id}`
- `PUT /api/v1/inspection-submissions/{id}/items/{child_id}`
- `DELETE /api/v1/inspection-submissions/{id}/items/{child_id}`
- `GET /api/v1/inventory-adjustment-reasons`
- `POST /api/v1/inventory-adjustment-reasons`
- `GET /api/v1/inventory-adjustment-reasons/{id}`
- `PUT /api/v1/inventory-adjustment-reasons/{id}`
- `DELETE /api/v1/inventory-adjustment-reasons/{id}`
- `GET /api/v1/inventory-journal-entries`
- `POST /api/v1/inventory-journal-entries`
- `GET /api/v1/inventory-journal-entries/{id}`
- `POST /api/v1/inventory-journal-entries/{id}/reverse`
- `GET /api/v1/issue-priorities`
- `POST /api/v1/issue-priorities`
- `GET /api/v1/issue-priorities/{id}`
- `PUT /api/v1/issue-priorities/{id}`
- `DELETE /api/v1/issue-priorities/{id}`
- `GET /api/v1/issues`
- `POST /api/v1/issues`
- `GET /api/v1/issues/facets`
- `GET /api/v1/issues/{id}`
- `PUT /api/v1/issues/{id}`
- `DELETE /api/v1/issues/{id}`
- `GET /api/v1/issues/{id}/assigned-to`
- `POST /api/v1/issues/{id}/assigned-to`
- `DELETE /api/v1/issues/{id}/assigned-to/{employee_id}`
- `GET /api/v1/issues/{id}/watchers`
- `POST /api/v1/issues/{id}/watchers`
- `DELETE /api/v1/issues/{id}/watchers/{employee_id}`
- `GET /api/v1/locations`
- `POST /api/v1/locations`
- `GET /api/v1/locations/{id}`
- `PUT /api/v1/locations/{id}`
- `DELETE /api/v1/locations/{id}`
- `GET /api/v1/me/permissions`
- `GET /api/v1/measurement-units`
- `POST /api/v1/measurement-units`
- `GET /api/v1/measurement-units/{id}`
- `PUT /api/v1/measurement-units/{id}`
- `DELETE /api/v1/measurement-units/{id}`
- `GET /api/v1/media`
- `POST /api/v1/media`
- `GET /api/v1/media/{id}`
- `PUT /api/v1/media/{id}`
- `DELETE /api/v1/media/{id}`
- `GET /api/v1/notifications`
- `POST /api/v1/notifications/{id}/read`
- `GET /api/v1/part-categories`
- `POST /api/v1/part-categories`
- `GET /api/v1/part-categories/{id}`
- `PUT /api/v1/part-categories/{id}`
- `DELETE /api/v1/part-categories/{id}`
- `GET /api/v1/part-inventory`
- `GET /api/v1/part-locations`
- `POST /api/v1/part-locations`
- `GET /api/v1/part-locations/{id}`
- `PUT /api/v1/part-locations/{id}`
- `DELETE /api/v1/part-locations/{id}`
- `GET /api/v1/part-manufacturers`
- `POST /api/v1/part-manufacturers`
- `GET /api/v1/part-manufacturers/{id}`
- `PUT /api/v1/part-manufacturers/{id}`
- `DELETE /api/v1/part-manufacturers/{id}`
- `GET /api/v1/parts`
- `POST /api/v1/parts`
- `GET /api/v1/parts/{id}`
- `PUT /api/v1/parts/{id}`
- `DELETE /api/v1/parts/{id}`
- `POST /api/v1/parts/{id}/archive`
- `GET /api/v1/parts/{id}/inventory`
- `POST /api/v1/parts/{id}/inventory`
- `GET /api/v1/parts/{id}/inventory/{child_id}`
- `PUT /api/v1/parts/{id}/inventory/{child_id}`
- `DELETE /api/v1/parts/{id}/inventory/{child_id}`
- `POST /api/v1/parts/{id}/restore`
- `GET /api/v1/purchase-order-line-items`
- `GET /api/v1/purchase-orders`
- `POST /api/v1/purchase-orders`
- `GET /api/v1/purchase-orders/{id}`
- `PUT /api/v1/purchase-orders/{id}`
- `DELETE /api/v1/purchase-orders/{id}`
- `POST /api/v1/purchase-orders/{id}/approve`
- `POST /api/v1/purchase-orders/{id}/close`
- `GET /api/v1/purchase-orders/{id}/line-items`
- `POST /api/v1/purchase-orders/{id}/line-items`
- `GET /api/v1/purchase-orders/{id}/line-items/{child_id}`
- `PUT /api/v1/purchase-orders/{id}/line-items/{child_id}`
- `DELETE /api/v1/purchase-orders/{id}/line-items/{child_id}`
- `POST /api/v1/purchase-orders/{id}/override-total`
- `POST /api/v1/purchase-orders/{id}/purchase`
- `POST /api/v1/purchase-orders/{id}/receive-full`
- `POST /api/v1/purchase-orders/{id}/receive-partial`
- `POST /api/v1/purchase-orders/{id}/reject`
- `POST /api/v1/purchase-orders/{id}/revise`
- `GET /api/v1/purchase-orders/{id}/status-logs`
- `GET /api/v1/purchase-orders/{id}/status-logs/{child_id}`
- `POST /api/v1/purchase-orders/{id}/submit`
- `GET /api/v1/roles`
- `POST /api/v1/roles`
- `GET /api/v1/roles/{id}`
- `PUT /api/v1/roles/{id}`
- `DELETE /api/v1/roles/{id}`
- `GET /api/v1/service-entries`
- `POST /api/v1/service-entries`
- `GET /api/v1/service-entries/{id}`
- `PUT /api/v1/service-entries/{id}`
- `DELETE /api/v1/service-entries/{id}`
- `GET /api/v1/service-entries/{id}/line-items`
- `POST /api/v1/service-entries/{id}/line-items`
- `GET /api/v1/service-entries/{id}/line-items/{child_id}`
- `PUT /api/v1/service-entries/{id}/line-items/{child_id}`
- `DELETE /api/v1/service-entries/{id}/line-items/{child_id}`
- `POST /api/v1/service-entries/{id}/override-total`
- `GET /api/v1/service-entry-line-items/{id}/issues`
- `POST /api/v1/service-entry-line-items/{id}/issues`
- `DELETE /api/v1/service-entry-line-items/{id}/issues/{issue_id}`
- `GET /api/v1/service-reminders`
- `POST /api/v1/service-reminders`
- `GET /api/v1/service-reminders/{id}`
- `PUT /api/v1/service-reminders/{id}`
- `DELETE /api/v1/service-reminders/{id}`
- `GET /api/v1/service-tasks`
- `POST /api/v1/service-tasks`
- `GET /api/v1/service-tasks/{id}`
- `PUT /api/v1/service-tasks/{id}`
- `DELETE /api/v1/service-tasks/{id}`
- `POST /api/v1/service-tasks/{id}/archive`
- `GET /api/v1/service-tasks/{id}/parts`
- `POST /api/v1/service-tasks/{id}/parts`
- `GET /api/v1/service-tasks/{id}/parts/{child_id}`
- `PUT /api/v1/service-tasks/{id}/parts/{child_id}`
- `DELETE /api/v1/service-tasks/{id}/parts/{child_id}`
- `POST /api/v1/service-tasks/{id}/restore`
- `GET /api/v1/tire-assignment-requests`
- `POST /api/v1/tire-assignment-requests`
- `GET /api/v1/tire-assignment-requests/{id}`
- `PUT /api/v1/tire-assignment-requests/{id}`
- `DELETE /api/v1/tire-assignment-requests/{id}`
- `POST /api/v1/tire-assignment-requests/{id}/approve`
- `GET /api/v1/tire-models`
- `POST /api/v1/tire-models`
- `GET /api/v1/tire-models/{id}`
- `PUT /api/v1/tire-models/{id}`
- `DELETE /api/v1/tire-models/{id}`
- `GET /api/v1/tire-mount-logs`
- `GET /api/v1/tires`
- `POST /api/v1/tires`
- `GET /api/v1/tires/{id}`
- `PUT /api/v1/tires/{id}`
- `DELETE /api/v1/tires/{id}`
- `GET /api/v1/tires/{id}/inspections`
- `POST /api/v1/tires/{id}/inspections`
- `GET /api/v1/tires/{id}/inspections/{child_id}`
- `PUT /api/v1/tires/{id}/inspections/{child_id}`
- `DELETE /api/v1/tires/{id}/inspections/{child_id}`
- `GET /api/v1/tires/{id}/installations`
- `POST /api/v1/tires/{id}/installations`
- `GET /api/v1/tires/{id}/installations/{child_id}`
- `PUT /api/v1/tires/{id}/installations/{child_id}`
- `DELETE /api/v1/tires/{id}/installations/{child_id}`
- `GET /api/v1/tires/{id}/mount-logs`
- `POST /api/v1/tires/{id}/mount-logs`
- `GET /api/v1/tires/{id}/mount-logs/{child_id}`
- `PUT /api/v1/tires/{id}/mount-logs/{child_id}`
- `DELETE /api/v1/tires/{id}/mount-logs/{child_id}`
- `GET /api/v1/trailer-classifications`
- `POST /api/v1/trailer-classifications`
- `GET /api/v1/trailer-classifications/{id}`
- `PUT /api/v1/trailer-classifications/{id}`
- `DELETE /api/v1/trailer-classifications/{id}`
- `POST /api/v1/uploads`
- `GET /api/v1/vehicle-makes`
- `POST /api/v1/vehicle-makes`
- `GET /api/v1/vehicle-makes/{id}`
- `PUT /api/v1/vehicle-makes/{id}`
- `DELETE /api/v1/vehicle-makes/{id}`
- `GET /api/v1/vehicle-models`
- `POST /api/v1/vehicle-models`
- `GET /api/v1/vehicle-models/{id}`
- `PUT /api/v1/vehicle-models/{id}`
- `DELETE /api/v1/vehicle-models/{id}`
- `GET /api/v1/vendors`
- `POST /api/v1/vendors`
- `GET /api/v1/vendors/{id}`
- `PUT /api/v1/vendors/{id}`
- `DELETE /api/v1/vendors/{id}`
- `POST /api/v1/vendors/{id}/archive`
- `POST /api/v1/vendors/{id}/restore`
- `GET /api/v1/warranties`
- `POST /api/v1/warranties`
- `GET /api/v1/warranties/{id}`
- `PUT /api/v1/warranties/{id}`
- `DELETE /api/v1/warranties/{id}`
- `GET /api/v1/weekly-mileage-goals`
- `POST /api/v1/weekly-mileage-goals`
- `GET /api/v1/weekly-mileage-goals/{id}`
- `PUT /api/v1/weekly-mileage-goals/{id}`
- `DELETE /api/v1/weekly-mileage-goals/{id}`
- `GET /api/v1/work-order-line-items/{id}/issues`
- `POST /api/v1/work-order-line-items/{id}/issues`
- `DELETE /api/v1/work-order-line-items/{id}/issues/{issue_id}`
- `GET /api/v1/work-order-line-items/{id}/sub-line-items`
- `POST /api/v1/work-order-line-items/{id}/sub-line-items`
- `GET /api/v1/work-order-line-items/{id}/sub-line-items/{child_id}`
- `PUT /api/v1/work-order-line-items/{id}/sub-line-items/{child_id}`
- `DELETE /api/v1/work-order-line-items/{id}/sub-line-items/{child_id}`
- `GET /api/v1/work-order-statuses`
- `POST /api/v1/work-order-statuses`
- `GET /api/v1/work-order-statuses/{id}`
- `PUT /api/v1/work-order-statuses/{id}`
- `DELETE /api/v1/work-order-statuses/{id}`
- `GET /api/v1/work-order-sub-line-items/{id}/labor-entries`
- `POST /api/v1/work-order-sub-line-items/{id}/labor-entries`
- `GET /api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id}`
- `PUT /api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id}`
- `DELETE /api/v1/work-order-sub-line-items/{id}/labor-entries/{child_id}`
- `GET /api/v1/work-orders`
- `POST /api/v1/work-orders`
- `GET /api/v1/work-orders/facets`
- `GET /api/v1/work-orders/{id}`
- `PUT /api/v1/work-orders/{id}`
- `DELETE /api/v1/work-orders/{id}`
- `GET /api/v1/work-orders/{id}/faults`
- `POST /api/v1/work-orders/{id}/faults`
- `DELETE /api/v1/work-orders/{id}/faults/{fault_id}`
- `GET /api/v1/work-orders/{id}/issues`
- `POST /api/v1/work-orders/{id}/issues`
- `DELETE /api/v1/work-orders/{id}/issues/{issue_id}`
- `GET /api/v1/work-orders/{id}/line-items`
- `POST /api/v1/work-orders/{id}/line-items`
- `GET /api/v1/work-orders/{id}/line-items/{child_id}`
- `PUT /api/v1/work-orders/{id}/line-items/{child_id}`
- `DELETE /api/v1/work-orders/{id}/line-items/{child_id}`
- `POST /api/v1/work-orders/{id}/override-total`
- `GET /api/v1/work-orders/{id}/status-logs`
- `GET /api/v1/work-orders/{id}/status-logs/{child_id}`
- `POST /auth/login`
- `POST /auth/logout`
- `POST /auth/refresh`
- `POST /auth/switch-company`
- `GET /healthz`

## Teardown

| Outcome | Count | Rows |
|---|---|---|
| Deleted | 49 | laborEntry#18, workOrderSubLineItem#19, wheelPosition#18, serviceEntryLineItem#18, purchaseOrderLineItem#16, workOrderLineItem#19, serviceTaskPart#2, axleConfig, trailer, vehicle, inspectionFormItem#16, trailerAssignment#18, weeklyMileageGoal#18, warranty#17, partInventory#20, fuelEntry#18, serviceEntry#19, purchaseOrder#19, issue#18, workOrder#20, employee#21, axleDefinition#19, tire#16, trailerAsset#46, asset#45, part#19, vehicleModel#23, tireModel#18, inspectionForm#18, fault#18, group#19, role#31, serviceTask#18, axleTemplate#19, vehicleMake#23, workOrderStatus#70, issuePriority#18, trailerClassification#18, fuelType#18, adjustmentReason#18, partLocation#20, measurementUnit#18, partManufacturer#18, partCategory#18, assetStatus#48, assetType#18, location#18, vendor#19, journalEntry#29 |
| Archived instead | 0 | — |
| Reversed (neutralised, not removed) | 0 | — |
| Failed to remove | 0 | — |
