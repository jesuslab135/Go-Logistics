# Backend work requested by the frontend

Capabilities the client (`fleet-admin-app`) needs and this API doesn't provide
yet. Each entry says what the frontend does today without it, why that's a
problem, and the concrete route/shape that would unblock it.

These were collected while reconnecting the client to this backend, route by
route. **Nothing here is a bug report** — for defects in routes that already
exist, see [`KNOWN-ISSUES.md`](./KNOWN-ISSUES.md).

Where an entry says "detail in `<path>`", that path is inside the **frontend**
repo (`fleet-admin-app`), which holds the full per-route field mappings and the
UI-side context. The specs below are self-contained; the link is for background.

Nothing in the client has been deleted over these gaps — UI is kept in place and
only the behavior is disabled, so reactivating is usually a matter of swapping a
query once the route exists.

---

## Priority 1 — blocking basic product flows

### 1.1 Set an employee's password

`POST /api/v1/employees/{id}/set-password` (admin-initiated), and/or
`POST /api/v1/employees/{id}/invite` (email a set-password link).

- Request: `{ "password": string }` for a direct set; empty for the invite variant.
- Auth: caller must be an admin (`role.is_admin`) of the target employee's company.
  Note this would be **the first route in the app enforcing such a check** — see 3.1.

**Why:** `EmployeeResponse`/`Create`/`Update` have no password field at all
(`password_hash excluded` in the router), so creating an employee produces a
personnel record that **cannot log in**. The only way to set a password today is
`cli setpass`, which isn't reachable from a browser. This blocks the most basic
"add a teammate" flow. The client shows a standing notice in the create drawer
rather than implying the new employee can sign in.

Detail: `docs/backend-migration/employees/README.md`.

### 1.2 Seed default `asset_status` rows on company creation

Not a route — a change to `POST /api/v1/companies`. Tracked as
[`KNOWN-ISSUES.md`](./KNOWN-ISSUES.md) #3 because it also has a defect character:
a company is currently created in a state where **no asset can be created**, since
`asset.status_id` has nothing to point at and no product surface can create a
status.

### 1.3 `asset_id` filter on work orders

`GET /api/v1/work-orders?asset_id=<id>`

**Why:** to show an asset's work orders, the client currently drains **every work
order for the whole company** in pages of 100 and filters client-side. Unlike the
small bounded catalogs, work-order history grows without limit — this is the most
expensive workaround in the whole migration and it degrades continuously.

Implementation: pass the param through to the `ListWorkOrders` query.

Detail: `docs/backend-migration/work-orders/README.md`.

---

## Priority 2 — capability gaps with real UI disabled behind them

### 2.1 List the companies an employee belongs to

`GET /api/v1/employees/me/companies` (or `/api/v1/employees/{id}/companies`)

- Auth: authenticated; returns rows for the **calling** employee (JWT `sub`), not a global list.
- Response: `CompanyResponse[]`, or a light `{id, name}[]` projection.

**Why:** the JWT carries exactly one `company_id`, fixed at login from
`employee.default_company_id`, and nothing exposes an employee's membership rows —
even though `employee_companies` exists in the schema with a
`UNIQUE(employee_id, company_id)` constraint. The company switcher in the top bar
is fully built and handles the N>1 case correctly; it just never sees more than
one company, so it never becomes reachable.

Frontend cost once it exists: swap one call in `use-my-companies.ts`. No UI changes.

### 2.2 Switch the active company mid-session

`POST /auth/switch-company`

- Request: `{ "company_id": number }`; authenticated.
- Server-side: verify the caller has an `employee_companies` row for that company,
  then issue a fresh pair via the same `TokenService.Issue` used by login/refresh.
- Response: `AuthTokenPair`.

**Why:** every request is scoped off the token's `company_id` claim, and no route
re-scopes an existing session. Re-logging in doesn't help either — the verifier
always resolves back to the same `default_company_id`. Depends on 2.1 to be useful.

Detail for both: `docs/backend-migration/companies/README.md`.

### 2.3 Approve a tire-assignment request (atomically)

`POST /api/v1/tire-assignment-requests/{id}/approve`

- Request: `{ "odometer_at_install": number, "tread_depth_at_install_32nds"?: number, "psi_at_install"?: string, "notes"?: string }`
- Response: the updated `TireAssignmentRequestResponse` (`state: APPROVED`,
  `approved_by_id`/`resolved_at` set server-side from the auth context).
- Server-side, **in one transaction**: create the `TireInstallation`, write a
  `TireMountLog` (`event_type: "INSTALL"`), and update the `Tire`'s
  `status`/`current_vehicle_id`/`current_position_code`.
- Should return 409 if the tire already has an active installation
  (`TireInstallation.tire_id` is unique).

**Why:** `/tire-assignment-requests` is a plain CRUD handler — create/update are
pure passthroughs with no side effects, and there's no approve action. Doing this
from the client would take 3-4 non-atomic requests, any of which can fail and
leave a request marked approved with the tire never actually mounted. One of them
also needs an `odometer_at_install` the request form never collects. The client
removed the approve button and documents the manual workaround instead.

Detail: `docs/backend-migration/tires/README.md`.

### 2.4 Reverse lookup: which truck is towing this trailer

Either of:

- `GET /api/v1/asset-trailer-assignments?trailer_id=<id>` (company-wide list with a filter), or
- `GET /api/v1/assets/{id}/towed-by` (id = the trailer).

**Why:** assignments are reachable **only** nested under the towing asset, so
"who tows trailer X" would require draining every asset's assignments company-wide.
The trailer profile's towing tab is present but always empty.

The first option is more reusable — it also solves 4.1 below.

Detail: `docs/backend-migration/asset-trailer-assignments/README.md`.

