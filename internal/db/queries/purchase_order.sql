-- name: GetPurchaseOrder :one
SELECT * FROM purchase_order WHERE id = $1 AND company_id = $2;

-- name: ListPurchaseOrders :many
SELECT * FROM purchase_order WHERE company_id = $1 ORDER BY created_at DESC, id LIMIT $2 OFFSET $3;

-- name: CountPurchaseOrders :one
SELECT count(*) FROM purchase_order WHERE company_id = $1;

-- CreatePurchaseOrder opens an order in DRAFT. state and the workflow stamps are
-- absent on purpose: an order reaches any other state by being moved through
-- POST /purchase-orders/{id}/{action}, which checks the move is legal and
-- records who made it. Accepting them here would let a client create an order
-- that is already approved, by nobody.
-- name: CreatePurchaseOrder :one
INSERT INTO purchase_order (
    company_id, number, description, state, vendor_id, destination_id, discount_type, discount, discount_percentage, tax_1_type, tax_1, tax_1_percentage, tax_2_type, tax_2, tax_2_percentage, shipping, subtotal, total_amount, created_by_id, labels, custom_fields, created_at, updated_at
) VALUES (
    sqlc.arg(company_id), sqlc.arg(number), sqlc.arg(description), 'DRAFT', sqlc.arg(vendor_id), sqlc.arg(destination_id), sqlc.arg(discount_type), sqlc.arg(discount), sqlc.arg(discount_percentage), sqlc.arg(tax_1_type), sqlc.arg(tax_1), sqlc.arg(tax_1_percentage), sqlc.arg(tax_2_type), sqlc.arg(tax_2), sqlc.arg(tax_2_percentage), sqlc.arg(shipping), sqlc.arg(subtotal), sqlc.arg(total_amount), sqlc.narg(created_by_id), sqlc.arg(labels), sqlc.arg(custom_fields), sqlc.arg(created_at), sqlc.arg(updated_at)
)
RETURNING *;

-- UpdatePurchaseOrder edits the order's content. It does not touch state, the
-- workflow timestamps, the actor ids or rejection_reason: those belong to the
-- transition routes, which are the only thing that can decide whether a move is
-- legal and who made it. A whole-record replace that accepted them would let a
-- client set CLOSED without ever passing through approval, and name anybody as
-- the approver.
-- name: UpdatePurchaseOrder :one
UPDATE purchase_order SET
    number = sqlc.arg(number), description = sqlc.arg(description), vendor_id = sqlc.arg(vendor_id), destination_id = sqlc.arg(destination_id),
    discount_type = sqlc.arg(discount_type), discount = sqlc.arg(discount), discount_percentage = sqlc.arg(discount_percentage),
    tax_1_type = sqlc.arg(tax_1_type), tax_1 = sqlc.arg(tax_1), tax_1_percentage = sqlc.arg(tax_1_percentage),
    tax_2_type = sqlc.arg(tax_2_type), tax_2 = sqlc.arg(tax_2), tax_2_percentage = sqlc.arg(tax_2_percentage),
    shipping = sqlc.arg(shipping), subtotal = sqlc.arg(subtotal), total_amount = sqlc.arg(total_amount),
    labels = sqlc.arg(labels), custom_fields = sqlc.arg(custom_fields), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id)
RETURNING *;

-- name: DeletePurchaseOrder :exec
DELETE FROM purchase_order WHERE id = $1 AND company_id = $2;
