-- name: GetInventoryAdjustmentReason :one
SELECT * FROM inventory_adjustment_reason WHERE id = $1 AND company_id = $2;

-- name: ListInventoryAdjustmentReasons :many
SELECT * FROM inventory_adjustment_reason WHERE company_id = $1 ORDER BY name, id LIMIT $2 OFFSET $3;

-- name: CountInventoryAdjustmentReasons :one
SELECT count(*) FROM inventory_adjustment_reason WHERE company_id = $1;

-- name: CreateInventoryAdjustmentReason :one
INSERT INTO inventory_adjustment_reason (
    company_id, name, created_at
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdateInventoryAdjustmentReason :one
UPDATE inventory_adjustment_reason SET name = $3
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteInventoryAdjustmentReason :exec
DELETE FROM inventory_adjustment_reason WHERE id = $1 AND company_id = $2;
