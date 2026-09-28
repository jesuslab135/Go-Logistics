# Scheduled PDF reports by email — design

**Date:** 2026-09-28
**Status:** design approved in conversation, spec pending review
**Scope:** phase 1 of the automation initiative. No AI, no paid service.

## Goal

Two reports that users build by hand today are produced by the system and
emailed to the tenant company as a PDF, without anyone clicking:

| Report | Kind | When | Period covered |
|---|---|---|---|
| Weekly fuel | `fuel_weekly` | Monday 06:00 | previous Monday 00:00 to Sunday 24:00 |
| Monthly maintenance costs | `maintenance_monthly` | the 1st, 06:00 | previous calendar month |

The same PDFs can be downloaded on demand from the API.

Success means: a company with recipients configured receives each report once
per period, in Spanish, with its own logo, at 06:00 in its own timezone, and a
restart or a second API container never produces a second email.

## Decisions taken

| Question | Decision |
|---|---|
| Paid APIs or services? | None. Every library is MIT/BSD, SMTP uses a free mailbox |
| AI in this phase? | No |
| What runs the clock? | An in-process ticker inside the API (`robfig/cron/v3`) |
| Who receives the email? | The tenant company, at addresses it configures |
| Which mailbox sends? | Undecided. Plain SMTP settings; empty host means "do not send" |
| PDF delivery | Attached. Signed links expire in 15 minutes, too short for email |
| Are PDFs stored? | No. They are rebuilt from the data on demand |
| Language | Spanish, matching the bulk-import templates |

## Out of scope

- The AI assistant (later phase, separate spec).
- Reports for a tenant's own customers, filtered by group or asset.
- User-configurable schedules or report layouts.
- A history of sent PDF files.
- Any report other than the two above.

## What exists today

- No scheduler, no PDF generation, no email, no advisory locks.
- `fuel_entry` has **no `company_id`**; tenancy comes through `asset`. It has no
  index on `date`.
- `company.timezone` (default `America/Mexico_City`) and `company.currency`
  are stored and never read.
- Totals on work orders and service entries are server-computed; a manual
  override wins (`effectiveTotal`, `handler/money.go`).
- A service entry may point at a work order (`service_entry.work_order_id`).
- The handoff of 2026-08-26 records that timed reminders were deliberately not
  built because nothing could run them. This spec adds that missing piece.

## Prerequisite

The graceful-shutdown drain fix (`shutdownDone` in `cmd/api/main.go`) exists on
the local `main` only, not on `origin/main`. The scheduler extends that
pattern, so the fix must be on the base branch before implementation starts.
Migration `000024_inventory_non_negative` is in the same position; this spec's
migration is numbered `000025` on that assumption.

## Architecture

Four packages under `internal/platform`. Each can be tested alone.

```
scheduler ──► reports ──► (db queries)
    │            │
    │            ▼
    │          pdf  ──► storage (company logo, read only)
    ▼
  mail
```

| Package | Does | Depends on | Does not know about |
|---|---|---|---|
| `reports` | Computes the period, runs the queries, returns a `FuelWeekly` or `MaintenanceMonthly` struct | `db/gen` | PDF, email, HTTP |
| `pdf` | Renders a report struct to `[]byte` | `maroto/v2`, `storage` | database, email |
| `mail` | Sends one message with attachments | `net/smtp`, `mime/multipart` | reports |
| `scheduler` | Decides what is due, claims the run, calls the three above, records the result | all of the above | HTTP |

The HTTP handlers and the CLI call `reports` and `pdf` directly. They never go
through the scheduler.

## Data model

Migration `000025_scheduled_reports`.

