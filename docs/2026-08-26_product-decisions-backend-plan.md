# Backend implementation plan — product decisions (2026-08-26)

Scope: the backend half of the decision set reviewed against
`PENDING-DECISIONS-TEAM-REVIEW.md` (compiled 2026-08-17) and
`backend-engineer-task-todo.md` (validated 2026-08-20). Frontend-only
decisions (A1–A4, B1, B5, C8) are out of scope here.

Validated against `main` @ `ce7d063`. Where the Django original
(`../Fleet-Personal-Backend`) had an answer, this plan ports it rather than
inventing one; where it did not, the decision is called out as new design and
justified.

---

## 0. Cross-cutting policies

These two answers govern every item below.

### 0.1 Migrations are backfill-safe and additive

A live deployment exists (`go-logistics.jesuslab135.com`). No migration in
this plan drops a column or a table. Columns being retired are left in place,
stopped being written, and documented as deprecated — a later cleanup
migration removes them once no deployed client reads them. Every new NOT NULL
column ships with a DEFAULT or a backfill in the same migration.

**Why:** the cost of keeping a dead column for one release is nil; the cost of
dropping one that a deployed client still reads is an outage.

### 0.2 Breaking route changes ship in one batch per phase, and fail loudly

C1, C2 and C3 each remove write verbs or change what a default list returns.
The frontend is the only consumer and is in active development, so these ship
as real breaks rather than behind a compatibility flag — but a removed verb
returns **405 with a documented error code** (`route_retired`, naming the
replacement), never a bare 404.

**Why:** a 404 is indistinguishable from a typo'd path. A 405 naming its
replacement turns a frontend bug report into a one-line fix.

---

## 1. B8 — Upload privacy (decision A: classify by type)

**Status:** nothing exists. `MinIO.setPublicReadPolicy` grants `s3:GetObject`
to `Principal:*` over the whole bucket at startup
(`internal/platform/storage/minio.go:135`). `PresignedGetURL` is implemented
and has zero callers. `purpose` narrows the MIME allowlist only.

### Decisions

| Question | Decision | Why |
|---|---|---|
| Mechanism | Presigned GET URLs, not an auth-proxy endpoint | `PresignedGetURL` already exists; a proxy would stream every byte through the API container, which is a single instance today |
| Bucket layout | Two buckets: `fleet-public`, `fleet-private` | A bucket policy is one object; per-object ACLs are N objects to get wrong. The split is static per purpose, so it needs no per-object decision |
| What is public | asset photos, company logo | Branding and vehicle imagery render in plain `<img>` across list screens; signing them buys nothing and costs a signature per row |
| What is private | receipts, signatures, inspection photos, documents, generic media | The decision's own classification |
| URL expiry | 15 minutes | Long enough for a page session, short enough that a leaked URL in a log or a screenshot is dead by the time anyone finds it |
| Storage in DB | object **key**, not absolute URL | A signed URL cannot be persisted — it expires. This is the change that makes the rest possible |
| Signing on lists | signed at render for private objects, per row | Rejected the alternative (sign on demand per item): it needs a second round trip per image and every consumer screen would have to implement it |

### Work

1. **Migration** `000010_upload_object_keys`: add a `*_key varchar(200)` column
   beside each of the five existing URL columns; backfill by stripping the
   public base prefix from the stored URL. Old URL columns keep being written
   for one release (policy 0.1).
2. **`storage`**: `MinIO` gains a second bucket handle. `Put` takes a
   visibility class derived from `purpose`. `URLFor(key, visibility)` returns
   either the permanent public URL or a 15-minute presigned URL.
3. **`upload.go`**: extend the purpose vocabulary from `photo|document|generic`
   to carry a class — `asset_photo`, `company_logo` (public); `receipt`,
   `signature`, `inspection_photo`, `document`, `media` (private). An
   unrecognised purpose stays a 400, as today.
