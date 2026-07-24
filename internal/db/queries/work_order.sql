-- name: GetWorkOrder :one
SELECT * FROM work_order WHERE id = $1 AND company_id = $2;

-- name: ListWorkOrders :many
SELECT * FROM work_order WHERE company_id = $1 ORDER BY issued_at DESC LIMIT $2 OFFSET $3;

-- name: CountWorkOrders :one
SELECT count(*) FROM work_order WHERE company_id = $1;

-- name: CreateWorkOrder :one
INSERT INTO work_order (
    location_id, company_id, number, description, asset_id, status_id, vendor_id, assigned_to_id, issued_by_id, fault_id, issued_at, scheduled_at, started_at, expected_completed_at, completed_at, starting_meter, ending_meter, duration_seconds, labor_time_seconds, parts_markup_type, parts_markup, parts_markup_percentage, labor_markup_type, labor_markup, labor_markup_percentage, parts_subtotal, labor_subtotal, subtotal, discount, discount_type, tax_1, tax_1_type, tax_1_percentage, tax_2, tax_2_type, tax_2_percentage, total_amount, invoice_number, purchase_order_number, comments_count, images_count, documents_count, labels, custom_fields, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46
)
RETURNING *;

-- name: UpdateWorkOrder :one
UPDATE work_order SET location_id = $3, number = $4, description = $5, asset_id = $6, status_id = $7, vendor_id = $8, assigned_to_id = $9, issued_by_id = $10, fault_id = $11, issued_at = $12, scheduled_at = $13, started_at = $14, expected_completed_at = $15, completed_at = $16, starting_meter = $17, ending_meter = $18, duration_seconds = $19, labor_time_seconds = $20, parts_markup_type = $21, parts_markup = $22, parts_markup_percentage = $23, labor_markup_type = $24, labor_markup = $25, labor_markup_percentage = $26, parts_subtotal = $27, labor_subtotal = $28, subtotal = $29, discount = $30, discount_type = $31, tax_1 = $32, tax_1_type = $33, tax_1_percentage = $34, tax_2 = $35, tax_2_type = $36, tax_2_percentage = $37, total_amount = $38, invoice_number = $39, purchase_order_number = $40, comments_count = $41, images_count = $42, documents_count = $43, labels = $44, custom_fields = $45, updated_at = $46
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteWorkOrder :exec
DELETE FROM work_order WHERE id = $1 AND company_id = $2;
