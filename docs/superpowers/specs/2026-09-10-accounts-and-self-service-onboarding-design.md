# Accounts and self-service onboarding — design

**Date:** 2026-09-10
**Status:** approved, pending implementation plan
**Scope:** subsystem 1 of 2. Subsystem 2 (directors, per-company roles) gets its own spec.

## Goal

Let the platform operator create a client and its top administrator in one act,
then let that administrator log in with no companies yet and create their own.

Today this is impossible, and the reason is structural rather than incidental.

## The hierarchy being built

```
Super super user   platform operator (the IT team)      is_platform_admin, belongs to no account
└── Account        a client organisation                exactly one owner at a time, transferable
    ├── Employees  belong to exactly ONE account
    └── Companies  belong to exactly ONE account        the tenants that hold fleet data
```

Subsystem 2 adds directors inside an account, each scoped to a subset of its
companies. This spec builds the container they will live in.

## How it works today

Three tables carry tenancy: `company`, `employee`, and the `employee_companies`
membership join. Three privilege flags gate behaviour, and **all three are global
to the person rather than scoped to a tenant**:

| Flag | Location | Gates |
|---|---|---|
| `is_account_owner` | `employee` column | `POST /api/v1/companies`, nothing else |
| `is_admin` | via `employee.role_id → role.is_admin` | `RequireAdminRole` and every module check |
| `is_platform_admin` | `employee` column, CLI-granted | `/api/v1/admin/*` |

The access token carries `employee_id`, `company_id` and `is_admin`. Every
authenticated request is scoped to the token's `company_id`.
`RequireIdentity` re-reads identity from the database on each request rather
than trusting the token, so revocations apply immediately — that behaviour is
correct and is preserved.

### The deadlock

```
login → resolveLoginCompany(default_company_id, memberships)
        no memberships → ErrNoCompanyMembership → 403, no token issued
                                    ↓
POST /api/v1/companies requires a token (Auth + RequireIdentity)
                       and is_account_owner
```

No membership means no token; no token means no company can be created; no
company means no membership. `internal/http/handler/credential.go:54-62` is
where the refusal happens.

The only way out today is `fleet-cli bootstrap --email <email> --company-id <id>`,
which creates the company and the membership out of band. That is precisely the
"company forcefully created and related" this work removes.

### What is missing

There is no entity between the platform and a company. Nothing in the schema
records that five companies belong to one client and three to another. The
platform therefore cannot answer "show me everything belonging to this client",
cannot prevent a client's administrator from being handed a company belonging to
another client, and cannot scope anything — billing, support, reporting — per
client.

## Model

### `account`

```sql
CREATE TABLE account (
    id                bigserial    PRIMARY KEY,
    name              varchar(200) NOT NULL,
    owner_employee_id bigint,
    is_active         boolean      NOT NULL DEFAULT true,
    created_at        timestamptz  NOT NULL DEFAULT now()
);
```

`owner_employee_id` is nullable because an account and its owner reference each
other. Creation is one transaction: insert the account, insert the owner
employee carrying `account_id`, then set `owner_employee_id`. It stays nullable
afterwards so ownership can be transferred without a window where the column
has no legal value.

### Account membership for people and companies

```sql
ALTER TABLE company  ADD COLUMN account_id bigint;  -- backfilled, then NOT NULL
ALTER TABLE employee ADD COLUMN account_id bigint;  -- NULL means platform staff
```

An employee with `account_id IS NULL` is platform staff. They reach tenant data
through `/api/v1/admin/*` and hold no company memberships.

### The isolation invariant, enforced by the database

A membership may only link an employee and a company **belonging to the same
account**. This is enforced structurally rather than in middleware:

```sql
ALTER TABLE employee ADD CONSTRAINT uq_employee_account UNIQUE (id, account_id);
ALTER TABLE company  ADD CONSTRAINT uq_company_account  UNIQUE (id, account_id);

ALTER TABLE employee_companies ADD COLUMN account_id bigint NOT NULL;
ALTER TABLE employee_companies
  ADD CONSTRAINT fk_ec_employee
      FOREIGN KEY (employee_id, account_id) REFERENCES employee(id, account_id),
  ADD CONSTRAINT fk_ec_company
      FOREIGN KEY (company_id, account_id)  REFERENCES company(id, account_id);
```

A cross-account membership becomes impossible to insert. Postgres rejects it —
from the API, from the CLI, from a hand-written `INSERT` during an incident.

The asymmetry this buys is exactly what the product needs: **within** an account,
memberships are granted freely, and which companies a person sees is an ordinary
data decision; **across** accounts, no mechanism exists to grant one at all.

A middleware check cannot provide that guarantee, because every future endpoint
that writes a membership would have to remember it. This codebase already
carries one privilege boundary that had to be retrofitted after the fact —
migration `000010_platform_admin` exists because `/admin/*` was gated on a
tenant's own administrator flag — and that is the failure mode a constraint
prevents.

## The company-less session

`resolveLoginCompany` currently refuses a token to an employee with no
memberships. The new rule:

| Employee state | Outcome |
|---|---|
| Has memberships | Unchanged: token scoped to `default_company_id`, or the lowest membership id |
| No memberships, has `account_id` | Token issued with **no company scope** |
| No memberships, no `account_id` | Refused, as today |

