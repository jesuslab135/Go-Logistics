-- Archiving, and the reference checks that decide whether a delete is allowed.
--
-- The rule: a record nothing points at can be deleted; a record something
-- points at is archived instead. Hard-deleting a referenced row either breaks a
-- foreign key or silently blanks a column somebody's report depends on, and
-- neither is a thing to discover from a stack trace.
--
-- Each check names every inbound relation. There is no age exemption: a
-- two-year-old purchase order still needs its vendor's name to render, so
-- "referenced long ago" is not the same as "no longer referenced".

-- name: ArchiveAsset :one
UPDATE asset SET archived_at = sqlc.arg(archived_at), updated_at = sqlc.arg(archived_at)
WHERE id = sqlc.arg(id)::bigint AND company_id = sqlc.arg(company_id) AND archived_at IS NULL
RETURNING *;

-- name: RestoreAsset :one
UPDATE asset SET archived_at = NULL, updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)::bigint AND company_id = sqlc.arg(company_id)
RETURNING *;

-- name: AssetReferences :one
SELECT (
    (SELECT count(*) FROM vehicle WHERE vehicle.asset_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM trailer WHERE trailer.asset_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM asset_trailer_assignment WHERE asset_trailer_assignment.asset_id = sqlc.arg(id)::bigint OR asset_trailer_assignment.trailer_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM work_order WHERE work_order.asset_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM issue WHERE issue.asset_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM service_reminder WHERE service_reminder.asset_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM service_entry WHERE service_entry.asset_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM tire WHERE tire.current_vehicle_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM tire_installation WHERE tire_installation.vehicle_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM tire_mount_log WHERE tire_mount_log.vehicle_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM vehicle_axle_config WHERE vehicle_axle_config.vehicle_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM tire_assignment_request WHERE tire_assignment_request.vehicle_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM fuel_entry WHERE fuel_entry.asset_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM inspection_submission WHERE inspection_submission.asset_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM media WHERE media.asset_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM warranty WHERE warranty.asset_id = sqlc.arg(id)::bigint)
)::bigint AS references_count;

-- name: ArchivePart :one
UPDATE part SET archived_at = sqlc.arg(archived_at), updated_at = sqlc.arg(archived_at)
WHERE id = sqlc.arg(id)::bigint AND company_id = sqlc.arg(company_id) AND archived_at IS NULL
RETURNING *;

-- name: RestorePart :one
UPDATE part SET archived_at = NULL, updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)::bigint AND company_id = sqlc.arg(company_id)
RETURNING *;

-- name: PartReferences :one
SELECT (
    (SELECT count(*) FROM part_inventory WHERE part_inventory.part_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM inventory_journal_entry WHERE inventory_journal_entry.part_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM work_order_sub_line_item WHERE work_order_sub_line_item.part_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM purchase_order_line_item WHERE purchase_order_line_item.part_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM service_task_part WHERE service_task_part.part_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM service_entry_line_item WHERE service_entry_line_item.part_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM warranty WHERE warranty.part_id = sqlc.arg(id)::bigint)
)::bigint AS references_count;

-- name: ArchiveVendor :one
UPDATE vendor SET archived_at = sqlc.arg(archived_at), updated_at = sqlc.arg(archived_at)
WHERE id = sqlc.arg(id)::bigint AND company_id = sqlc.arg(company_id) AND archived_at IS NULL
RETURNING *;

-- name: RestoreVendor :one
UPDATE vendor SET archived_at = NULL, updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)::bigint AND company_id = sqlc.arg(company_id)
RETURNING *;

-- name: VendorReferences :one
SELECT (
    (SELECT count(*) FROM asset WHERE asset.lease_vendor_id = sqlc.arg(id)::bigint OR asset.loan_vendor_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM inventory_journal_entry WHERE inventory_journal_entry.vendor_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM work_order WHERE work_order.vendor_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM purchase_order WHERE purchase_order.vendor_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM service_entry WHERE service_entry.vendor_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM tire WHERE tire.vendor_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM fuel_entry WHERE fuel_entry.vendor_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM warranty WHERE warranty.provider_id = sqlc.arg(id)::bigint)
)::bigint AS references_count;

-- name: ArchiveServiceTask :one
UPDATE service_task SET archived_at = sqlc.arg(archived_at), updated_at = sqlc.arg(archived_at)
WHERE id = sqlc.arg(id)::bigint AND company_id = sqlc.arg(company_id) AND archived_at IS NULL
RETURNING *;

-- name: RestoreServiceTask :one
UPDATE service_task SET archived_at = NULL, updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)::bigint AND company_id = sqlc.arg(company_id)
RETURNING *;

-- A service task can be its own parent's child, so the self-reference counts:
-- deleting a task that other tasks hang off would orphan them.
-- name: ServiceTaskReferences :one
SELECT (
    (SELECT count(*) FROM service_task WHERE service_task.parent_task_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM service_task_part WHERE service_task_part.service_task_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM service_reminder WHERE service_reminder.service_task_id = sqlc.arg(id)::bigint)
  + (SELECT count(*) FROM service_entry_line_item WHERE service_entry_line_item.service_task_id = sqlc.arg(id)::bigint)
)::bigint AS references_count;

-- name: ArchiveInspectionForm :one
UPDATE inspection_form SET archived_at = sqlc.arg(archived_at), updated_at = sqlc.arg(archived_at)
WHERE id = sqlc.arg(id)::bigint AND company_id = sqlc.arg(company_id) AND archived_at IS NULL
RETURNING *;

-- name: RestoreInspectionForm :one
UPDATE inspection_form SET archived_at = NULL, updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)::bigint AND company_id = sqlc.arg(company_id)
RETURNING *;

-- A form's own items do not count: they are the form's content, deleted with it
-- by the cascade, not an outside reference to it. Submissions do — they are
-- evidence somebody recorded, and they outlive the form they used.
-- name: InspectionFormReferences :one
SELECT (SELECT count(*) FROM inspection_submission WHERE inspection_submission.form_id = sqlc.arg(id)::bigint) ::bigint AS references_count;