4. **Response DTOs**: every DTO exposing a file URL resolves it through
   `URLFor` at serialization time. Add `expires_at` alongside private URLs so
   the client knows when to re-fetch rather than discovering it via a 403.
5. **`storage_cleanup.go`**: `reclaimObjects` learns which bucket a key lives in.
6. **Deployment**: create the second bucket, move existing private-class
   objects into it, drop the anonymous-read policy from `fleet-private`.

**Effort:** large (~3–4 days). **Blocks:** B4 inspections, fuel receipts.

---

## 2. B9 — Platform admin (decision A: explicit identity)

**Status:** `/admin/*` is gated on `RequireAdminRole` = `employee.is_admin`.
Because `employee.role_id` is a single global FK, any company admin reaches
every cross-tenant route. `admin_employee.go`'s `validateGrantableCompanies`
delta constraint exists to contain that; it is a workaround, not a fix.

### Decisions

| Question | Decision | Why |
|---|---|---|
| Where the identity lives | `employee.is_platform_admin boolean NOT NULL DEFAULT false` | A separate table buys nothing while the flag is per-person and boolean. Revisit if platform roles ever need granularity |
| How it is granted | CLI only (`cmd/cli`), no API route | The population is a handful of internal staff. A self-service surface is an attack surface with no user |
| Scope of the claim | `/admin/*` only — it does **not** bypass company scoping on ordinary routes | A platform admin acting inside a tenant should still act *as* that tenant, so ordinary routes stay honest and auditable |
| Fate of `RequireAdminRole` on `/admin/*` | Replaced by `RequirePlatformAdmin` on all 7 routes | Keeping both would mean the delta constraint has to stay forever |
| Membership audit | New `membership_audit` table, append-only | Decision B9-A names audited grants explicitly |

**Consequence, stated plainly:** the day this ships, company admins lose the
admin company pages they can use today. That is the point of the decision, but
it is a visible regression for them and the frontend needs the same release.
`validateGrantableCompanies` can then be deleted — the delta constraint is only
needed while non-platform admins reach the namespace.

### Work

1. **Migration** `000011_platform_admin`: `employee.is_platform_admin`;
   `membership_audit (id, actor_employee_id, subject_employee_id, company_id,
   action, occurred_at)` where action ∈ `granted|revoked`.
2. **`auth`**: `is_platform_admin` into the JWT claims and `Identity`.
3. **`middleware`**: `RequirePlatformAdmin()`, applied to the 7 `/admin/*`
   routes in `router.go:302-310`.
4. **`admin_employee.go`**: write `membership_audit` rows inside the existing
   replace transaction; delete `validateGrantableCompanies`.
5. **`cmd/cli`**: `grant-platform-admin` / `revoke-platform-admin`.

**Effort:** medium (~2 days).

---

## 3. B2 — Money math (decision A: server-authoritative)

**Status:** pure pass-through in all three handlers. **The Django original
never computed these either** — `total_amount`, `subtotal`, `tax_*` are plain
`DecimalField(default=0)` with no `save()` override and no calculation service.
The only computed values anywhere are three read-only serializer properties on
work orders (`total_labor`, `total_parts`, `grand_total`), each a bare
`sum(quantity × unit_cost)` over sub-line-items that ignores markup, discount
and tax entirely.

**So there is no formula to port.** This is new design, and the plan says so
rather than pretending otherwise. What follows is grounded in the field layout
the schema already has and in the model it was copied from —
`work_order_model.py:181` comments the markup block as *"Fleetio soporta markup
sobre partes y labor"*.

### The formula

Per surface, because the three surfaces do not have the same fields:

