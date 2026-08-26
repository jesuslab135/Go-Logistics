-- The status log is append-only and is written by the operation that changes the
-- status, inside that transaction — never through a route of its own. There is
-- deliberately no update or delete here: a history a client can rewrite records
-- nothing. See internal/http/handler/work_order.go.

-- name: ListWorkOrderStatusLogs :many
SELECT c.* FROM work_order_status_log c JOIN work_order p ON p.id = c.work_order_id
WHERE c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.changed_at DESC, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountWorkOrderStatusLogs :one
SELECT count(*) FROM work_order_status_log c JOIN work_order p ON p.id = c.work_order_id WHERE c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetWorkOrderStatusLog :one
SELECT c.* FROM work_order_status_log c JOIN work_order p ON p.id = c.work_order_id
WHERE c.id = sqlc.arg(id) AND c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateWorkOrderStatusLog :one
INSERT INTO work_order_status_log (
    work_order_id, status_id, changed_at, actor_employee_id, actor_type
)
SELECT sqlc.arg(parent_id), sqlc.arg(status_id), sqlc.arg(changed_at), sqlc.narg(actor_employee_id), sqlc.arg(actor_type)
WHERE EXISTS (SELECT 1 FROM work_order WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;