Token claims gain `account_id`. A company-less session is represented by
**omitting `company_id` from the claims entirely**, not by encoding `0`. A zero
would flow into `GetEmployeeIdentity` as a real argument and quietly return
`is_member = false`, which is the right answer for the wrong reason — and any
future query that treats `company_id` as trusted input would be scoping to a
company id that cannot exist rather than refusing outright. `Claims.CompanyID`
therefore becomes `*int64`, and code paths that need a company must handle its
absence explicitly rather than inheriting a plausible-looking default.

A company-less session passes `RequireIdentity` and fails `RequireCompanyMember`,
so it can reach nothing company-scoped. Exactly three routes are reachable:

- `GET /api/v1/me/permissions` — so a client can discover its own state
- `POST /api/v1/companies` — the reason the session exists
- `POST /auth/switch-company` — to enter the company once created

Every other authenticated route already sits behind `RequireCompanyMember` and
will refuse such a session without any change.

## Company creation becomes account-scoped

`RequireAccountOwner` changes meaning. It currently reads the global
`employee.is_account_owner` boolean. It becomes: **the caller is the owner of the
account this session belongs to** — `account.owner_employee_id = identity.EmployeeID`.

The new company takes its `account_id` from the caller's identity, never from
the request body. A request that supplies one is ignored rather than honoured;
accepting it would let an owner plant a company inside another client.

`bootstrap.Company` already seeds default roles, work-order statuses and asset
statuses, and links the creator as admin and member. That behaviour is unchanged,
so an owner who creates their first company can immediately switch into it.

**`default_company_id` is set when the caller has none.** Creating a company from
a company-less session sets the creator's `default_company_id` to it, so their
next login resolves to a real company rather than issuing another company-less
token. It is left alone when already set: silently repointing a person's default
tenant because they happened to create a company is a surprise, not a
convenience. This is the only write this spec makes to an existing employee row.

`employee.is_account_owner` becomes redundant. It is retained but unused by this
spec's gates, and its removal is left to subsystem 2 so this migration stays
additive.

## Platform-admin provisioning

Three routes in the existing `/api/v1/admin` namespace, already behind
`RequirePlatformAdmin`:

| Route | Purpose |
|---|---|
| `POST /admin/accounts` | Create account, owner employee and initial password in one transaction |
| `GET /admin/accounts` | List clients with company and employee counts |
| `POST /admin/accounts/{id}/set-owner` | Transfer ownership |

`POST /admin/accounts` accepts the initial password directly. The two-call
alternative — create, then set password — can leave an account whose owner
exists but cannot log in if the second call never happens, and "provisioned but
unusable" is a state worth designing out. The response returns the account, the
owner's employee id and email. **The password is never echoed back and never
logged**, matching how `POST /employees/{id}/set-password` already behaves.

`POST /admin/accounts/{id}/set-owner` mirrors the existing
`/admin/companies/{id}/set-owner`, and requires the new owner to already belong
to that account — which the composite constraint makes checkable in one query.

## Migration and backfill

The database currently holds one company and its employees.

1. Create `account`, add the nullable columns.
2. Insert one account named for the existing company; set every existing
   company's `account_id` to it, and every existing employee's likewise.
3. Backfill `employee_companies.account_id` from the company's account.
4. Add the unique constraints, the composite foreign keys, and the `NOT NULL`
   on `company.account_id` and `employee_companies.account_id`.
5. Set `account.owner_employee_id` to the employee currently holding
   `is_account_owner`.

Ordering matters: the composite foreign keys cannot be added before the backfill
completes, because existing rows would violate them.

**One transitional wrinkle, stated rather than silently resolved.** Employee 1 is
currently both the client's owner and a platform admin. After this migration
those are different roles and should be different people. The migration leaves
employee 1 as-is — holding both — because splitting the operator's own login
during a migration is a worse risk than the inconsistency. Creating a dedicated
internal platform-admin employee and revoking the flag from employee 1 is
follow-up work, not part of this change.

## Out of scope

- **Directors and per-company roles.** `employee.role_id` remains a single global
  foreign key. In this subsystem the account owner is administrator throughout
  their own account, which is correct for an owner, so nothing here requires the
  move. Subsystem 2 relocates `role_id` onto `employee_companies` so a director
  can be administrator of two of an account's companies and hold no access to a
  third.
- **Removing `employee.is_account_owner`.** Retained and unused so this migration
  stays additive.
- **Account-level billing, quotas or plans.** No requirement yet.

## Error handling

- A company-less session reaching a company-scoped route receives the existing
  `403` from `RequireCompanyMember`. No new error path.
- `POST /admin/accounts` with a duplicate owner email fails the existing unique
  constraint and must surface as `409`, naming the field — matching the quality
  of the current `422 validation_failed` responses rather than the current
  `409 foreign_key_violation`, which names nothing.
- A cross-account membership insert raises a foreign-key violation. It should be
  unreachable through the API; if it surfaces, it is a bug in the caller and a
  `500` is the honest answer rather than a `400` implying the client could fix it.

## Testing

- Unit tests for `resolveLoginCompany` covering all three states in the table above.
- A handler test asserting a company-less session is refused by a company-scoped
  route and admitted to the three permitted ones.
- A migration test asserting a cross-account membership insert fails.
- A test asserting `POST /api/v1/companies` ignores an `account_id` supplied in
  the body and uses the caller's.
- An end-to-end test: provision an account, log in as its owner with no
  companies, create one, switch into it, and read a company-scoped route.

## Success criteria

- A platform admin can create a client and its owner in one API call.
- That owner can log in with zero companies and receives a usable token.
- That owner can create a company and immediately work inside it.
- A membership linking an employee to a company in a different account cannot be
  created by any means, including direct SQL.
- Every existing login continues to work unchanged after the migration.
