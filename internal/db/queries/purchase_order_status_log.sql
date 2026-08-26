-- Append-only, written by the transition that caused it, inside that
-- transaction. There is deliberately no update or delete.

-- name: CreatePurchaseOrderStatusLog :one
INSERT INTO purchase_order_status_log (
    purchase_order_id, from_state, to_state, actor_employee_id, actor_type, reason, changed_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(from_state), sqlc.arg(to_state), sqlc.narg(actor_employee_id), sqlc.arg(actor_type), sqlc.arg(reason), sqlc.arg(changed_at)
WHERE EXISTS (SELECT 1 FROM purchase_order WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: ListPurchaseOrderStatusLogs :many
SELECT c.* FROM purchase_order_status_log c JOIN purchase_order p ON p.id = c.purchase_order_id
WHERE c.purchase_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.changed_at DESC, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountPurchaseOrderStatusLogs :one
SELECT count(*) FROM purchase_order_status_log c JOIN purchase_order p ON p.id = c.purchase_order_id
WHERE c.purchase_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetPurchaseOrderStatusLog :one
SELECT c.* FROM purchase_order_status_log c JOIN purchase_order p ON p.id = c.purchase_order_id
WHERE c.id = sqlc.arg(id) AND c.purchase_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- LockPurchaseOrder reads the order and holds it for the transaction, so the
-- state a transition validates against is the state it then replaces. Without
-- it two concurrent approvals both see PENDING_APPROVAL and both stamp an
-- approval.
-- name: LockPurchaseOrder :one
SELECT * FROM purchase_order WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id) FOR UPDATE;

-- ApplyPurchaseOrderTransition moves the state and stamps the one timestamp the
-- transition owns. Each timestamp column is set by exactly one action, so the
-- caller passes only the pair it is responsible for and every other column is
-- left as it was.
-- name: ApplyPurchaseOrderTransition :one
UPDATE purchase_order SET
    state               = sqlc.arg(state),
    rejection_reason    = sqlc.arg(rejection_reason),
    submitted_at        = COALESCE(sqlc.narg(submitted_at), submitted_at),
    submitted_by_id     = COALESCE(sqlc.narg(submitted_by_id), submitted_by_id),
    approved_at         = COALESCE(sqlc.narg(approved_at), approved_at),
    approved_by_id      = COALESCE(sqlc.narg(approved_by_id), approved_by_id),
    rejected_at         = COALESCE(sqlc.narg(rejected_at), rejected_at),
    rejected_by_id      = COALESCE(sqlc.narg(rejected_by_id), rejected_by_id),
    purchased_at        = COALESCE(sqlc.narg(purchased_at), purchased_at),
    received_partial_at = COALESCE(sqlc.narg(received_partial_at), received_partial_at),
    received_full_at    = COALESCE(sqlc.narg(received_full_at), received_full_at),
    closed_at           = COALESCE(sqlc.narg(closed_at), closed_at),
    updated_at          = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id)
RETURNING *;
