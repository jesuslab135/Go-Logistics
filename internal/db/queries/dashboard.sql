-- Dashboard KPIs in one roundtrip. Django had no equivalent endpoint — the
-- frontend issued six list requests and counted client-side — so the definitions
-- below are chosen from the schema rather than ported:
--
--   total_assets        assets not archived.
--   active_work_orders  work orders whose status is not a completing status
--                       (work_order_status.marks_as_completed).
--   overdue_issues      open issues (neither resolved nor closed) that the
--                       overdue flag or a past due_date marks as late.
--   upcoming_reminders  active, un-snoozed reminders due within the caller's
--                       window (see DASHBOARD_UPCOMING_DAYS).
--   low_stock_parts     stocked inventory rows at or below their reorder point,
--                       counted only where reorder points are switched on.
--   pending_inspections submissions with failed items that have not yet produced
--                       an issue — i.e. failures still awaiting triage.

-- name: GetDashboardStats :one
SELECT
    (SELECT count(*) FROM asset a
       WHERE a.company_id = sqlc.arg(company_id)
         AND a.archived_at IS NULL)::bigint AS total_assets,

    (SELECT count(*) FROM work_order wo
       JOIN work_order_status ws ON ws.id = wo.status_id
       WHERE wo.company_id = sqlc.arg(company_id)
         AND ws.marks_as_completed = false)::bigint AS active_work_orders,

    (SELECT count(*) FROM issue i
       WHERE i.company_id = sqlc.arg(company_id)
         AND i.resolved_at IS NULL
         AND i.closed_at IS NULL
         AND (i.overdue OR (i.due_date IS NOT NULL AND i.due_date < CURRENT_DATE)))::bigint AS overdue_issues,

    (SELECT count(*) FROM service_reminder sr
       WHERE sr.company_id = sqlc.arg(company_id)
         AND sr.is_active
         AND sr.next_due_at IS NOT NULL
         AND sr.next_due_at <= now() + make_interval(days => sqlc.arg(upcoming_days)::int)
         AND (sr.snooze_until IS NULL OR sr.snooze_until <= now()))::bigint AS upcoming_reminders,

    (SELECT count(*) FROM part_inventory pi
       JOIN part p ON p.id = pi.part_id
       WHERE p.company_id = sqlc.arg(company_id)
         AND p.archived_at IS NULL
         AND pi.active
         AND pi.track_inventory
         AND pi.reorder_point_enabled
         AND pi.reorder_point IS NOT NULL
         AND pi.available_quantity <= pi.reorder_point)::bigint AS low_stock_parts,

    (SELECT count(*) FROM inspection_submission s
       WHERE s.company_id = sqlc.arg(company_id)
         AND s.failed_items_count > 0
         AND NOT EXISTS (
             SELECT 1 FROM issue i WHERE i.inspection_submission_id = s.id
         ))::bigint AS pending_inspections;