```sql
CREATE TABLE report_recipient (
  id          bigserial PRIMARY KEY,
  company_id  bigint       NOT NULL,
  report_kind varchar(30)  NOT NULL,   -- 'fuel_weekly' | 'maintenance_monthly'
  email       varchar(254) NOT NULL,
  is_active   boolean      NOT NULL DEFAULT true,
  created_at  timestamptz  NOT NULL DEFAULT now()
);
-- fk to company ON DELETE CASCADE
-- UNIQUE (company_id, report_kind, lower(email))
-- CHECK report_kind IN (...)

CREATE TABLE report_run (
  id           bigserial PRIMARY KEY,
  company_id   bigint      NOT NULL,
  report_kind  varchar(30) NOT NULL,
  period_start date        NOT NULL,   -- local date, first day of the period
  status       varchar(20) NOT NULL,   -- see below
  attempts     int         NOT NULL DEFAULT 0,
  error        text,
  recipients   int         NOT NULL DEFAULT 0,
  started_at   timestamptz NOT NULL DEFAULT now(),
  finished_at  timestamptz
);
-- fk to company ON DELETE CASCADE
-- UNIQUE (company_id, report_kind, period_start)

CREATE INDEX idx_fuel_entry_asset_date ON fuel_entry (asset_id, date);
```

`report_run.status`:

| Status | Meaning | Retried? |
|---|---|---|
| `running` | claimed, in progress | yes, if older than 30 minutes (crashed run) |
| `sent` | email accepted by the SMTP server | no |
| `not_sent` | built correctly, but `SMTP_HOST` is empty | no |
| `skipped_empty` | no data in the period | no |
| `skipped_no_recipients` | no active recipient | no |
| `failed` | error; `error` holds the message | yes, until `attempts` = 3 |

The unique constraint is the send-once guarantee. The advisory lock below only
avoids wasted work.

## Scheduling

`robfig/cron` fires **one** job every 15 minutes. The job, not the cron
expression, decides what is due, because each company has its own timezone:

1. Take `pg_try_advisory_lock` on a fixed key, on a dedicated connection. If it
   is held, return.
2. For each company of an active account, and each report kind:
   - load the company's location (`time.LoadLocation`; on error use
     `America/Mexico_City` and log it);
   - compute the latest period whose due time (06:00 local on the Monday, or
     on the 1st) has passed;
   - claim it: insert the `report_run` row as `running`, or take over a
     `failed` row with `attempts < 3`, or a stale `running` row. If nothing was
     claimed, continue.
3. Run the claimed report (next section), write the final status.
4. Release the lock.

Consequences: an API that was down at 06:00 sends when it comes back; only the
latest period is considered, so a long outage does not produce a flood of old
reports.

The binary imports `time/tzdata`, because the distroless image may have no
timezone database.

**Lifecycle.** `cmd/api/main.go` starts the scheduler after `queries` and
`blobs` are built, with the signal context. On shutdown it stops accepting
ticks and `run` waits for the job in flight (bounded at 60 seconds) before
`pool.Close()`.

## Running one report

1. `reports.Build…(ctx, companyID, period)` returns the struct.
2. No rows: status `skipped_empty`, stop.
3. No active recipient: status `skipped_no_recipients`, stop.
4. `pdf.Render(report)` returns the bytes.
5. `SMTP_HOST` empty: status `not_sent`, stop.
6. `mail.Send` — one message, all recipients in `To`, PDF attached. Status
   `sent`, or `failed` with the error.

An error in one company is recorded and the loop continues with the next.

## Report content

Both PDFs: header with company logo and name, report title, period, generation
date; footer with page number. Amounts carry `company.currency`. Dates are
`dd/mm/yyyy` in the company timezone.

### Weekly fuel — "Reporte semanal de combustible"

Source: `fuel_entry` joined to `asset`, filtered by `asset.company_id` and
`fuel_entry.date` within the period.

- **Per vehicle:** number of fills, volume, total cost, distance, efficiency
  (distance / volume over full-tank series, as `fuel_efficiency` is derived
  today), average unit cost.
- **Totals:** cost for the company; volume and distance grouped by unit,
  because units are set per asset (`fuel_volume_units`, `meter_unit`).
- **Comparison:** total cost and volume against the previous week, as a
  difference and a percentage.

All entries count, including those flagged `personal`.

### Monthly maintenance — "Reporte mensual de costos de mantenimiento"

Sources, both filtered by `company_id` and `completed_at` within the period:

