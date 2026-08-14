-- name: ListTireInstallations :many
SELECT c.* FROM tire_installation c JOIN tire p ON p.id = c.tire_id
WHERE c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.install_date DESC, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountTireInstallations :one
SELECT count(*) FROM tire_installation c JOIN tire p ON p.id = c.tire_id WHERE c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetTireInstallation :one
SELECT c.* FROM tire_installation c JOIN tire p ON p.id = c.tire_id
WHERE c.id = sqlc.arg(id) AND c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateTireInstallation :one
INSERT INTO tire_installation (
    tire_id, vehicle_id, position_code, install_date, odometer_at_install, tread_depth_at_install_32nds, psi_at_install, installed_by_id, notes
)
SELECT sqlc.arg(parent_id), sqlc.arg(vehicle_id), sqlc.arg(position_code), sqlc.arg(install_date), sqlc.arg(odometer_at_install), sqlc.arg(tread_depth_at_install_32nds), sqlc.arg(psi_at_install), sqlc.arg(installed_by_id), sqlc.arg(notes)
WHERE EXISTS (SELECT 1 FROM tire WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateTireInstallation :one
UPDATE tire_installation AS c SET vehicle_id = sqlc.arg(vehicle_id), position_code = sqlc.arg(position_code), install_date = sqlc.arg(install_date), odometer_at_install = sqlc.arg(odometer_at_install), tread_depth_at_install_32nds = sqlc.arg(tread_depth_at_install_32nds), psi_at_install = sqlc.arg(psi_at_install), installed_by_id = sqlc.arg(installed_by_id), notes = sqlc.arg(notes)
FROM tire p
WHERE c.id = sqlc.arg(id) AND c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.tire_id
RETURNING c.*;

-- name: DeleteTireInstallation :exec
DELETE FROM tire_installation AS c USING tire p
WHERE c.id = sqlc.arg(id) AND c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.tire_id;

-- name: TirePositionOccupied :one
SELECT EXISTS(SELECT 1 FROM tire_installation WHERE vehicle_id = $1 AND position_code = $2)::boolean AS occupied;

-- name: TireHasInstallation :one
SELECT EXISTS(SELECT 1 FROM tire_installation WHERE tire_id = $1)::boolean AS installed;