### 2.5 Self-service signup

`POST /auth/signup` — public, no `/api` prefix, matching `/auth/login`.

- Request: `{ first_name, last_name, email, password }`.
- Server-side: create the `employee` with a bcrypt hash (same as
  `EmployeeCredentialVerifier`/`cli setpass`), **and** decide how the new employee
  gets a `default_company_id` — that decision is the actual blocker, not the insert.
- Response: `AuthTokenPair` (log them straight in).

Related: reactivating self-serve company creation additionally needs an
`is_account_owner`-equivalent **JWT claim**, sourced from the `employee.is_account_owner`
column that already exists but nothing reads or sets.

The signup page keeps all its fields and validation, wrapped in a disabled
`<fieldset>` with an explanatory notice.

Detail: `docs/backend-migration/auth/README.md`.

---

## Priority 3 — security and admin surface

### 3.1 An admin authorization gate

There is currently **no permission check on any route** beyond "is authenticated".
`role.is_admin` exists as a boolean and `role.permissions` is a jsonb column
nothing reads. Several requests above (1.1, 3.2) need an admin gate to be
implementable safely, and [`KNOWN-ISSUES.md`](./KNOWN-ISSUES.md) #6 is the same
problem showing up as a defect on `/api/v1/companies`.

Worth deciding the model once — claim in the JWT vs. per-request role lookup —
rather than per route.

### 3.2 Cross-company admin surface

All of this needs 3.1 first. The client's `features/admin` has these built and
disabled in place:

- **Company owner:** an `owner` field on `CompanyResponse` (computed from
  `employee.is_account_owner`), or `GET /api/v1/admin/companies/{id}/owner`; plus
  `POST /api/v1/admin/companies/{id}/set-owner` with `{ employee_id }`, atomically
  clearing the previous owner. Note Go has no separate auth-user id — an `employee`
  row *is* the identity.
- **Cross-company employee directory:** `GET /api/v1/admin/employees` (unlike
  `/employees`, not scoped to the caller's company), plus
  `PUT /api/v1/admin/employees/{id}/companies` as a full replace over
  `employee_companies`: `{ company_ids: number[], default_company_id: number | null }`.
- **Provisioning verification:** `/roles` and `/work-order-statuses` are scoped to
  the caller's own JWT with no override, so they can't answer "was company X seeded
  correctly". Needs either admin-scoped
  `GET /api/v1/admin/companies/{id}/roles` (+ `/work-order-statuses`), or a
  `company_id` override honored only for admin tokens.

Detail: `docs/backend-migration/missing-backend-capabilities.md`.

### 3.3 Refresh-token revocation (logout)

`POST /auth/logout` — authenticated, body `{ refresh_token }`.

Refresh tokens are stateless JWTs validated purely by signature and expiry, so
there is nothing to revoke against and no logout route. **A "logged out" user's
refresh token stays valid until its natural `JWT_REFRESH_TTL` expiry — 7 days by
default** — if it was captured beforehand.

Needs: a `jti` claim on the token pair, a `revoked_refresh_token(jti, expires_at)`
table, an insert on logout, and a denylist check in `TokenService.ParseRefresh`.

Flagging this as a security decision rather than a frontend nicety; the client's
logout already works locally and calls a documented no-op.

Detail: `docs/backend-migration/auth/README.md`.

---

## Priority 4 — performance and polish

Each of these is worked around client-side today and works correctly; they're
listed so they aren't lost.

### 4.1 Denormalized names on nested responses

Several responses carry only raw FK ids, forcing an extra request or a client-side
join per row:

| Response | Missing | Client workaround today |
|---|---|---|
| `AssetTrailerAssignmentResponse` | `trailer_name` | one `GET /assets/{id}` per assignment |
| `WorkOrderResponse` | `status_name` | shows `#<status_id>` |
| `WarrantyResponse` | `provider_name` | shows `#<provider_id>` |

Bounded and acceptable at current scale (a truck has 1-3 trailers), but they are
per-row round-trips.

### 4.2 List filters on assets

`GET /api/v1/assets` takes only `page`/`page_size`. Useful params: `vehicle_type`,
`status_id`, and `q` (free-text over name/vin_sn/license_plate). The client filters
in memory today — and did so under Django too, so this is not a regression.

### 4.3 Bulk trailer fields

`trailer_type`, `classification` and `size` live only on the nested
`GET /api/v1/assets/{id}/trailer`, so the trailers registry can't show them
without one request per row. Three columns and the subtype filter tabs are
present but inert because of it.

Either fold those fields into `AssetResponse` for trailer-typed assets (smaller
change, mirrors how `make`/`model` already live on Asset), or add a
`GET /api/v1/trailers` list that joins Asset+Trailer server-side.

Detail: `docs/backend-migration/trailer/README.md`.

### 4.4 `category` filter on catalog options

`GET /api/v1/catalog-options?category=<name>`. The client drains all pages once
and filters in memory, cached an hour. Cheap at current scale — noted so it isn't
forgotten if catalogs grow.

### 4.5 Ordering and pagination on nested fuel entries

`GET /api/v1/assets/{id}/fuel-entries` takes no params at all; the client sorts
one page newest-first in memory. Low priority — one request, not a loop.

---

## Not gaps (recorded so they aren't re-investigated)

- **`asset_type` and `catalog_option` empty on a new company** — unlike
  `asset_status` (1.2), neither blocks the client: nothing reads `asset_type`, and
  catalog options are created on demand by the combobox.
- **Django's `auth.Group`** — no equivalent here and nothing in the client uses it.
- **Django's `/api/schema/`** — replaced by `/swagger/doc.json`, already wired into
  the client's codegen.
