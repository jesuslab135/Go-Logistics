-- name: ListPurchaseOrderLineItems :many
SELECT c.* FROM purchase_order_line_item c JOIN purchase_order p ON p.id = c.purchase_order_id
WHERE c.purchase_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.position, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountPurchaseOrderLineItems :one
SELECT count(*) FROM purchase_order_line_item c JOIN purchase_order p ON p.id = c.purchase_order_id WHERE c.purchase_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetPurchaseOrderLineItem :one
SELECT c.* FROM purchase_order_line_item c JOIN purchase_order p ON p.id = c.purchase_order_id
WHERE c.id = sqlc.arg(id) AND c.purchase_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreatePurchaseOrderLineItem :one
INSERT INTO purchase_order_line_item (
    purchase_order_id, part_id, quantity, total_received, unit_cost, subtotal, position, created_at, updated_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(part_id), sqlc.arg(quantity), sqlc.arg(total_received), sqlc.arg(unit_cost), sqlc.arg(subtotal), sqlc.arg(position), sqlc.arg(created_at), sqlc.arg(updated_at)
WHERE EXISTS (SELECT 1 FROM purchase_order WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdatePurchaseOrderLineItem :one
UPDATE purchase_order_line_item AS c SET part_id = sqlc.arg(part_id), quantity = sqlc.arg(quantity), total_received = sqlc.arg(total_received), unit_cost = sqlc.arg(unit_cost), subtotal = sqlc.arg(subtotal), position = sqlc.arg(position), updated_at = sqlc.arg(updated_at)
FROM purchase_order p
WHERE c.id = sqlc.arg(id) AND c.purchase_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.purchase_order_id
RETURNING c.*;

-- name: DeletePurchaseOrderLineItem :exec
DELETE FROM purchase_order_line_item AS c USING purchase_order p
WHERE c.id = sqlc.arg(id) AND c.purchase_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.purchase_order_id;
