-- Line-item sums, and the writes that store a recomputed set of totals.
--
-- The sums are done in SQL rather than by loading the lines: a document's total
-- must not depend on how many rows a page happened to return, and the recompute
-- runs inside the transaction that changed a line, where a second round trip per
-- line would be the expensive part.

-- SumWorkOrderLineCosts splits a work order's sub-line-items by what they are.
-- The parts/labor split is what the two markup rates apply to separately, so it
-- has to come out of the query rather than being reconstructed afterwards.
-- name: SumWorkOrderLineCosts :one
SELECT
    COALESCE(SUM(sli.quantity * sli.unit_cost) FILTER (WHERE sli.item_type = 'PART'), 0)::numeric(14,2)  AS parts_subtotal,
    COALESCE(SUM(sli.quantity * sli.unit_cost) FILTER (WHERE sli.item_type = 'LABOR'), 0)::numeric(14,2) AS labor_subtotal
FROM work_order_sub_line_item sli
JOIN work_order_line_item li ON li.id = sli.line_item_id
WHERE li.work_order_id = sqlc.arg(work_order_id);

-- SumServiceEntryLineCosts uses the line's own parts_cost/labor_cost columns,
-- which already carry the split; line_item_type describes what the line is for
-- (a task, a part, labour), not which bucket its money belongs in.
-- name: SumServiceEntryLineCosts :one
SELECT
    COALESCE(SUM(parts_cost), 0)::numeric(14,2) AS parts_subtotal,
    COALESCE(SUM(labor_cost), 0)::numeric(14,2) AS labor_subtotal
FROM service_entry_line_item
WHERE service_entry_id = sqlc.arg(service_entry_id);

-- A purchase order has one undifferentiated subtotal: it buys parts, and has no
-- markup columns to apply per category.
-- name: SumPurchaseOrderLineCosts :one
SELECT COALESCE(SUM(quantity * unit_cost), 0)::numeric(14,2) AS subtotal
FROM purchase_order_line_item
WHERE purchase_order_id = sqlc.arg(purchase_order_id);

-- name: StoreWorkOrderTotals :exec
UPDATE work_order SET
    parts_subtotal = sqlc.arg(parts_subtotal),
    labor_subtotal = sqlc.arg(labor_subtotal),
    subtotal       = sqlc.arg(subtotal),
    total_amount   = sqlc.arg(total_amount),
    updated_at     = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: StoreServiceEntryTotals :exec
UPDATE service_entry SET
    parts_subtotal = sqlc.arg(parts_subtotal),
    labor_subtotal = sqlc.arg(labor_subtotal),
    subtotal       = sqlc.arg(subtotal),
    total_amount   = sqlc.arg(total_amount),
    updated_at     = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: StorePurchaseOrderTotals :exec
UPDATE purchase_order SET
    subtotal     = sqlc.arg(subtotal),
    total_amount = sqlc.arg(total_amount),
    updated_at   = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- The override writes. Clearing is a separate statement from setting because
-- clearing happens implicitly on any line-item edit, where there is no actor to
-- record and nothing to say.

-- name: SetWorkOrderTotalOverride :one
UPDATE work_order SET
    total_override = sqlc.arg(total_override), total_override_reason = sqlc.arg(reason),
    total_override_by_id = sqlc.narg(actor), total_override_at = sqlc.arg(at), updated_at = sqlc.arg(at)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id)
RETURNING *;

-- name: ClearWorkOrderTotalOverride :exec
UPDATE work_order SET
    total_override = NULL, total_override_reason = '', total_override_by_id = NULL, total_override_at = NULL
WHERE id = sqlc.arg(id) AND total_override IS NOT NULL;

-- name: SetPurchaseOrderTotalOverride :one
UPDATE purchase_order SET
    total_override = sqlc.arg(total_override), total_override_reason = sqlc.arg(reason),
    total_override_by_id = sqlc.narg(actor), total_override_at = sqlc.arg(at), updated_at = sqlc.arg(at)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id)
