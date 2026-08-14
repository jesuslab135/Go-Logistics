-- name: GetInventoryJournalEntry :one
SELECT * FROM inventory_journal_entry WHERE id = $1 AND company_id = $2;

-- name: ListInventoryJournalEntries :many
SELECT * FROM inventory_journal_entry WHERE company_id = $1 ORDER BY created_at DESC, id LIMIT $2 OFFSET $3;

-- name: CountInventoryJournalEntries :one
SELECT count(*) FROM inventory_journal_entry WHERE company_id = $1;

-- name: CreateInventoryJournalEntry :one
INSERT INTO inventory_journal_entry (
    company_id, part_id, part_location_detail_id, user_id, previous_quantity, adjustment_quantity, current_quantity, unit_cost, reason_id, work_order_id, purchase_order_line_id, vendor_id, adjustment_type, transfer_part_location_id, notes, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING *;

-- name: UpdateInventoryJournalEntry :one
UPDATE inventory_journal_entry SET part_id = $3, part_location_detail_id = $4, user_id = $5, previous_quantity = $6, adjustment_quantity = $7, current_quantity = $8, unit_cost = $9, reason_id = $10, work_order_id = $11, purchase_order_line_id = $12, vendor_id = $13, adjustment_type = $14, transfer_part_location_id = $15, notes = $16
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteInventoryJournalEntry :exec
DELETE FROM inventory_journal_entry WHERE id = $1 AND company_id = $2;