```
parts_subtotal  = Σ(quantity × unit_cost) over PART sub-line-items
labor_subtotal  = Σ(quantity × unit_cost) over LABOR sub-line-items

# markup: work orders only — purchase orders and service entries have no markup columns
parts_marked    = parts_subtotal + resolve(parts_markup_type, parts_markup, parts_markup_percentage, parts_subtotal)
labor_marked    = labor_subtotal + resolve(labor_markup_type, labor_markup, labor_markup_percentage, labor_subtotal)

subtotal        = parts_marked + labor_marked        # work orders
subtotal        = Σ line_item.subtotal               # purchase orders, service entries

discount_amount = resolve(discount_type, discount, discount_percentage, subtotal)
net             = subtotal - discount_amount

tax_1_amount    = resolve(tax_1_type, tax_1, tax_1_percentage, net)
tax_2_amount    = resolve(tax_2_type, tax_2, tax_2_percentage, net)   # on net, NOT on net + tax_1

shipping        = shipping                            # purchase orders only, not taxed
total_amount    = net + tax_1_amount + tax_2_amount + shipping

resolve(type, fixed, pct, base) = fixed             if type == FIXED
                                = base × pct / 100  if type == PERCENTAGE
```

| Question | Decision | Why |
|---|---|---|
| Tax before or after discount? | **After** — tax applies to the net | `company.currency` defaults to `MXN`; Mexican IVA is charged on the value after commercial discounts. Also matches Fleetio |
| Markup per line or aggregate? | **Aggregate**, per category | The columns are `parts_markup`/`labor_markup` on the parent, not on the line. Per-line markup would need columns that do not exist |
| tax_1 + tax_2 compounding? | **No** — both on the same net base | They are two jurisdictions (IVA + retención), not a tax on a tax |
| Is shipping taxed? | **No** | It sits after the tax lines in the schema and is PO-only; taxing freight would need its own rate column |
| Rounding | Half-up to 2 dp at each named component, not only at the total | Every component is a persisted `numeric(12,2)`; rounding only at the end would store values that do not re-add to the total |
| Work order vs service entry | Same formula, markup terms absent on service entries | Service entries genuinely have no markup columns in either codebase |
| Frozen totals | **Yes** — recomputation stops once a PO is `CLOSED` or a work order reaches a status flagged closed | A tax-rate change must not silently re-total last quarter's documents |
| Override | Whole-document `total_amount` only, audited | Per-field overrides multiply the audit surface for a case nobody has asked for |
| Override survives edits? | **No** — a line-item write clears the override and recomputes | A stale override silently misstating a total is the failure this decision exists to prevent. Clearing is visible; drifting is not |

### Work

1. **Migration** `000012_money`: add `discount_percentage numeric(5,2) NOT NULL
   DEFAULT 0` to `work_order` and `service_entry` — only `purchase_order` has it
   today, so `discount_type='PERCENTAGE'` is currently unrepresentable on two of
   the three surfaces; add `total_override numeric(12,2) NULL`,
   `total_override_reason text`, `total_override_by_id bigint`,
   `total_override_at timestamptz` to all three.
2. **New package** `internal/domain/money`: `resolve`, plus `WorkOrderTotals`,
   `PurchaseOrderTotals`, `ServiceEntryTotals`. Pure functions over
   `decimal.Decimal`, table-driven tests — this is the one place the formula lives.
3. **Recompute triggers**: parent create/update, and every line-item and
   sub-line-item write, inside the same transaction as the write that caused it.
4. **DTOs**: the ~18 money fields per surface become read-only in responses;
   inputs keep only the *policy* fields (rates, types, percentages, shipping).
   `POST /{resource}/{id}/override-total` takes `{amount, reason}`.

**Effort:** large (~4 days). This is the item most worth reviewing before
building, because it is design rather than a port.

---

## 4. B3 — Purchase-order workflow (decision B: permission-gated transitions)

**Status:** `PUT /purchase-orders/{id}` is the only write. Good news the audit
surfaced: `purchase_order.state` is already a real column with the eight values
`DRAFT, PENDING_APPROVAL, REJECTED, APPROVED, PURCHASED, RECEIVED_PARTIAL,
RECEIVED_FULL, CLOSED`, already validated as a filter enum
(`list_filters.go:886`). So this is a state machine over existing data.

### Decisions

