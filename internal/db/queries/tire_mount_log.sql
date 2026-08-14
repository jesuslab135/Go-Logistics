-- name: ListTireMountLogs :many
SELECT c.* FROM tire_mount_log c JOIN tire p ON p.id = c.tire_id
WHERE c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.event_date DESC, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountTireMountLogs :one
SELECT count(*) FROM tire_mount_log c JOIN tire p ON p.id = c.tire_id WHERE c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetTireMountLog :one
SELECT c.* FROM tire_mount_log c JOIN tire p ON p.id = c.tire_id
WHERE c.id = sqlc.arg(id) AND c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateTireMountLog :one
INSERT INTO tire_mount_log (
    tire_id, vehicle_id, position_code, event_type, event_date, odometer, tread_depth_32nds, psi, performed_by_id, reason
)
SELECT sqlc.arg(parent_id), sqlc.arg(vehicle_id), sqlc.arg(position_code), sqlc.arg(event_type), sqlc.arg(event_date), sqlc.arg(odometer), sqlc.arg(tread_depth_32nds), sqlc.arg(psi), sqlc.arg(performed_by_id), sqlc.arg(reason)
WHERE EXISTS (SELECT 1 FROM tire WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateTireMountLog :one
UPDATE tire_mount_log AS c SET vehicle_id = sqlc.arg(vehicle_id), position_code = sqlc.arg(position_code), event_type = sqlc.arg(event_type), event_date = sqlc.arg(event_date), odometer = sqlc.arg(odometer), tread_depth_32nds = sqlc.arg(tread_depth_32nds), psi = sqlc.arg(psi), performed_by_id = sqlc.arg(performed_by_id), reason = sqlc.arg(reason)
FROM tire p
WHERE c.id = sqlc.arg(id) AND c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.tire_id
RETURNING c.*;

-- name: DeleteTireMountLog :exec
DELETE FROM tire_mount_log AS c USING tire p
WHERE c.id = sqlc.arg(id) AND c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.tire_id;
