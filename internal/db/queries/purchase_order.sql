-- name: GetPurchaseOrder :one
SELECT * FROM purchase_order WHERE id = $1 AND company_id = $2;

-- name: ListPurchaseOrders :many
SELECT * FROM purchase_order WHERE company_id = $1 ORDER BY created_at DESC, id LIMIT $2 OFFSET $3;

-- name: CountPurchaseOrders :one
SELECT count(*) FROM purchase_order WHERE company_id = $1;

-- name: CreatePurchaseOrder :one
INSERT INTO purchase_order (
    company_id, number, description, state, vendor_id, destination_id, discount_type, discount, discount_percentage, tax_1_type, tax_1, tax_1_percentage, tax_2_type, tax_2, tax_2_percentage, shipping, subtotal, total_amount, created_by_id, submitted_at, submitted_by_id, rejected_at, rejected_by_id, approved_at, approved_by_id, purchased_at, received_partial_at, received_full_at, closed_at, labels, custom_fields, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33
)
RETURNING *;

-- name: UpdatePurchaseOrder :one
UPDATE purchase_order SET number = $3, description = $4, state = $5, vendor_id = $6, destination_id = $7, discount_type = $8, discount = $9, discount_percentage = $10, tax_1_type = $11, tax_1 = $12, tax_1_percentage = $13, tax_2_type = $14, tax_2 = $15, tax_2_percentage = $16, shipping = $17, subtotal = $18, total_amount = $19, created_by_id = $20, submitted_at = $21, submitted_by_id = $22, rejected_at = $23, rejected_by_id = $24, approved_at = $25, approved_by_id = $26, purchased_at = $27, received_partial_at = $28, received_full_at = $29, closed_at = $30, labels = $31, custom_fields = $32, updated_at = $33
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeletePurchaseOrder :exec
DELETE FROM purchase_order WHERE id = $1 AND company_id = $2;