| Question | Decision | Why |
|---|---|---|
| Permission model | New `purchase_orders` module in `middleware.Modules` with a custom `approve` action | POs currently live under the `inventory` module, so `purchase_orders.approve` as written in the decision names a module that does not exist. `RequireAction("tire_approvals","approve")` already proves the mechanism |
| Who may submit | `purchase_orders.update` | Submitting is not privileged; it is asking |
| Who may approve/reject | `purchase_orders.approve` | The decision |
| Who may purchase/receive/close | `purchase_orders.update` | Receiving goods is warehouse work, not authorization |
| Reject terminal? | **No** — `REJECTED` → `DRAFT` via an explicit revise action, rejection reason required | The decision says return to review; procurement rejection is revise-and-resubmit |
| Self-approval | **Allowed** | No separation-of-duties rule exists in the product, and inventing one would block single-admin tenants from ever approving anything |
| Approval threshold | **None** | No amount-threshold column exists; adding one is a product feature, not part of this decision |
| Actor | From the JWT, always | The decision, and it matches `issues`' resolve/reopen/close |

Legal transitions:

```
DRAFT            → PENDING_APPROVAL   (submit)
PENDING_APPROVAL → APPROVED           (approve)
PENDING_APPROVAL → REJECTED           (reject, reason required)
REJECTED         → DRAFT              (revise)
APPROVED         → PURCHASED          (purchase)
PURCHASED        → RECEIVED_PARTIAL   (receive-partial)
PURCHASED        → RECEIVED_FULL      (receive-full)
RECEIVED_PARTIAL → RECEIVED_FULL      (receive-full)
RECEIVED_FULL    → CLOSED             (close)
```

Anything else is 409 `invalid_transition`, naming the current state.

### Work

1. **Migration** `000013_po_workflow`: `rejection_reason text NOT NULL DEFAULT ''`;
   `purchase_order_status_log (id, purchase_order_id, from_state, to_state,
   actor_employee_id, actor_type, changed_at)` — same shape as C1's work-order
   log, for the same reason.
2. **`middleware.Modules`**: add `purchase_orders`; move the PO routes out of the
   `inventory` module group.
3. **8 routes**: `/submit`, `/approve`, `/reject`, `/revise`, `/purchase`,
   `/receive-partial`, `/receive-full`, `/close`. Each validates the transition,
   stamps state + timestamp + actor, and writes the log in one transaction.
4. `PUT /purchase-orders/{id}` stops accepting `state` and the workflow
   timestamps — they become response-only.

**Effort:** medium (~2 days).

---

## 5. C1 — Work-order status audit (decision A: append-only, actor nullable)

**Status:** `work_order_status_log` is `(id, work_order_id, status_id,
changed_at)` — no actor. It is registered as a **writable nested CRUD resource**
(`router.go:221`), so status changes and log rows are unrelated writes: a work
order can change status with no log, and past rows can be edited or deleted.

**The Django original is explicit here** — `work_order_model.py:458` documents
the model as *"Historial inmutable de transiciones... Se alimenta automáticamente
mediante signals"*, and `signals.py` creates a row on `post_save` when the record
is created or `last_log.status != instance.status`. **The writable CRUD route is
a port regression, not a faithful copy.** This item mostly restores intended
behaviour; only the actor columns are new.

### Decisions

| Question | Decision | Why |
|---|---|---|
| Actor columns | `actor_employee_id bigint NULL` + `actor_type varchar(20)` | The decision |
| `actor_type` vocabulary | `employee` \| `system` | Two producers exist. A third value can be added without a migration |
| Existing rows | Backfilled `actor_type='system'`, actor null | They were written by the port's CRUD route with no actor recorded; claiming a human did it would be a fabrication |
| Route verbs | GET only; POST/PUT/DELETE return 405 `route_retired` | Restores Django's immutability |
| Write trigger | Inside the work-order create/update transaction, when `status_id` changes | Ports `signals.py` faithfully, minus the signal indirection |

### Work

