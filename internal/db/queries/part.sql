-- name: GetPart :one
SELECT * FROM part WHERE id = $1 AND company_id = $2;

-- Archived rows are excluded unless include_archived is true. An archived
-- record is one somebody retired; showing it in the default list, and in the
-- pickers built from that list, is how it gets referenced again.
-- name: ListParts :many
SELECT * FROM part
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.arg(include_archived)::boolean OR archived_at IS NULL)
ORDER BY part_number, id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountParts :one
SELECT count(*) FROM part
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.arg(include_archived)::boolean OR archived_at IS NULL);

-- name: CreatePart :one
INSERT INTO part (
    company_id, part_number, description, useful_life_months, useful_life_distance, part_category_id, part_manufacturer_id, measurement_unit_id, manufacturer_part_number, supplier_part_number, upc, unit_cost, inventory_item, custom_fields, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING *;

-- name: UpdatePart :one
UPDATE part SET part_number = $3, description = $4, useful_life_months = $5, useful_life_distance = $6, part_category_id = $7, part_manufacturer_id = $8, measurement_unit_id = $9, manufacturer_part_number = $10, supplier_part_number = $11, upc = $12, unit_cost = $13, inventory_item = $14, custom_fields = $15, updated_at = $16
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeletePart :exec
DELETE FROM part WHERE id = $1 AND company_id = $2;
