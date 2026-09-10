# Per-company roles and directors — design

**Date:** 2026-09-10
**Status:** approved, pending implementation plan
**Scope:** subsystem 2 of 2. Subsystem 1 (accounts, self-service onboarding) is merged and deployed.

## Goal

Let one person be an administrator of one company in their account and a
limited user — or nothing at all — in another.

That is the "director" layer: inside a client organisation, a director oversees
some of its companies and not others. It cannot be expressed today.

## Why it cannot be expressed today

`employee.role_id` is a **single global foreign key**, while `role` is
company-scoped (`role.company_id`). So a person whose `role_id` points at
company A's administrator role is an administrator in **every company they
belong to**.

The system already documents this as a hazard. Migration
`000010_platform_admin` exists precisely because `/admin/*` was gated on
`is_admin`, and a tenant administrator turned out to be an administrator
everywhere they held membership — which is what made that gate reach further
than it looked.

Subsystem 1 gave each account a hard boundary. Within an account, which
companies a person can see is already an ordinary membership decision. What is
still missing is **what they may do in each one**.

## Decisions taken

| Question | Decision |
|---|---|
| Where does the role live? | On the membership — `employee_companies.role_id` |
| What happens to `employee.role_id`? | Backfilled, then **dropped** |
| Who appoints directors? | The account owner, self-service |
| A membership with no role? | Grants **nothing** — every module gate refuses |
| Who edits a company's roles? | Administrators **of that company** |
| Is the account owner special? | Yes — admin throughout their own account by virtue of ownership |

A separate `director` table was considered and rejected. `employee_companies`
already links a person to a company; a second table linking the same two things
would be two places to check and two places to drift. A director is simply a
person holding an administrator role in some of the account's companies.

## Model

### The role moves onto the membership

```sql
ALTER TABLE employee_companies ADD COLUMN role_id bigint;  -- nullable
ALTER TABLE employee DROP COLUMN role_id;                  -- after backfill
```

`role_id` is nullable because membership and permission are different facts. A
row in `employee_companies` says "this person is associated with this company".
What they may do comes only from the role. A membership carrying no role
therefore grants nothing, and forgetting to assign one **fails closed**.

### A role from the wrong company is impossible to attach

```sql
ALTER TABLE role ADD CONSTRAINT uq_role_company UNIQUE (id, company_id);

ALTER TABLE employee_companies ADD CONSTRAINT fk_ec_role_company
    FOREIGN KEY (role_id, company_id) REFERENCES role(id, company_id);
```

Attaching company A's role to a membership in company B is rejected by
Postgres — from the API, from the CLI, from a hand-written `INSERT`.

This is the same construction subsystem 1 used for cross-account membership, and
for the same reason: a rule enforced in middleware has to be remembered by every
future endpoint that writes the row, and this one will not be.

Because `employee_companies.company_id` is already part of the composite key
referencing `company(id, account_id)`, the two constraints compose: a membership
names one account, one company within it, and one role belonging to that company.

## The identity query

Subsystem 1 left `GetEmployeeIdentity` already receiving `company_id` and
already scoping membership by it. Only the role join was unscoped. Fixing that
makes the query smaller:

