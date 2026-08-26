-- The journal is append-only: an entry is a stock movement that happened, and
-- history is not editable. There is deliberately no update or delete here.
-- Correcting an entry means reversing it, which is itself an entry.

-- name: GetInventoryJournalEntry :one
SELECT * FROM inventory_journal_entry WHERE id = $1 AND company_id = $2;

-- name: ListInventoryJournalEntries :many
SELECT * FROM inventory_journal_entry WHERE company_id = $1 ORDER BY created_at DESC, id LIMIT $2 OFFSET $3;

-- name: CountInventoryJournalEntries :one
SELECT count(*) FROM inventory_journal_entry WHERE company_id = $1;

-- name: CreateInventoryJournalEntry :one
INSERT INTO inventory_journal_entry (
    company_id, part_id, part_location_detail_id, user_id, previous_quantity, adjustment_quantity, current_quantity, unit_cost, reason_id, work_order_id, purchase_order_line_id, vendor_id, adjustment_type, transfer_part_location_id, notes, reversal_of_id, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING *;

-- LockPartInventory reads the stock row the entry will move, and holds it for
-- the rest of the transaction. Without the lock two concurrent adjustments read
-- the same previous_quantity and the second overwrites the first, losing a
-- movement the ledger still claims happened.
--
-- It is scoped through part so a caller cannot move another tenant's stock by
-- naming its part_inventory id.
-- name: LockPartInventory :one
SELECT pi.* FROM part_inventory pi
JOIN part p ON p.id = pi.part_id
WHERE pi.id = sqlc.arg(id) AND p.company_id = sqlc.arg(company_id)
FOR UPDATE OF pi;

-- name: ApplyPartInventoryAdjustment :one
UPDATE part_inventory
SET available_quantity = available_quantity + sqlc.arg(adjustment_quantity),
    available_quantity_updated_at = sqlc.arg(updated_at),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: GetInventoryJournalReversal :one
-- The reversal of an entry, if one exists. Reversing twice is refused, so this
-- is what a second attempt is checked against.
SELECT * FROM inventory_journal_entry WHERE reversal_of_id = sqlc.arg(reversal_of_id);

-- InventoryDrift reports where the ledger and the stock row disagree. Entries
-- written before the journal moved stock recorded movements the inventory never
-- made, so the two diverge by an unknown amount per part; this is what
-- `fleet-cli inventory-drift` prints rather than silently reconciling.
-- name: InventoryDrift :many
SELECT
    pi.id                                  AS part_inventory_id,
    pi.part_id,
    p.company_id,
    pi.available_quantity,
    COALESCE(SUM(ije.adjustment_quantity), 0)::numeric(14,2) AS ledger_sum
FROM part_inventory pi
JOIN part p ON p.id = pi.part_id
LEFT JOIN inventory_journal_entry ije ON ije.part_location_detail_id = pi.id
GROUP BY pi.id, pi.part_id, p.company_id, pi.available_quantity
HAVING pi.available_quantity <> COALESCE(SUM(ije.adjustment_quantity), 0)
ORDER BY p.company_id, pi.part_id, pi.id;
