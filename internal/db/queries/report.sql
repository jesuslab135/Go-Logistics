-- Scheduled reports: the aggregates the PDFs are built from, who receives
-- them, and the record of what was sent.
--
-- The scheduler has no request and therefore no company in its context, so
-- every query here takes company_id as an explicit argument.

-- name: ListReportCompanies :many
-- Every company the scheduler owes reports to: those of an active account.
SELECT c.id, c.name, c.logo, c.timezone, c.currency
FROM company c
JOIN account a ON a.id = c.account_id
WHERE a.is_active
ORDER BY c.id;

-- name: GetReportCompany :one
SELECT c.id, c.name, c.logo, c.timezone, c.currency
FROM company c
WHERE c.id = sqlc.arg(id);

-- name: ReportFuelByAsset :many
-- One row per asset and fuel type. fuel_entry carries no company_id; tenancy
-- comes through asset.
--
-- eff_distance and eff_volume cover only the entries whose efficiency the
-- server could derive (full tank to full tank, see fuel_recalc.go), so their
-- ratio is a real consumption figure and not distance over every litre bought.
--
-- Which day an entry belongs to: an exact UTC midnight is a date picked
-- without a time (the frontend and the bulk importer both send a bare date as
-- that day's UTC midnight), so it counts on that calendar date. Any other
-- instant counts on the day it falls on in the company's timezone. Reading a
-- bare date as an instant would put a Monday fill in the previous week in
-- every timezone west of UTC. The one-day margin on the raw range keeps the
-- index on date usable; the day comparison decides.
SELECT
    a.id                AS asset_id,
    a.name              AS asset_name,
    a.license_plate     AS license_plate,
    a.meter_unit        AS meter_unit,
    a.fuel_volume_units AS volume_unit,
    fe.fuel_type        AS fuel_type,
    count(*)::bigint                                    AS fills,
    COALESCE(sum(fe.quantity), 0)::numeric              AS volume,
    COALESCE(sum(fe.total_cost), 0)::numeric            AS cost,
    COALESCE(sum(fe.miles_traveled), 0)::numeric        AS distance,
    COALESCE(sum(fe.miles_traveled) FILTER (WHERE fe.fuel_efficiency IS NOT NULL), 0)::numeric AS eff_distance,
    COALESCE(sum(fe.quantity) FILTER (WHERE fe.fuel_efficiency IS NOT NULL), 0)::numeric       AS eff_volume
FROM fuel_entry fe
JOIN asset a ON a.id = fe.asset_id
WHERE a.company_id = sqlc.arg(company_id)
  AND fe.date >= sqlc.arg(period_start)::timestamptz - interval '1 day'
  AND fe.date <  sqlc.arg(period_end)::timestamptz + interval '1 day'
  AND CASE WHEN (fe.date AT TIME ZONE 'UTC')::time = '00:00'
           THEN (fe.date AT TIME ZONE 'UTC')::date
           ELSE (fe.date AT TIME ZONE sqlc.arg(timezone)::text)::date
      END >= (sqlc.arg(period_start)::timestamptz AT TIME ZONE sqlc.arg(timezone)::text)::date
  AND CASE WHEN (fe.date AT TIME ZONE 'UTC')::time = '00:00'
           THEN (fe.date AT TIME ZONE 'UTC')::date
           ELSE (fe.date AT TIME ZONE sqlc.arg(timezone)::text)::date
      END <  (sqlc.arg(period_end)::timestamptz AT TIME ZONE sqlc.arg(timezone)::text)::date
GROUP BY a.id, a.name, a.license_plate, a.meter_unit, a.fuel_volume_units, fe.fuel_type
ORDER BY cost DESC, a.name, fe.fuel_type;

-- name: ReportMaintenanceByAsset :many
-- One row per asset, over the jobs completed in the period.
--
-- A job is a completed service entry, or a completed work order that no
-- service entry points at. When a service entry is linked to a work order they
-- describe the same job, and counting both would double its cost; the service
-- entry is the one counted.
--
-- completed_at is optional on both tables. A service entry without one is
-- dated by when it was recorded (created_at). A work order without one is
-- dated by issued_at, not updated_at: updated_at moves on every edit, and a
-- job dated by it could land in two different monthly reports.
--
-- The total honours a manual override, as effectiveTotal does in money.go.
--
-- A job's date is placed on a day by the same rule as ReportFuelByAsset: an
-- exact UTC midnight is a bare calendar date, anything else counts on its
-- local day.
WITH job AS (
    SELECT se.asset_id,
           se.parts_subtotal,
           se.labor_subtotal,
           COALESCE(se.total_override, se.total_amount) AS total,
           (se.total_override IS NOT NULL)              AS overridden,
           COALESCE(se.completed_at, se.created_at)     AS done_at
    FROM service_entry se
    WHERE se.company_id = sqlc.arg(company_id)
      AND upper(se.status) = 'COMPLETED'
      AND COALESCE(se.completed_at, se.created_at) >= sqlc.arg(period_start)::timestamptz - interval '1 day'
      AND COALESCE(se.completed_at, se.created_at) <  sqlc.arg(period_end)::timestamptz + interval '1 day'
    UNION ALL
    SELECT wo.asset_id,
           wo.parts_subtotal,
           wo.labor_subtotal,
           COALESCE(wo.total_override, wo.total_amount) AS total,
           (wo.total_override IS NOT NULL)              AS overridden,
           COALESCE(wo.completed_at, wo.issued_at)      AS done_at
    FROM work_order wo
    JOIN work_order_status ws ON ws.id = wo.status_id
    WHERE wo.company_id = sqlc.arg(company_id)
      AND ws.marks_as_completed
      AND COALESCE(wo.completed_at, wo.issued_at) >= sqlc.arg(period_start)::timestamptz - interval '1 day'
      AND COALESCE(wo.completed_at, wo.issued_at) <  sqlc.arg(period_end)::timestamptz + interval '1 day'
      AND NOT EXISTS (SELECT 1 FROM service_entry linked WHERE linked.work_order_id = wo.id)
),
dated AS (
    SELECT j.*,
           CASE WHEN (j.done_at AT TIME ZONE 'UTC')::time = '00:00'
                THEN (j.done_at AT TIME ZONE 'UTC')::date
                ELSE (j.done_at AT TIME ZONE sqlc.arg(timezone)::text)::date
           END AS done_day
    FROM job j
)
SELECT
    a.id            AS asset_id,
    a.name          AS asset_name,
    a.license_plate AS license_plate,
    count(*)::bigint                             AS jobs,
    COALESCE(sum(j.parts_subtotal), 0)::numeric  AS parts,
    COALESCE(sum(j.labor_subtotal), 0)::numeric  AS labor,
    COALESCE(sum(j.total), 0)::numeric           AS total,
    bool_or(j.overridden)::boolean               AS has_override
FROM dated j
JOIN asset a ON a.id = j.asset_id
WHERE a.company_id = sqlc.arg(company_id)
  AND j.done_day >= (sqlc.arg(period_start)::timestamptz AT TIME ZONE sqlc.arg(timezone)::text)::date
  AND j.done_day <  (sqlc.arg(period_end)::timestamptz AT TIME ZONE sqlc.arg(timezone)::text)::date
GROUP BY a.id, a.name, a.license_plate
ORDER BY total DESC, a.name;

-- name: ListReportRecipients :many
SELECT * FROM report_recipient
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.narg(report_kind)::text IS NULL OR report_kind = sqlc.narg(report_kind)::text)
ORDER BY report_kind, lower(email);