- `service_entry` with `status = 'COMPLETED'`;
- `work_order` whose status has `marks_as_completed = true` **and** which no
  service entry references.

That rule counts a job once: when a service entry is linked to a work order,
the service entry is the one counted.

- **Per vehicle:** number of jobs, parts, labor, total.
- **Totals:** parts, labor, total for the company.
- **Top 5** vehicles by cost.
- **Comparison** with the previous month.

Total per job is `COALESCE(total_override, total_amount)`. When an override is
set, parts plus labor may differ from the total; the report prints the three
values as stored and marks the row with an asterisk and a footnote.

## API

New RBAC module `reports`, appended to `middleware.Modules`. Administrators
get it automatically; other roles need it granted. No role backfill.

| Method and path | Action | Result |
|---|---|---|
| `GET /api/v1/reports/recipients` | read | list, filter `?report_kind=` |
| `POST /api/v1/reports/recipients` | create | `{report_kind, email}`; 422 on bad kind or email, 409 on duplicate |
| `DELETE /api/v1/reports/recipients/{id}` | delete | 204; 404 if it belongs to another company |
| `GET /api/v1/reports/fuel/weekly?week=2026-09-21` | read | `application/pdf`; `week` is any date inside the week, default last full week |
| `GET /api/v1/reports/maintenance/monthly?month=2026-08` | read | `application/pdf`; default last full month |
| `GET /api/v1/reports/runs` | read | paginated run history for the company |

Downloads return 200 with a PDF that says "Sin registros en el periodo" when
the period is empty. They do not write `report_run`.

All handlers take the company from the JWT, as every other handler does.

## CLI

```
cli reports send --company 12 --report fuel_weekly --period 2026-09-21 [--force]
```

Sends one report now. Without `--force` it respects `report_run`; with it, it
re-sends and updates the row. This is how `not_sent` runs are delivered after a
mailbox is configured.

## Configuration

Nested `SMTPConfig` and `ReportsConfig` in `internal/config`, following
`JWTConfig`. Documented in `.env.example`, `deploy/.env.production.example`
and the README table.

| Variable | Default | Notes |
|---|---|---|
| `REPORTS_ENABLED` | `false` | the scheduler does not start unless `true` |
| `REPORTS_TICK` | `*/15 * * * *` | cron expression of the single job |
| `REPORTS_SEND_HOUR` | `6` | local hour at which a period becomes due |
| `SMTP_HOST` | empty | empty means build but do not send |
| `SMTP_PORT` | `587` | STARTTLS; `465` uses implicit TLS |
| `SMTP_USER`, `SMTP_PASSWORD` | empty | |
| `SMTP_FROM` | empty | required when `SMTP_HOST` is set |

`SMTP_PASSWORD` lives only in `/opt/fleet/.env`.

## Testing

**Unit (no database):**
- period calculation across timezones, year boundaries, and the hour before
  and after the due time;
- totals, comparison percentages, division by zero when the previous period is
  empty;
- MIME message construction (headers, attachment encoding), against a fake
  SMTP server on a local listener;
- PDF render returns bytes starting with `%PDF` for a full and an empty report.

**Integration (`//go:build integration`, existing throwaway-database harness):**
- fuel query returns only the company's entries and only the period's;
- a service entry linked to a work order is counted once;
- an override total is used;
- two concurrent scheduler runs produce one `report_run` row and one send;
- a `failed` run is retried and stops after three attempts;
- recipient endpoints reject another company's ids.

CI gates unchanged: build, vet, gofmt, tests, sqlc drift, swag drift.

## To verify first in the implementation plan

These are stated as decisions above and must be confirmed against the code
before the queries are written. If one is false, the spec is corrected first.

1. `fuel_entry.miles_traveled` is an odometer difference in the asset's
   `meter_unit`, not a value converted to miles.
2. Completing a work order creates or links a service entry, so that
   "service entry wins" does not drop work orders.
3. `maroto/v2` builds under the repository's Go version with CGO disabled.

## Follow-up outside this spec

A frontend guide for the recipient screen and the download buttons, written
beside the repository after the endpoints exist.