```sql
SELECT
    e.id,
    e.is_active,
    e.account_id,
    e.is_platform_admin,
    ec.role_id,
    ec.id IS NOT NULL                    AS is_member,
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

The `EXISTS` subquery for membership disappears — the join that resolves the
role answers membership at the same time.

**Both ownership expressions must be wrapped in `COALESCE`.** Platform staff have
`employee.account_id IS NULL`, so the `LEFT JOIN account` produces no row and
`a.owner_employee_id = e.id` evaluates to `NULL`, not `false`. In SQL
`false OR NULL` is `NULL`, and sqlc would scan that into a Go `bool` as a
failure or a zero depending on the driver path — an authorization value decided
by NULL semantics rather than by a rule. The `COALESCE` makes the absent-account
case explicitly `false`, which is correct: platform staff own no account.

### Ownership as a backstop

`is_admin` is true when the company role says so **or** when the caller owns the
account. This is deliberate.

With roles on memberships it becomes possible to remove your own administrator
role from a company you own. Without the backstop, nobody in the account could
repair it and every such mistake becomes a support request to the platform
operator. The account owner also sits *above* the companies in the product's own
mental model, not inside one of them.

The cost is one special case in the permission check. It is worth it.

## The backfill must not silently revoke access

Today a person with a global administrator role is an administrator in **every
company they belong to**. Moving `role_id` onto memberships naively would leave
them an administrator only where that particular role lives, silently stripping
access everywhere else — with no error, and no indication anything changed.

The backfill therefore preserves current effective access exactly:

For each row in `employee_companies`, attach the role **of the same name** in
that membership's company. Where no role of that name exists there, create one
copying `is_admin` and `permissions` from the employee's current global role.

`bootstrap.Company` already seeds `Administrador` and `Almacén` into every
company, so name matching resolves for the common case and the copy path is the
exception.

Two details the migration must pin down, because neither is guaranteed by the
schema:

**Role names are not unique per company.** There is no
`UNIQUE (company_id, name)` on `role` — only `fk_role_company`. So "the role of
the same name" can match more than one row. The backfill takes the **lowest
`id`** among matches, so the result is deterministic and a re-run selects the
same row. Adding the unique constraint is deliberately *not* part of this work:
it would fail if any company already holds duplicate role names, and discovering
that during this migration is worse than leaving it for its own change.

**An employee with no global role gets nothing.** `employee.role_id` is nullable
today. Where it is NULL, every membership for that employee keeps
`role_id = NULL` — which under this design grants no module access, exactly as it
does now, since a NULL role already yields empty permissions.

After this migration every person's permissions are unchanged. Only *future*
grants can differ per company — which is the entire point.

## Login resolves the role after the company

`GetEmployeeAuthByEmail` currently reads `is_admin` from the global role, which
will no longer exist.

`Verify` already resolves the session's company before issuing a token. It will
then call `GetEmployeeIdentity` for that company to obtain `is_admin`, rather
than carrying a second copy of the role logic. A company-less session — an
account owner with no companies yet — resolves `is_admin` from ownership alone.

This removes duplicated logic rather than adding any: the token's `is_admin`
claim and the request-time gate will come from the same query.

## The account owner appoints directors

Two routes, account-scoped, behind `RequireAccountOwner`:

| Route | Purpose |
|---|---|
| `GET /api/v1/account/employees` | The account's people, each with their memberships and the role held in each |
| `PUT /api/v1/account/employees/{id}/companies` | Replace that person's memberships, each carrying a role |

These are deliberately **not** under `/api/v1/admin`. That namespace is
cross-tenant and platform-admin only; this one operates strictly inside the
caller's own account.

The replace route writes the existing `membership_audit` table, exactly as the
platform-admin route does. Granting a company with an administrator role is the
highest-privilege write an account owner can make, and it must leave a trace.

It refuses a target outside the caller's account **explicitly**, with `404`,
rather than relying on the composite foreign key to raise an error. Structural
enforcement is the backstop, not the user-facing behaviour.

## Role editing needs no new code

`roles := member.Group("", middleware.RequireAdminRole())` already exists in the
router. Once `is_admin` means "administrator of the token's company", that group
means "an administrator of this company may edit this company's roles" — which is
the required behaviour, reached by changing nothing.

A director administering two companies can shape roles in exactly those two.

## What changes elsewhere

- **`CreateEmployee` / `UpdateEmployee`** take `role_id` today. It moves to the
  membership, so creating an employee and granting them access in a company
  become separate acts. This is the largest non-schema change.
- **`notification.sql`** joins `role` through `employee.role_id`; it must scope
  through the membership instead.
- **`company.sql`'s owner setup** assigns `role_id` on the employee row; it must
  assign to the membership.
- **`/me/permissions`** reports the caller's permissions for the token's company.
  It gains the role's id and name so a client can display which role is in
  effect.

## Out of scope

- **Changing what a permissions document can express.** The `Modules` registry
  and the `{module: {action: bool}}` shape are unchanged; only which role applies
  changes.
- **Cross-account anything.** Subsystem 1 made a cross-account membership
  impossible; nothing here touches that.
- **Removing `employee.is_account_owner`.** Still retained and unread, as
  subsystem 1 left it.
- **Splitting employee 1's dual client-owner / platform-admin roles.** Follow-up
  work, deliberately not bundled with a migration.

## Error handling

- A membership whose role belongs to another company raises a foreign-key
  violation. It should be unreachable through the API; if it surfaces, the caller
  is at fault and a `500` is the honest answer rather than a `400` implying the
  client could correct it.
- `PUT /account/employees/{id}/companies` naming a company outside the caller's
  account returns `404`, not `403` — the caller has no legitimate way to know
  that company exists.
- A role id that does not belong to the named company returns `422`, naming the
  field, matching how the existing validation errors report.

## Testing

This repository has no database-backed unit tests; the integration suite added in
subsystem 1 (`//go:build integration`) is where behaviour spanning Go and the
schema is proven. That is where the load-bearing tests for this change belong —
subsystem 1 shipped three defects that a green `go test ./...` did not catch.

- Unit: the permission predicate, including that ownership grants admin and that
  a nil role grants nothing.
- Unit: role resolution for a company-less session.
- Integration: a person granted admin in company A and a read-only role in
  company B is refused a write in B and permitted it in A. **This is the
  behaviour the whole subsystem exists to deliver and it must be proven
  end to end.**
- Integration: attaching a role from another company fails.
- Integration: an account owner stripped of their explicit role retains admin.
- Migration: after the backfill, every employee's effective permissions in every
  company they belong to are identical to before. This is the assertion that
  protects live customers.

## Success criteria

- A person can hold an administrator role in one company of their account and a
  limited role in another.
- A membership carrying no role grants no module access anywhere.
- An account owner cannot lock themselves out of a company they own.
- A role belonging to another company cannot be attached to a membership by any
  means, including direct SQL.
- After the migration, no existing user's effective permissions change.
- `employee.role_id` no longer exists.
