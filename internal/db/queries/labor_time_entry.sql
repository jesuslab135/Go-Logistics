-- name: ListLaborTimeEntries :many
SELECT c.* FROM labor_time_entry c JOIN work_order_sub_line_item p0 ON p0.id = c.sub_line_item_id JOIN work_order_line_item p1 ON p1.id = p0.line_item_id JOIN work_order p2 ON p2.id = p1.work_order_id
WHERE c.sub_line_item_id = sqlc.arg(parent_id) AND p2.company_id = sqlc.arg(company_id)
ORDER BY c.started_at DESC LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountLaborTimeEntries :one
SELECT count(*) FROM labor_time_entry c JOIN work_order_sub_line_item p0 ON p0.id = c.sub_line_item_id JOIN work_order_line_item p1 ON p1.id = p0.line_item_id JOIN work_order p2 ON p2.id = p1.work_order_id WHERE c.sub_line_item_id = sqlc.arg(parent_id) AND p2.company_id = sqlc.arg(company_id);

-- name: GetLaborTimeEntry :one
SELECT c.* FROM labor_time_entry c JOIN work_order_sub_line_item p0 ON p0.id = c.sub_line_item_id JOIN work_order_line_item p1 ON p1.id = p0.line_item_id JOIN work_order p2 ON p2.id = p1.work_order_id
WHERE c.id = sqlc.arg(id) AND c.sub_line_item_id = sqlc.arg(parent_id) AND p2.company_id = sqlc.arg(company_id);

-- name: CreateLaborTimeEntry :one
INSERT INTO labor_time_entry (
    sub_line_item_id, technician_id, started_at, ended_at, duration_seconds, is_active, clock_in_latitude, clock_in_longitude, clock_out_latitude, clock_out_longitude, created_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(technician_id), sqlc.arg(started_at), sqlc.arg(ended_at), sqlc.arg(duration_seconds), sqlc.arg(is_active), sqlc.arg(clock_in_latitude), sqlc.arg(clock_in_longitude), sqlc.arg(clock_out_latitude), sqlc.arg(clock_out_longitude), sqlc.arg(created_at)
WHERE EXISTS (SELECT 1 FROM work_order_sub_line_item p0 JOIN work_order_line_item p1 ON p1.id = p0.line_item_id JOIN work_order p2 ON p2.id = p1.work_order_id WHERE p0.id = sqlc.arg(parent_id) AND p2.company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateLaborTimeEntry :one
UPDATE labor_time_entry AS c SET technician_id = sqlc.arg(technician_id), started_at = sqlc.arg(started_at), ended_at = sqlc.arg(ended_at), duration_seconds = sqlc.arg(duration_seconds), is_active = sqlc.arg(is_active), clock_in_latitude = sqlc.arg(clock_in_latitude), clock_in_longitude = sqlc.arg(clock_in_longitude), clock_out_latitude = sqlc.arg(clock_out_latitude), clock_out_longitude = sqlc.arg(clock_out_longitude)
FROM work_order_sub_line_item p0, work_order_line_item p1, work_order p2
WHERE c.id = sqlc.arg(id) AND c.sub_line_item_id = sqlc.arg(parent_id) AND p2.company_id = sqlc.arg(company_id) AND p0.id = c.sub_line_item_id AND p1.id = p0.line_item_id AND p2.id = p1.work_order_id
RETURNING c.*;

-- name: DeleteLaborTimeEntry :exec
DELETE FROM labor_time_entry AS c USING work_order_sub_line_item p0, work_order_line_item p1, work_order p2
WHERE c.id = sqlc.arg(id) AND c.sub_line_item_id = sqlc.arg(parent_id) AND p2.company_id = sqlc.arg(company_id) AND p0.id = c.sub_line_item_id AND p1.id = p0.line_item_id AND p2.id = p1.work_order_id;
