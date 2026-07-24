-- name: ListPartInventories :many
SELECT c.* FROM part_inventory c JOIN part p ON p.id = c.part_id
WHERE c.part_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountPartInventories :one
SELECT count(*) FROM part_inventory c JOIN part p ON p.id = c.part_id WHERE c.part_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetPartInventory :one
SELECT c.* FROM part_inventory c JOIN part p ON p.id = c.part_id
WHERE c.id = sqlc.arg(id) AND c.part_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreatePartInventory :one
INSERT INTO part_inventory (
    part_id, location_id, available_quantity, expiry_date, aisle, "row", bin, reorder_point, reorder_point_enabled, reorder_quantity, reorder_point_lead_time_days, active, track_inventory, average_unit_cost, available_quantity_updated_at, created_at, updated_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(location_id), sqlc.arg(available_quantity), sqlc.arg(expiry_date), sqlc.arg(aisle), sqlc.arg(row), sqlc.arg(bin), sqlc.arg(reorder_point), sqlc.arg(reorder_point_enabled), sqlc.arg(reorder_quantity), sqlc.arg(reorder_point_lead_time_days), sqlc.arg(active), sqlc.arg(track_inventory), sqlc.arg(average_unit_cost), sqlc.arg(available_quantity_updated_at), sqlc.arg(created_at), sqlc.arg(updated_at)
WHERE EXISTS (SELECT 1 FROM part WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdatePartInventory :one
UPDATE part_inventory AS c SET location_id = sqlc.arg(location_id), available_quantity = sqlc.arg(available_quantity), expiry_date = sqlc.arg(expiry_date), aisle = sqlc.arg(aisle), "row" = sqlc.arg(row), bin = sqlc.arg(bin), reorder_point = sqlc.arg(reorder_point), reorder_point_enabled = sqlc.arg(reorder_point_enabled), reorder_quantity = sqlc.arg(reorder_quantity), reorder_point_lead_time_days = sqlc.arg(reorder_point_lead_time_days), active = sqlc.arg(active), track_inventory = sqlc.arg(track_inventory), average_unit_cost = sqlc.arg(average_unit_cost), available_quantity_updated_at = sqlc.arg(available_quantity_updated_at), updated_at = sqlc.arg(updated_at)
FROM part p
WHERE c.id = sqlc.arg(id) AND c.part_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.part_id
RETURNING c.*;

-- name: DeletePartInventory :exec
DELETE FROM part_inventory AS c USING part p
WHERE c.id = sqlc.arg(id) AND c.part_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.part_id;
