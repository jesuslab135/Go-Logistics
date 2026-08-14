-- name: ListWorkOrderLineItems :many
SELECT c.* FROM work_order_line_item c JOIN work_order p ON p.id = c.work_order_id
WHERE c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.position, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountWorkOrderLineItems :one
SELECT count(*) FROM work_order_line_item c JOIN work_order p ON p.id = c.work_order_id WHERE c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetWorkOrderLineItem :one
SELECT c.* FROM work_order_line_item c JOIN work_order p ON p.id = c.work_order_id
WHERE c.id = sqlc.arg(id) AND c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateWorkOrderLineItem :one
INSERT INTO work_order_line_item (
    work_order_id, line_item_type, title, description, position, service_task, parts_cost, labor_cost, subtotal, created_at, updated_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(line_item_type), sqlc.arg(title), sqlc.arg(description), sqlc.arg(position), sqlc.arg(service_task), sqlc.arg(parts_cost), sqlc.arg(labor_cost), sqlc.arg(subtotal), sqlc.arg(created_at), sqlc.arg(updated_at)
WHERE EXISTS (SELECT 1 FROM work_order WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateWorkOrderLineItem :one
UPDATE work_order_line_item AS c SET line_item_type = sqlc.arg(line_item_type), title = sqlc.arg(title), description = sqlc.arg(description), position = sqlc.arg(position), service_task = sqlc.arg(service_task), parts_cost = sqlc.arg(parts_cost), labor_cost = sqlc.arg(labor_cost), subtotal = sqlc.arg(subtotal), updated_at = sqlc.arg(updated_at)
FROM work_order p
WHERE c.id = sqlc.arg(id) AND c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.work_order_id
RETURNING c.*;

-- name: DeleteWorkOrderLineItem :exec
DELETE FROM work_order_line_item AS c USING work_order p
WHERE c.id = sqlc.arg(id) AND c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.work_order_id;
