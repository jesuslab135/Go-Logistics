-- name: ListWorkOrderStatusLogs :many
SELECT c.* FROM work_order_status_log c JOIN work_order p ON p.id = c.work_order_id
WHERE c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.changed_at DESC LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountWorkOrderStatusLogs :one
SELECT count(*) FROM work_order_status_log c JOIN work_order p ON p.id = c.work_order_id WHERE c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetWorkOrderStatusLog :one
SELECT c.* FROM work_order_status_log c JOIN work_order p ON p.id = c.work_order_id
WHERE c.id = sqlc.arg(id) AND c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateWorkOrderStatusLog :one
INSERT INTO work_order_status_log (
    work_order_id, status_id, changed_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(status_id), sqlc.arg(changed_at)
WHERE EXISTS (SELECT 1 FROM work_order WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateWorkOrderStatusLog :one
UPDATE work_order_status_log AS c SET status_id = sqlc.arg(status_id), changed_at = sqlc.arg(changed_at)
FROM work_order p
WHERE c.id = sqlc.arg(id) AND c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.work_order_id
RETURNING c.*;

-- name: DeleteWorkOrderStatusLog :exec
DELETE FROM work_order_status_log AS c USING work_order p
WHERE c.id = sqlc.arg(id) AND c.work_order_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.work_order_id;
