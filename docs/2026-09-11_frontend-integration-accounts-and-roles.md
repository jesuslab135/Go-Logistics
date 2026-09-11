# Frontend integration guide: accounts, onboarding and per-company roles

**Date:** 2026-09-11
**Audience:** the frontend developer (`go-logistics.netlify.app`)
**Backend:** live at `https://go-logistics.jesuslab135.com`, `main` @ `f2d91ab`
**Covers:** subsystem 1, accounts and self-service onboarding (PR #8), and
subsystem 2, per-company roles and directors (PR #9). Both are deployed.
**Machine-readable contract:** `docs/swagger.json` / `docs/swagger.yaml` in this repo.

This guide covers everything the frontend has to change or build. Each item
gives the exact endpoint, payload, response and error. The last section maps
each item onto the screens the live frontend already has.

---

## 0. Work list, in priority order

| # | Priority | What | Why | Section |
|---|---|---|---|---|
| 1 | **P0** | Build the "create your first company" flow on `/onboarding/company` | A newly provisioned **account owner** lands there and is told to "ask your account owner". They *are* the owner, so they are stuck | [4.2](#42-the-owners-first-login-and-first-company) |
| 2 | **P0** | Gate the app on `company_id`, not on `is_admin` or `modules` | A company-less owner gets `is_admin: true` with every module listed, but every module route answers `403` | [3.2](#32-the-company-less-trap) |
| 3 | **P0** | Replace the global "Rol" everywhere on the Employees screen | `role_id` no longer exists on employees. The Rol column shows "Sin rol" for everyone, the Rol filter matches nobody, and the Rol dropdown in the form saves nothing | [5.2](#52-what-broke-on-existing-screens) |
| 4 | **P0** | Build per-company role assignment (account owner) | This is the only way to give anyone a role now | [5.3](#53-list-the-accounts-people-and-their-roles), [5.4](#54-assign-companies-and-roles-to-a-person) |
| 5 | P1 | Roles screen: fix the "Empleados" count | It was computed from `employee.role_id` and now always shows 0 | [5.5](#55-roles-screen) |
| 6 | P1 | Remove or rework the "Propietario de la cuenta" toggle | It writes a legacy flag that no permission check reads | [5.2](#52-what-broke-on-existing-screens) |
| 7 | P1 | Platform admin: add an **Accounts** page (create client + owner, list, transfer ownership) | Provisioning a client is currently API-only | [4.1](#41-platform-admin-provisions-a-client) |
| 8 | P1 | Show the role in effect for the current company | `/me/permissions` now reports it (`role`) | [3.1](#31-get-apiv1mepermissions) |
| 9 | P2 | Handle `403 no_company_membership` on token refresh | A revoked membership now invalidates the refresh | [2.3](#23-refresh-and-switch-company) |

---

## 1. The model in one page

```
Platform admin     the IT team. is_platform_admin; belongs to NO account; uses /api/v1/admin/*
└── Account        one client organisation. Exactly one owner (transferable)
    ├── Companies  the tenants that hold fleet data. Each belongs to exactly one account
    └── Employees  each belongs to exactly one account
        └── Membership (employee ⇄ company), carrying a ROLE for that company
```

Rules the UI must reflect:

1. **A role belongs to a membership, not a person.** The same person can be an
   administrator in company A, read-only in company B, and absent from company C.
   `employee.role_id` no longer exists.
2. **A membership with no role grants nothing.** The person can switch into the
   company, but every module screen refuses them, reads included.
3. **The account owner is an administrator in every company of their account**
   because they own it, whatever role their own membership carries. They cannot
   lock themselves out.
4. **A "director"** is not a separate entity. It is a person holding an
   administrator role in some of the account's companies.
5. **Accounts are isolated by the database.** Nothing in the UI can place a
   person or role across accounts. The API refuses with `404`, as if the target
   did not exist.
6. **There is no public sign-up.** Clients are provisioned by a platform admin.
   The existing `/signup` page, a static "self-service sign-up isn't available"
   notice, is correct as it is.

---

## 2. Auth and session

### 2.1 Token pair and claims

All of `/auth/login`, `/auth/refresh` and `/auth/switch-company` return:

```json
{ "access_token": "…", "refresh_token": "…", "token_type": "Bearer", "expires_in": 3600 }
```

Access-token claims:

| Claim | Type | Meaning |
|---|---|---|
| `sub` | string | employee id |
| `company_id` | number, **may be absent** | the session's company. **Absent** (the key is missing, not `0` or `null`) for a company-less session |
| `account_id` | number, may be absent | the employee's account. Absent for platform staff |
| `is_admin` | bool | admin in the session's company at issue time. **Do not use it for UI decisions**; use `/me/permissions`, which is re-read from the database on every request |
| `typ` | string | `access` or `refresh` |

The login page already routes on `claims.companyId` (`/app/dashboard` or
`/onboarding/company`). Keep that.

### 2.2 `POST /auth/login`

Body: `{ "email": "…", "password": "…" }`

| Situation | Result |
|---|---|
| Has one or more company memberships | `200`, token scoped to `default_company_id`, else the oldest membership |
| Belongs to an account but has **no memberships**: a new account owner, or an employee whose access was all revoked | `200`, **company-less** token (no `company_id` claim) |
| Platform staff with no memberships | `403` `{"error":{"code":"no_company_membership", …}}` |
| Wrong email or password | `401` `unauthorized` |

### 2.3 Refresh and switch-company

**`POST /auth/refresh`** body `{ "refresh_token": "…" }`. It now re-checks that the
employee is still a member of the token's company. If not, it answers:

```json
{ "error": { "code": "no_company_membership",
             "message": "this account is no longer a member of the company this token was issued for" } }
```

with status `403`. **Handle it by clearing the session and sending the user to
`/login`.** Logging in again lands them in a company they still belong to, or
company-less. A company-less refresh token is simply re-issued company-less.

**`POST /auth/switch-company`** body `{ "company_id": 12 }`, with the current
access token as Bearer. It returns a new pair scoped to that company, and
`is_admin` is recomputed **for that company's role**. Switching company
therefore switches role, so after switching, refetch `/me/permissions` (the app
already clears its query cache here). Not a member: `403` "you are not a member of
this company". It works from a company-less session too; that is how an owner
enters their first company.

### 2.4 What a company-less session can reach

Only these:

- `GET /api/v1/me/permissions`
- `POST /api/v1/companies` (account owner only)
- `GET /api/v1/account/employees`, `PUT /api/v1/account/employees/{id}/companies` (account owner only)
- `POST /auth/switch-company`, `/auth/refresh`, `/auth/logout`

Every company-scoped route answers `403` with the message
"no active employee profile is associated with this company".

---

## 3. `/me/permissions`: the single source of truth for the UI

### 3.1 `GET /api/v1/me/permissions`

```json
{
  "employee": {
    "id": 42, "first_name": "Dana", "last_name": "Director", "email": "dana@acme.mx",
    "job_title": "Director", "is_active": true, "is_technician": false,
    "is_vehicle_operator": false, "default_company_id": 7
  },
  "company_id": 7,
  "role": { "id": 19, "name": "Administrador", "is_admin": true },
  "is_admin": true,
  "is_account_owner": false,
  "is_platform_admin": false,
  "permissions": {
    "assets": { "read": true, "create": true, "update": true, "delete": true },
    "fuel":   { "read": true, "create": true, "update": true, "delete": true }
  },
  "modules": ["assets", "company", "employees", "fuel"],
  "companies": [ { "id": 7, "name": "Acme Norte", "logo": null },
                 { "id": 9, "name": "Acme Sur",   "logo": null } ]
}
```

| Field | Use it for |
|---|---|
| `company_id` | **Absent** for a company-less session. Check it first (see 3.2) |
| `role` | **New.** The role held in *this* company. `null` when the membership has no role, or the session is company-less. The owner may have `role: null` and still `is_admin: true` |
| `is_admin` | Admin in *this* company: the role says so **or** the caller owns the account |
| `is_account_owner` | The caller owns their account. This is the **real** ownership (`account.owner_employee_id`). It gates company creation and role assignment |
| `is_platform_admin` | Show the `/app/admin/*` area |
| `permissions` / `modules` | Build navigation and enable buttons, **only when `company_id` is present** |
| `companies` | The company switcher (already used by `use-my-companies`). It lists only companies the person has a membership in, which is how a director sees just their companies |

Suggested label for the role in the shell:
`role?.name ?? (is_account_owner ? "Propietario de la cuenta" : "Sin rol")`.

### 3.2 The company-less trap

For an account owner with no company yet, `/me/permissions` returns
`is_admin: true` (from ownership), a `permissions` map with **every action
`true`**, and **every module** in `modules`, but **no `company_id`** and
`companies: []`. Every module route still answers `403`, because there is no
company to be a member of.

**Rule: if `company_id` is absent, render onboarding and never the app shell.**
Build navigation from `modules` only when `company_id` is present. (A backend
change to report empty permissions in this state is listed in section 8.)

---

## 4. Subsystem 1: accounts and onboarding

### 4.1 Platform admin provisions a client

Show this only when `is_platform_admin`. Add it to the existing admin area, next
to `/app/admin/companies` and `/app/admin/users`, e.g. as `/app/admin/accounts`.

**Create a client and its owner in one call:** `POST /api/v1/admin/accounts`

```json
{ "name": "Acme Fleet", "owner_first_name": "Ada", "owner_last_name": "Byron",
  "owner_email": "ada@acme.mx", "owner_password": "at-least-12-chars" }
```

`201`:

```json
{ "id": 3, "name": "Acme Fleet", "owner_employee_id": 57, "owner_email": "ada@acme.mx",
  "is_active": true, "created_at": "2026-09-11T14:02:11Z", "company_count": 0, "employee_count": 1 }
```

Errors:
- `422 validation_failed`: `details` names **every** bad field at once, e.g.
  `{"name":"this field is required","owner_password":"must be at least 12 characters"}`.
  All five fields are required, and the password must be at least 12 characters.
- `409 conflict`: the owner email is already used by an active employee.
  `details` is `{"email": "already in use"}`. The key is **`email`**, not
  `owner_email`, so map it onto the owner-email input yourself.

The password is never returned. Hand it to the client out of band.

**List clients:** `GET /api/v1/admin/accounts?limit=20&offset=0`

```json
{ "data": [ { "id": 3, "name": "Acme Fleet", "owner_employee_id": 57, "is_active": true,
              "created_at": "…", "company_count": 2, "employee_count": 9 } ],
  "total": 4, "limit": 20, "offset": 0, "has_next": false }
```

`owner_email` is **not** filled in the list, only in the create response.

**Transfer ownership:** `POST /api/v1/admin/accounts/{id}/set-owner` with
`{ "employee_id": 61 }` returns `200` with the account. `404` means no such
account. `422 {"employee_id":"must be an employee who belongs to this account"}`
means the person belongs to a different client.

> ⚠️ The existing **set-owner dialog** in the admin area was built on
> `POST /api/v1/admin/companies/{id}/set-owner` and `GET …/owner`. Those routes
> read and write the **legacy** `employee.is_account_owner` flag, which no
> permission check uses any more. To make someone the real owner (the person
> who can create companies and assign roles), use
> `/admin/accounts/{id}/set-owner` above.

### 4.2 The owner's first login and first company

1. The owner logs in and gets a **company-less** token. The login page already
   sends them to `/onboarding/company`.
2. **Today that page is a dead end.** It shows "No company yet. You're not
   linked to a company yet. Ask your account owner to add you, then sign in
   again." That message is right for an ordinary employee with no access. It is
   wrong for the owner.
3. Branch the page on `/me/permissions`:
   - **`is_account_owner: true`**: show a create-company form (the page already
     has the `{name, tax_id}` form state wired, just not submitted).
   - **otherwise**: keep the current message.
4. On submit: `POST /api/v1/companies`

   ```json
   { "name": "Acme Norte", "tax_id": "ANO-010101-AB1",
     "address": "", "phone": "", "email": "", "website": "", "logo": null,
     "city": "", "region": "", "postal_code": "",
     "country": "MX", "timezone": "America/Mexico_City", "currency": "MXN",
     "system_of_measurement": "metric" }
   ```

   Only `name` (max 200) and `tax_id` (max 50) are required. Omitted
   `country`, `timezone`, `currency` and `system_of_measurement` default to
   `MX`, `America/Mexico_City`, `MXN` and `metric`. `system_of_measurement` must be
   `metric` or `imperial`. **Do not send `account_id`**: the account always comes
   from the session, and a body value is ignored.

   `201` returns the company (`id`, `name`, `tax_id`, …, `created_at`). In the same
   transaction the backend:
   - seeds the roles **`Administrador`** (admin) and **`Almacén`** (warehouse),
     plus default work-order and asset statuses;
   - gives the owner a membership carrying `Administrador`;
   - sets the owner's `default_company_id` to it if they had none, so their next
     login lands there.
5. **Then enter it.** The current token is still company-less:
   `POST /auth/switch-company {"company_id": <new id>}`, store the new pair,
   refetch `/me/permissions`, and navigate to `/app/dashboard`. Skipping this
   step bounces the owner straight back to onboarding.

Errors on `POST /api/v1/companies`: `403` "account owner privileges are
required for this action" for a non-owner. `403` "only a member of a client
account can create a company" for platform staff. `422 validation_failed` with
field `details`.

An owner can create more companies later from any session with the same call.
The new company appears in `me.companies`, and `switch-company` enters it.

---

## 5. Subsystem 2: per-company roles

### 5.1 Who can do what

| Action | Who | Endpoint |
|---|---|---|
| See everyone in the account, with their role in each company | Account owner | `GET /api/v1/account/employees` |
| Grant or revoke companies and set the role in each | Account owner | `PUT /api/v1/account/employees/{id}/companies` |
| Create, edit or delete **a company's roles** | An administrator **of that company** (session's company) | `/api/v1/roles` |
| Create an employee | A user with the `employees` module in the session's company | `POST /api/v1/employees` |

A newly created employee (`POST /api/v1/employees`) gets **one membership, in the
session's company, with no role.** They can log in and switch into it, but can
do nothing until the account owner assigns a role (5.4). After "Agregar
empleado" succeeds, the UI should say so. If the current user is the owner,
offer "Asignar rol" right away.

### 5.2 What broke on existing screens

The employee payloads (`GET/POST/PUT /api/v1/employees`) **no longer contain
`role_id`**, in either requests or responses. If the client still sends it, the
backend silently ignores it, so nothing errors. It just does nothing.

| Screen (current UI) | Symptom today | Fix |
|---|---|---|
| Organización → Empleados → **Rol** column | "Sin rol" for everyone | For the owner, read it from `GET /account/employees` (the membership for the session's company). Otherwise hide the column, or show it only for the current user from `me.role` |
| Empleados → **Rol** filter | Matches nobody | Same data source as above, or remove |
| Editar empleado → Empleo → **Rol** dropdown ("Determina si este empleado ve el panel de administración") | Shows "Sin rol"; saving changes nothing | **Remove it.** Role assignment moves to the per-company screen in 5.4 |
| Editar empleado → **Propietario de la cuenta** toggle | Writes `employee.is_account_owner`, which nothing reads | Remove it, or show real ownership read-only. Ownership is transferred only by a platform admin (4.1) |
| Organización → Roles → **Empleados** count | Always 0 | See 5.5 |

### 5.3 List the account's people and their roles

`GET /api/v1/account/employees` (account owner only; works in any session,
company-less included):

```json
[
  { "employee_id": 57, "first_name": "Ada", "last_name": "Byron", "email": "ada@acme.mx",
    "is_active": true,
    "memberships": [
      { "company_id": 7, "company_name": "Acme Norte", "role_id": 19, "role_name": "Administrador", "role_is_admin": true },
      { "company_id": 9, "company_name": "Acme Sur",   "role_id": null, "role_is_admin": false }
    ] },
  { "employee_id": 61, "first_name": "Dana", "last_name": "Director", "email": "dana@acme.mx",
    "is_active": true, "memberships": [] }
]
```

- A plain array, not paginated.
- `memberships: []` means the person belongs to no company yet. They still appear, and
  this is where they get their first one.
- `role_id: null` means a membership with no role (grants nothing). `role_name` is then **omitted**.

Non-owner: `403` "account owner privileges are required for this action".

### 5.4 Assign companies and roles to a person

`PUT /api/v1/account/employees/{id}/companies` (account owner only):

```json
{ "grants": [
    { "company_id": 7, "role_id": 19 },
    { "company_id": 9, "role_id": 23 },
    { "company_id": 11, "role_id": null }
] }
```

**It is a full overwrite, not a patch:**
- every company listed gets exactly that role, including companies the person
  already belonged to, which is how a role is changed;
- **every company not listed is revoked**;
- `role_id: null` (or omitting it) gives membership with no role;
- `{"grants": []}` revokes everything.

So build the body from the person's **current** `memberships` (5.3) plus the
edits, and send the whole set. Sending only the changed company revokes the
others.

`200` returns that one person in the shape of 5.3, already updated. Use it to
update the row without refetching. Every grant and revoke is written to an
audit log.

Errors:

| Status | `details` / message | Meaning, and what to show |
|---|---|---|
| `422` | `{"role_id":"does not belong to the named company"}` | the role picked belongs to another company |
| `422` | `{"company_id":"named more than once"}` | the same company appears twice in `grants` |
| `422` | `{"company_id":"must be positive"}` | bad id |
| `404` | "employee not found" | the person is not in the caller's account |
| `404` | "company not found" | the company is not in the caller's account |
| `400` | "invalid id" / "invalid request body" | malformed request |

The owner may edit their own row. Removing their own role in a company does
not lock them out (rule 3 in section 1).

**Role options for the picker.** `GET /api/v1/roles` lists the roles of the
**session's company only**. The API has no account-scoped route listing
another company's roles yet (see section 8). Until there is one:
- company of the current session: `GET /api/v1/roles`;
- any other company: call `POST /auth/switch-company {"company_id": X}`, use the
  returned access token **only** for `GET /api/v1/roles`, and discard it. The
  current session's tokens stay valid, since issuing a pair revokes nothing.
  Every company created through the API is seeded with `Administrador` and
  `Almacén`, so those two always exist.

### 5.5 Roles screen

`/api/v1/roles` (administrator of the session's company) is unchanged: list,
create, update, delete. It is scoped to the session's company, so an admin of
two companies manages each company's roles after switching into it.

```json
{ "id": 19, "company_id": 7, "name": "Administrador", "is_admin": true, "permissions": {} }
```

`permissions` is a JSON object keyed by module:
- `{"assets": {}}`: an **empty object grants the whole module**;
- `{"fuel": {"read": true, "create": true}}`: only those actions;
- a module not listed is denied;
- actions are `read`, `create`, `update`, `delete`, plus `approve` for
  `tire_approvals` and `purchase_orders`;
- when `is_admin: true`, the permissions document is ignored and everything is allowed.

Modules: `assets`, `company`, `employees`, `fuel`, `inspections`, `inventory`,
`issues`, `mileage_goals`, `parts`, `purchase_orders`, `roles`, `service`,
`tire_approvals`, `tires`, `vendors`, `warranties`, `work_orders`.

Three of these are reported but **not** enforced as modules. The `company` and
`roles` screens (`/api/v1/companies`, `/api/v1/roles`) require `is_admin`, so
granting `{"roles": {}}` in a role document does not open the Roles screen.
`tire_approvals` is checked through its `approve` action. In a role editor,
show `company` and `roles` as "administrators only" rather than as toggles
that appear to work.

**The "Empleados" count** was derived from `employee.role_id`. The correct count
is the number of memberships in this company whose `role_id` equals the role's
id. It can be derived from `GET /account/employees`, but only the owner can
call that. For non-owners, hide the column.

---

## 6. Errors

Every error uses one envelope:

```json
{ "error": { "code": "validation_failed", "message": "validation failed",
             "details": { "role_id": "does not belong to the named company" } } }
```

`details` is present only when it adds information. For `422` it is always a
`{field: reason}` map, so you can mark the inputs.

| Status | `code` | Typical messages |
|---|---|---|
| 400 | `bad_request` | "invalid request body", "invalid id" |
| 401 | `unauthorized` | "missing bearer token", "invalid or expired token", "access token required", "invalid credentials", "invalid refresh token" |
| 403 | `forbidden` | "no active employee profile is associated with this company" (not a member / company-less), "an administrator role is required for this action", "you do not have access to this module", "you do not have permission to perform this action", "account owner privileges are required for this action", "platform administrator privileges are required for this action", "you are not a member of this company" (switch) |
| 403 | `no_company_membership` | login (platform staff with nothing) or refresh (membership revoked) |
| 404 | `not_found` | target not found or outside your account |
| 409 | `conflict` | "email is already in use" with `details: {"email": "already in use"}` |
| 422 | `validation_failed` | field map in `details` |

Every `403` gate shares the code `forbidden`. Don't branch on message text.
Decide what to render from `/me/permissions` so the user never reaches a `403`
in normal use.

---

## 7. Where each change lands in the current frontend

Observed in the live bundle on 2026-09-11. Names are routes and source modules,
not hashed file names.

| Route / module | Today | Change |
|---|---|---|
| `/login` | Routes on `claims.companyId` → `/app/dashboard` or `/onboarding/company` | Keep |
| `/onboarding/company` | Static "No company yet… ask your account owner" | Owner: create company, then switch-company (4.2). Others: keep the message |
| App shell guard (`/app/*`) | Redirects to onboarding when the token has no company | Keep. Also never build nav from `modules` without `company_id` (3.2) |
| `use-my-permissions` / `use-my-companies` | Reads `/me/permissions`; company list from `companies` | Add `role` display (3.1) |
| Company switcher (`/auth/switch-company`) | Works; clears the cache | Refetch `/me/permissions` after switching (role changes) |
| `/app/organization/employees` | Rol column, Rol filter, form Rol dropdown, account-owner toggle | 5.2, and add per-company role assignment for the owner (5.3, 5.4) |
| `/app/organization/roles` | Employee count 0 | 5.5 |
| `/app/admin/companies`, `/app/admin/users`, set-owner dialog | Built on `/admin/companies/{id}/set-owner` (legacy flag) | Add `/app/admin/accounts` (4.1); point ownership transfer at `/admin/accounts/{id}/set-owner` |
| `/signup` | Static "self-service sign-up isn't available" notice | Keep (by design) |
| Token refresh handler | — | On `403 no_company_membership`: clear session and go to `/login` (2.3) |

Suggested screen for 5.3 and 5.4: a person, then one row per account company,
each with a checkbox (member) and a role select. Save sends the whole set in
one `PUT`.

---

## 8. Known backend gaps (don't wait on these; work around them as described)

1. **Company-less `/me/permissions` reports admin rights.** See 3.2. The frontend
   rule (gate on `company_id`) is correct either way. The backend may later
   return empty permissions in that state.
2. **No account-scoped listing of another company's roles.** See the workaround
   in 5.4. A candidate is `GET /api/v1/account/companies/{id}/roles`.
3. **Only the account owner can assign roles.** A company administrator
   (director) can create employees but cannot give them a role. If directors
   need that, it needs a new backend route.
4. **Legacy `employee.is_account_owner`** still exists and is still written by
   the employee form and by `/admin/companies/{id}/set-owner`. It grants
   nothing. It will be removed later.
5. **Employee 1** currently holds both the "Go Logistics" account ownership and
   the platform-admin flag. Splitting them is pending.

---

## 9. Acceptance checklist

Use a platform-admin login.

- [ ] **Provision:** Admin → Accounts → create "QA Client" with owner `qa-owner@…`
      and a password of at least 12 characters. It appears in the list with `company_count 0`.
- [ ] **Owner first login:** log in as the owner, land on `/onboarding/company`,
      see the create-company form (not the "ask your account owner" text).
- [ ] **First company:** create "QA Norte", land on the dashboard inside it.
      The shell shows role "Administrador". Log out and back in: you land in QA Norte directly.
- [ ] **Second company:** create "QA Sur" from inside the app. It appears in the
      switcher, and switching works.
- [ ] **Add an employee** in QA Norte. They can log in, but every module is refused
      and the UI explains they have no role yet.
- [ ] **Assign roles (owner):** give that employee `Administrador` in QA Norte and
      `Almacén` in QA Sur. As the employee: in QA Norte, can create an asset. In QA Sur,
      tires/parts are read-only and asset creation is refused. The shell role label
      changes when switching.
- [ ] **Revoke:** owner removes QA Sur from the employee. It disappears from their
      switcher. If their session was in QA Sur, the next refresh sends them to `/login`.
- [ ] **Owner safety:** owner sets their own role in QA Norte to "no role". They
      can still do everything there.
- [ ] **Isolation:** as the QA owner, no person or company of "Go Logistics" is visible
      anywhere.
- [ ] **Employees / Roles screens** no longer show "Sin rol" / 0 for people who hold roles.