RETURNING *;

-- name: ClearPurchaseOrderTotalOverride :exec
UPDATE purchase_order SET
    total_override = NULL, total_override_reason = '', total_override_by_id = NULL, total_override_at = NULL
WHERE id = sqlc.arg(id) AND total_override IS NOT NULL;

-- name: SetServiceEntryTotalOverride :one
UPDATE service_entry SET
    total_override = sqlc.arg(total_override), total_override_reason = sqlc.arg(reason),
    total_override_by_id = sqlc.narg(actor), total_override_at = sqlc.arg(at), updated_at = sqlc.arg(at)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id)
RETURNING *;

-- name: ClearServiceEntryTotalOverride :exec
UPDATE service_entry SET
    total_override = NULL, total_override_reason = '', total_override_by_id = NULL, total_override_at = NULL
WHERE id = sqlc.arg(id) AND total_override IS NOT NULL;

-- Parent lookups for a recompute triggered from a child row, so a line-item
-- write can find the document it belongs to without the handler carrying it.

-- name: GetWorkOrderIDForLineItem :one
SELECT work_order_id FROM work_order_line_item WHERE id = sqlc.arg(id);

-- name: GetWorkOrderIDForSubLineItem :one
SELECT li.work_order_id FROM work_order_sub_line_item sli
JOIN work_order_line_item li ON li.id = sli.line_item_id
WHERE sli.id = sqlc.arg(id);

-- Unscoped single-document reads for the recompute. The company scope is
-- already established by the write that triggered it — a recompute reached
-- through a line item the caller was allowed to write must not fail because the
-- recompute itself forgot to carry the tenant.

-- name: GetWorkOrderByID :one
SELECT * FROM work_order WHERE id = sqlc.arg(id);

-- name: GetServiceEntryByID :one
SELECT * FROM service_entry WHERE id = sqlc.arg(id);

-- name: GetPurchaseOrderByID :one
SELECT * FROM purchase_order WHERE id = sqlc.arg(id);

-- Line-level derivations. A line's own money columns are computed the same way
-- the document's are, so a reader is never shown a line that disagrees with the
-- document it is part of.

-- SumWorkOrderSubLineCosts is the parts/labor split of one work-order line.
-- work_order_line_item.line_item_type is SERVICE_TASK/FREE_TEXT/ISSUE — what the
-- line is about — so the money split comes from the sub-line item_type instead.
-- name: SumWorkOrderSubLineCosts :one
SELECT
    COALESCE(SUM(quantity * unit_cost) FILTER (WHERE item_type = 'PART'), 0)::numeric(14,2)  AS parts_cost,
    COALESCE(SUM(quantity * unit_cost) FILTER (WHERE item_type = 'LABOR'), 0)::numeric(14,2) AS labor_cost,
    COALESCE(SUM(quantity * unit_cost), 0)::numeric(14,2)                                    AS subtotal
FROM work_order_sub_line_item
WHERE line_item_id = sqlc.arg(line_item_id);

-- name: StoreWorkOrderLineItemTotals :exec
UPDATE work_order_line_item SET
    parts_cost = sqlc.arg(parts_cost), labor_cost = sqlc.arg(labor_cost),
    subtotal = sqlc.arg(subtotal), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- A purchase-order line and a service-entry line each price themselves:
-- quantity x unit_cost, both columns on the row.
-- name: StorePurchaseOrderLineItemSubtotal :exec
UPDATE purchase_order_line_item SET
    subtotal = ROUND(quantity * unit_cost, 2), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: StoreServiceEntryLineItemSubtotal :exec
UPDATE service_entry_line_item SET
    subtotal = ROUND(quantity * unit_cost, 2), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: GetPurchaseOrderIDForLineItem :one
SELECT purchase_order_id FROM purchase_order_line_item WHERE id = sqlc.arg(id);

-- name: GetServiceEntryIDForLineItem :one
SELECT service_entry_id FROM service_entry_line_item WHERE id = sqlc.arg(id);
