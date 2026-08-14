-- name: GetServiceTask :one
SELECT * FROM service_task WHERE id = $1 AND company_id = $2;

-- name: ListServiceTasks :many
SELECT * FROM service_task WHERE company_id = $1 ORDER BY name, id LIMIT $2 OFFSET $3;

-- name: CountServiceTasks :one
SELECT count(*) FROM service_task WHERE company_id = $1;

-- name: CreateServiceTask :one
INSERT INTO service_task (
    company_id, name, description, expected_duration_seconds, parent_task_id, archived_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateServiceTask :one
UPDATE service_task SET name = $3, description = $4, expected_duration_seconds = $5, parent_task_id = $6, archived_at = $7, updated_at = $8
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteServiceTask :exec
DELETE FROM service_task WHERE id = $1 AND company_id = $2;