1. **Migration** `000014_status_log_actor`: two columns, backfill
   `actor_type='system'`, then set NOT NULL on `actor_type`.
2. **`work_order.go`**: `Create` logs the initial status; `Update` logs when
   `status_id` differs from the stored row. Both inside the existing transaction.
3. **`router.go:221`**: swap `NewNestedHandler` for a read-only registration.

**Effort:** small-medium (~1 day).

---

## 6. C3 — Inventory journal (decision A: append-only with reversals)

**Status:** full CRUD, and `Create` never touches
`part_inventory.available_quantity`.

**What the Django source shows:** the ordinary journal endpoint does not move
stock either — but the Excel import action does, and it is unambiguous about the
intended semantics (`inventory_journal_entry_viewset.py:378-391`):

```python
prev_qty = pld.available_quantity
pld.available_quantity += qty
pld.save(update_fields=['available_quantity', 'available_quantity_updated_at', ...])
InventoryJournalEntry.objects.create(
    previous_quantity=prev_qty, adjustment_quantity=qty,
    current_quantity=pld.available_quantity, ...)
```

The `previous_quantity` / `adjustment_quantity` / `current_quantity` triplet is
meaningless unless the entry is what moved the stock. So making `Create`
authoritative is a port of the import path's semantics, not new design.

### Decisions

| Question | Decision | Why |
|---|---|---|
| Is the journal the source of truth? | **Yes** — `Create` moves `part_inventory.available_quantity` in the same transaction | The three quantity columns only mean something under this reading |
| Edit / delete | Retired (405 `route_retired`) | The decision |
| Reversal | `POST /inventory-journal-entries/{id}/reverse`, writes an offsetting entry + the stock move in one transaction | The decision |
| Who may reverse | `inventory.update` | Reversing is an ordinary correction, not an approval |
| Double reversal | 409 — an entry already reversed cannot be reversed again | Otherwise the ledger oscillates |
| Existing desynced rows | **Not** retroactively reconciled; a one-off CLI reconcile command reports drift per part | Silently rewriting historical stock levels to match a ledger that was never authoritative would destroy the only evidence of what actually happened |

### Work

1. **Migration** `000015_journal_reversal`: `reversal_of_id bigint NULL` (unique,
   self-FK) — uniqueness is what enforces "reverse once".
2. **`inventory_journal_entry.go`**: `Create` takes a transaction, reads
   `part_inventory` `FOR UPDATE`, writes previous/current from the actual row
   rather than trusting the client, and updates the stock.
3. `Update`/`Delete` retired; `Reverse` added.
4. **`cmd/cli`**: `inventory-drift-report`.

**Effort:** medium (~2 days).

---

## 7. C2 — Archive / soft-delete (decision A)

**Status:** `archived_at` exists on `asset`, `part`, `vendor`, `service_task`,
`inspection_form` as a plain writable timestamp. No archive/restore routes, no
reference checks, archived rows appear in every list.

### Decisions

| Question | Decision | Why |
|---|---|---|
| "Referenced" means | any inbound FK row, no age exemption | An age rule would need a policy per relation, and a two-year-old PO still needs its vendor's name to render |
| DELETE on a referenced row | 409 `resource_referenced`, naming the blocking relation | Better than the current silent FK break |
| DELETE on an unreferenced row | proceeds, as today | The decision |
| Archived in lists | excluded by default; `include_archived=true` opts in | The decision |
| Archived as a pick target | rejected on write with 422 | Archiving that still lets new references form achieves nothing |
| Who may restore | same permission as update on the resource | Archiving is not privileged; it is editorial |
| `archived_at` as a writable field | retired from create/update DTOs | Two ways to archive is one too many |

### Work

1. No migration.
2. **`crud`**: an optional archivable behaviour — `POST /{resource}/{id}/archive`
   and `/restore`, an `include_archived` list filter, and a pre-delete reference
   check driven by a per-resource table of inbound relations.
3. Apply to the five resources; drop `archived_at` from their write DTOs.

