-- name: GetWorkOrderStatus :one
SELECT * FROM work_order_status WHERE id = $1 AND company_id = $2;

-- name: ListWorkOrderStatuses :many
SELECT * FROM work_order_status WHERE company_id = $1 ORDER BY position, id LIMIT $2 OFFSET $3;

-- name: CountWorkOrderStatuses :one
SELECT count(*) FROM work_order_status WHERE company_id = $1;

-- name: CreateWorkOrderStatus :one
INSERT INTO work_order_status (
    company_id, name, description, color, is_default, marks_as_completed, position
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateWorkOrderStatus :one
UPDATE work_order_status SET name = $3, description = $4, color = $5, is_default = $6, marks_as_completed = $7, position = $8
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteWorkOrderStatus :exec
DELETE FROM work_order_status WHERE id = $1 AND company_id = $2;
