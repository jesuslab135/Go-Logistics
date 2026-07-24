-- name: ListServiceEntryLineItems :many
SELECT c.* FROM service_entry_line_item c JOIN service_entry p ON p.id = c.service_entry_id
WHERE c.service_entry_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.position LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountServiceEntryLineItems :one
SELECT count(*) FROM service_entry_line_item c JOIN service_entry p ON p.id = c.service_entry_id WHERE c.service_entry_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetServiceEntryLineItem :one
SELECT c.* FROM service_entry_line_item c JOIN service_entry p ON p.id = c.service_entry_id
WHERE c.id = sqlc.arg(id) AND c.service_entry_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateServiceEntryLineItem :one
INSERT INTO service_entry_line_item (
    service_entry_id, line_item_type, description, service_task_id, part_id, technician_id, tire_id, service_reminder_id, unit_cost, quantity, parts_cost, labor_cost, subtotal, position, created_at, updated_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(line_item_type), sqlc.arg(description), sqlc.arg(service_task_id), sqlc.arg(part_id), sqlc.arg(technician_id), sqlc.arg(tire_id), sqlc.arg(service_reminder_id), sqlc.arg(unit_cost), sqlc.arg(quantity), sqlc.arg(parts_cost), sqlc.arg(labor_cost), sqlc.arg(subtotal), sqlc.arg(position), sqlc.arg(created_at), sqlc.arg(updated_at)
WHERE EXISTS (SELECT 1 FROM service_entry WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateServiceEntryLineItem :one
UPDATE service_entry_line_item AS c SET line_item_type = sqlc.arg(line_item_type), description = sqlc.arg(description), service_task_id = sqlc.arg(service_task_id), part_id = sqlc.arg(part_id), technician_id = sqlc.arg(technician_id), tire_id = sqlc.arg(tire_id), service_reminder_id = sqlc.arg(service_reminder_id), unit_cost = sqlc.arg(unit_cost), quantity = sqlc.arg(quantity), parts_cost = sqlc.arg(parts_cost), labor_cost = sqlc.arg(labor_cost), subtotal = sqlc.arg(subtotal), position = sqlc.arg(position), updated_at = sqlc.arg(updated_at)
FROM service_entry p
WHERE c.id = sqlc.arg(id) AND c.service_entry_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.service_entry_id
RETURNING c.*;

-- name: DeleteServiceEntryLineItem :exec
DELETE FROM service_entry_line_item AS c USING service_entry p
WHERE c.id = sqlc.arg(id) AND c.service_entry_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.service_entry_id;
