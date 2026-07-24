-- name: GetServiceEntry :one
SELECT * FROM service_entry WHERE id = $1 AND company_id = $2;

-- name: ListServiceEntries :many
SELECT * FROM service_entry WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountServiceEntries :one
SELECT count(*) FROM service_entry WHERE company_id = $1;

-- name: CreateServiceEntry :one
INSERT INTO service_entry (
    company_id, reference, status, asset_id, vendor_id, work_order_id, started_at, completed_at, meter_value, parts_subtotal, labor_subtotal, subtotal, discount, discount_type, tax_1, tax_1_type, tax_1_percentage, tax_2, tax_2_type, tax_2_percentage, total_amount, general_notes, is_roadside_assistance, labor_time_seconds, labels, custom_fields, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
)
RETURNING *;

-- name: UpdateServiceEntry :one
UPDATE service_entry SET reference = $3, status = $4, asset_id = $5, vendor_id = $6, work_order_id = $7, started_at = $8, completed_at = $9, meter_value = $10, parts_subtotal = $11, labor_subtotal = $12, subtotal = $13, discount = $14, discount_type = $15, tax_1 = $16, tax_1_type = $17, tax_1_percentage = $18, tax_2 = $19, tax_2_type = $20, tax_2_percentage = $21, total_amount = $22, general_notes = $23, is_roadside_assistance = $24, labor_time_seconds = $25, labels = $26, custom_fields = $27, updated_at = $28
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteServiceEntry :exec
DELETE FROM service_entry WHERE id = $1 AND company_id = $2;
