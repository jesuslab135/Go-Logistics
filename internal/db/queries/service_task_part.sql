-- name: ListServiceTaskParts :many
SELECT c.* FROM service_task_part c JOIN service_task p ON p.id = c.service_task_id
WHERE c.service_task_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.position, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountServiceTaskParts :one
SELECT count(*) FROM service_task_part c JOIN service_task p ON p.id = c.service_task_id WHERE c.service_task_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetServiceTaskPart :one
SELECT c.* FROM service_task_part c JOIN service_task p ON p.id = c.service_task_id
WHERE c.id = sqlc.arg(id) AND c.service_task_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateServiceTaskPart :one
INSERT INTO service_task_part (
    service_task_id, part_id, quantity, position
)
SELECT sqlc.arg(parent_id), sqlc.arg(part_id), sqlc.arg(quantity), sqlc.arg(position)
WHERE EXISTS (SELECT 1 FROM service_task WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateServiceTaskPart :one
UPDATE service_task_part AS c SET part_id = sqlc.arg(part_id), quantity = sqlc.arg(quantity), position = sqlc.arg(position)
FROM service_task p
WHERE c.id = sqlc.arg(id) AND c.service_task_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.service_task_id
RETURNING c.*;

-- name: DeleteServiceTaskPart :exec
DELETE FROM service_task_part AS c USING service_task p
WHERE c.id = sqlc.arg(id) AND c.service_task_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.service_task_id;