-- name: ListActiveReportRecipientEmails :many
SELECT email FROM report_recipient
WHERE company_id = sqlc.arg(company_id)
  AND report_kind = sqlc.arg(report_kind)
  AND is_active
ORDER BY lower(email);

-- name: CreateReportRecipient :one
INSERT INTO report_recipient (company_id, report_kind, email)
VALUES (sqlc.arg(company_id), sqlc.arg(report_kind), sqlc.arg(email))
RETURNING *;

-- name: DeleteReportRecipient :execrows
-- The company predicate makes another tenant's id indistinguishable from a
-- missing one.
DELETE FROM report_recipient
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id);

-- name: ClaimReportRun :one
-- Takes ownership of one company's report for one period, or returns no row
-- when somebody else has it or it is already done.
--
-- It is a single statement on purpose. Two schedulers racing both run it, the
-- unique constraint lets exactly one insert, and the other falls into the
-- DO UPDATE whose WHERE then refuses it.
--
-- A row can be taken over in three cases: it failed and has attempts left; it
-- has been running for longer than any run takes, so its owner died; or the
-- caller forces a re-send.
--
-- attempts is returned so a later FinishReportRun can prove it is still the
-- owner: a takeover bumps attempts, so a stalled earlier owner finishing late
-- carries a value that no longer matches.
INSERT INTO report_run (company_id, report_kind, period_start, status, attempts, started_at)
VALUES (sqlc.arg(company_id), sqlc.arg(report_kind), sqlc.arg(period_start), 'running', 1, now())
ON CONFLICT ON CONSTRAINT uq_report_run_company_kind_period DO UPDATE
SET status      = 'running',
    attempts    = report_run.attempts + 1,
    error       = NULL,
    recipients  = 0,
    started_at  = now(),
    finished_at = NULL
WHERE sqlc.arg(force)::boolean
   OR (report_run.status = 'failed'
       AND report_run.attempts < sqlc.arg(max_attempts)::integer)
   OR (report_run.status = 'running'
       AND report_run.attempts < sqlc.arg(max_attempts)::integer
       AND report_run.started_at < now() - make_interval(mins => sqlc.arg(stale_minutes)::integer))
RETURNING id, attempts;

-- name: FinishReportRun :execrows
-- The attempts predicate is the ownership check: it matches only the run
-- this caller itself claimed. A stalled earlier owner, or one whose claim was
-- taken over by a stale or forced re-claim, finishes with an attempts value
-- that no longer matches the row and affects no rows.
UPDATE report_run
SET status      = sqlc.arg(status),
    error       = sqlc.narg(error),
    recipients  = sqlc.arg(recipients),
    finished_at = now()
WHERE id = sqlc.arg(id) AND attempts = sqlc.arg(attempts);

-- name: AbandonStaleReportRuns :execrows
-- A run whose process died on its last attempt can never be claimed again, so
-- it would read "running" forever. Close it as failed so the history says so.
UPDATE report_run
SET status = 'failed',
    error = 'abandoned: the process stopped before the run finished',
    finished_at = now()
WHERE status = 'running'
  AND attempts >= sqlc.arg(max_attempts)::integer
  AND started_at < now() - make_interval(mins => sqlc.arg(stale_minutes)::integer);

-- name: ListReportRuns :many
SELECT * FROM report_run
WHERE company_id = sqlc.arg(company_id)
ORDER BY started_at DESC, id DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountReportRuns :one
SELECT count(*) FROM report_run
WHERE company_id = sqlc.arg(company_id);