**Effort:** medium (~2 days), most of it the reference-check table.

---

## 8. C4 — Primary photo (decision A)

**Status:** `fuel_photo.is_primary` is a bare boolean, no index, no transaction.
Blob GC already exists (`storage_cleanup.go`).

### Decisions

| Question | Decision | Why |
|---|---|---|
| Invariant | zero photos valid; if any exist, at most one primary | The decision's own wording |
| Enforcement | partial unique index `(entry_id) WHERE is_primary`, plus a set-primary write that clears siblings in one transaction | The index makes the invariant true regardless of code path |
| Deleting the primary | **no auto-promotion** | Auto-promotion silently designates a photo nobody chose; "no primary" is a state the UI must handle anyway, because zero photos is valid |
| Scope | `fuel_photo` only for now | It is the only collection with an `is_primary` column |

### Work

1. **Migration** `000016_fuel_photo_primary`: resolve existing duplicates (keep
   lowest id), then `CREATE UNIQUE INDEX ... WHERE is_primary`.
2. **`fuel_photo.go`**: `POST /fuel-entries/{id}/photos/{photo_id}/set-primary`
   in a transaction; `is_primary` retired from create/update.

**Effort:** small (~half a day).

---

## 9. C9 — Notifications (decision A: in-app, expand producers)

**Status:** table + list + mark-read. Two producers, both tire-related.

### Decisions

| Question | Decision | Why |
|---|---|---|
| Retention | 180 days, hard delete | Long enough to cover a quarter-end look-back; the table has no other consumer |
| Cleanup mechanism | opportunistic sweep on the list endpoint, capped per call — **not** a cron | The deployment is a single API container with no scheduler; adding one for a delete is disproportionate |
| Service-reminder-due producer | **Deferred to a follow-up**, because it genuinely needs a scheduler | Honest sequencing beats a fake "due" check that only fires when someone happens to load a page |
| Work-order-assignment producer | Ships now — assignee only, not watchers | Notifying watchers on assignment turns the bell into noise for anyone tracking a busy asset |
| `url` validation | must match `^/[a-z0-9/_-]*$` — internal paths only | The column is unconstrained `text` today; a stored absolute URL is a phishing vector inside the bell |

### Work

1. No migration (the existing index already covers the retention predicate).
2. Producer in the work-order assignment path.
3. `url` validation at the one write site.
4. Sweep in `NotificationHandler.List`.

**Effort:** small-medium (~1 day). The service-reminder producer is a separate
item, gated on a scheduler decision.

---

## 10. C7 — Trailer vocabulary (decision A)

**Status:** `classification` / `classification_2` are `varchar(100)` free text.
**Django is the same** — `blank=True`, no choices — while sibling fields
`trailer_type` and `financing` *do* carry choices. So the free text was a gap in
the original too, not a porting loss.

Confirmed while auditing: **there is no `protocols` column and no trailer meter
fields in the schema at all.** That half of decision C7 is entirely frontend
cleanup — nothing to remove here.

### Decisions

| Question | Decision | Why |
|---|---|---|
| Enum or catalog? | **Company-scoped catalog table** | The values are ops vocabulary nobody has written down, and they differ per fleet. Every comparable list in this system (`fuel_type`, `measurement_unit`, `asset_status`, `issue_priority`) is already a company-scoped catalog — this follows the established pattern instead of inventing a second one |
| Existing free-text values | migrated into the catalog per company (distinct non-empty values become rows), then the column becomes an FK | Nothing is lost and no tenant has to re-enter their vocabulary |
| Unknown value on write | 422 once migrated | The point of the decision |
| Seeded defaults | **None** | Seeding a guessed vocabulary would put words in ops' mouths; the migration seeds each company from its own data |

### Work

1. **Migration** `000017_trailer_classification`:
   `trailer_classification (id, company_id, name, position)`; seed from distinct
   existing values per company; add `classification_id` /
   `classification_2_id`, backfill, leave the varchar columns in place per
   policy 0.1.
