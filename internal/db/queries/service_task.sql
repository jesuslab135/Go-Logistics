-- name: GetServiceTask :one
SELECT * FROM service_task WHERE id = $1 AND company_id = $2;

-- Archived rows are excluded unless include_archived is true. An archived
-- record is one somebody retired; showing it in the default list, and in the
-- pickers built from that list, is how it gets referenced again.
-- name: ListServiceTasks :many
SELECT * FROM service_task
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.arg(include_archived)::boolean OR archived_at IS NULL)
ORDER BY name, id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountServiceTasks :one
SELECT count(*) FROM service_task
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.arg(include_archived)::boolean OR archived_at IS NULL);

-- name: CreateServiceTask :one
INSERT INTO service_task (
    company_id, name, description, expected_duration_seconds, parent_task_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateServiceTask :one
UPDATE service_task SET name = $3, description = $4, expected_duration_seconds = $5, parent_task_id = $6, updated_at = $7
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteServiceTask :exec
DELETE FROM service_task WHERE id = $1 AND company_id = $2;
