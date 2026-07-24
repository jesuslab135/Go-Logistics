-- name: ListWorkOrderSubLineItems :many
SELECT c.* FROM work_order_sub_line_item c JOIN work_order_line_item p0 ON p0.id = c.line_item_id JOIN work_order p1 ON p1.id = p0.work_order_id
WHERE c.line_item_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id)
ORDER BY c.position LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountWorkOrderSubLineItems :one
SELECT count(*) FROM work_order_sub_line_item c JOIN work_order_line_item p0 ON p0.id = c.line_item_id JOIN work_order p1 ON p1.id = p0.work_order_id WHERE c.line_item_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id);

-- name: GetWorkOrderSubLineItem :one
SELECT c.* FROM work_order_sub_line_item c JOIN work_order_line_item p0 ON p0.id = c.line_item_id JOIN work_order p1 ON p1.id = p0.work_order_id
WHERE c.id = sqlc.arg(id) AND c.line_item_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id);

-- name: CreateWorkOrderSubLineItem :one
INSERT INTO work_order_sub_line_item (
    line_item_id, item_type, description, position, part_id, part_location_detail_id, technician_id, unit_cost, quantity, created_at, updated_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(item_type), sqlc.arg(description), sqlc.arg(position), sqlc.arg(part_id), sqlc.arg(part_location_detail_id), sqlc.arg(technician_id), sqlc.arg(unit_cost), sqlc.arg(quantity), sqlc.arg(created_at), sqlc.arg(updated_at)
WHERE EXISTS (SELECT 1 FROM work_order_line_item p0 JOIN work_order p1 ON p1.id = p0.work_order_id WHERE p0.id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateWorkOrderSubLineItem :one
UPDATE work_order_sub_line_item AS c SET item_type = sqlc.arg(item_type), description = sqlc.arg(description), position = sqlc.arg(position), part_id = sqlc.arg(part_id), part_location_detail_id = sqlc.arg(part_location_detail_id), technician_id = sqlc.arg(technician_id), unit_cost = sqlc.arg(unit_cost), quantity = sqlc.arg(quantity), updated_at = sqlc.arg(updated_at)
FROM work_order_line_item p0, work_order p1
WHERE c.id = sqlc.arg(id) AND c.line_item_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id) AND p0.id = c.line_item_id AND p1.id = p0.work_order_id
RETURNING c.*;

-- name: DeleteWorkOrderSubLineItem :exec
DELETE FROM work_order_sub_line_item AS c USING work_order_line_item p0, work_order p1
WHERE c.id = sqlc.arg(id) AND c.line_item_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id) AND p0.id = c.line_item_id AND p1.id = p0.work_order_id;