2. CRUD for the catalog under Settings, following `fuel_type`.
3. Trailer write path validates the FKs.

**Effort:** small-medium (~1 day).

---

## 11. C6 — Custom fields (decision A: company-scoped definitions)

**Status:** `json.RawMessage` passthrough on asset, employee, issue, part,
purchase order. No definitions, no validation, no index.

### Decisions

| Question | Decision | Why |
|---|---|---|
| Types | `text`, `number`, `date`, `boolean`, `select` (single) | Covers what a fleet actually records. Multi-select is deferred — it changes the storage shape from scalar to array and doubles the filter work |
| Scope of a definition | per company **and** per resource | A field meaningful on an asset is rarely meaningful on a purchase order; one global set would show every field on every form |
| Required fields | supported, enforced on create and update | Without it the feature cannot express "every truck must have a cost centre", which is the main thing people want it for |
| Existing unmatched JSON | preserved, never rejected; reported by a CLI audit | Rejecting on next write would make previously-saved records uneditable — the user cannot fix data the form no longer shows |
| Filtering | equality only, GIN index on the `custom_fields` column | Ranges need typed extraction per definition; equality covers the stated need |

### Work

1. **Migration** `000018_custom_field_definitions`:
   `custom_field_definition (id, company_id, resource, key, label, type,
   required, options jsonb, position)`, unique on `(company_id, resource, key)`;
   GIN index on each `custom_fields` column.
2. **New package** `internal/domain/customfield`: validation of a
   `json.RawMessage` payload against a company's definitions for a resource.
3. CRUD for definitions (admin-gated).
4. Validation wired into the five resources' create/update paths.
5. `custom_fields.<key>=<value>` list filter.

**Effort:** large (~3 days). Nothing else depends on it — last.

---

## 12. A5 follow-up — company-scoped employee list

**Status:** `GET /admin/employees` calls `ListAllEmployees` with pagination only.

**Decision:** add a `company_id` query filter joined through
`employee_companies` (membership, not `default_company_id` — membership is what
the page means by "belongs to"). Rejected the separate
`GET /admin/companies/{id}/employees` route: it would return the same DTO from a
second path, and the existing route already carries the pagination and admin
gate.

**Work:** one query with an optional join, one param, one swag annotation
(`limit`/`offset` shape per the standing convention).

**Effort:** ~2 hours.

---

## Sequencing

| Phase | Items | Rationale |
|---|---|---|
| 1 | **B8**, **B9** | B8 is a live cross-tenant exposure and blocks two committed features. B9 defines the identity B8's private-object authorization checks against, and closes the "any company admin is a platform admin" hole |
| 2 | **C1**, **C3** | Both restore intended Django semantics the port lost; they share a pattern (write inside the causing transaction, retire the write verbs) and one breaking-change release |
| 3 | **B3**, **A5 follow-up** | B3 needs B9's permission work in place; A5 is small enough to ride along |
| 4 | **B2** | Independent, but the largest design risk — better after the audit/transaction patterns from phase 2 are established |
| 5 | **C2**, **C4**, **C7**, **C9** | Independent polish, any order |
| 6 | **C6** | Largest, nothing depends on it |

Phases 2 and 3 each carry a breaking change and need a coordinated frontend
release; phases 1, 4, 5 and 6 are additive.

## Items closed with no backend work

- **C5** — already implemented. `fuel_recalc.go` derives distance and efficiency
  per (asset, fuel type) series and recomputes the tail on backdated insert,
  edit and delete. `state` stays free text and GPS stays uncaptured, which is
  what decision C matches.
- **B6** — no `/auth/signup` exists; decision A is to keep it that way. Item 6 of
  the 2026-08-20 task list can be closed as *won't do*.
- **B7** — `middleware/rbac.go` already enforces role permissions only; groups
  gate nothing.
- **A1–A4, B1, B5, C8** — frontend.
- **C7's "drop Protocols and meters"** — no such columns exist here.
