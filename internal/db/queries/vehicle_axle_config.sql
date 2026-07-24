-- name: GetVehicleAxleConfig :one
SELECT c.* FROM vehicle_axle_config c JOIN asset a ON a.id = c.vehicle_id
WHERE c.vehicle_id = sqlc.arg(parent_id) AND a.company_id = sqlc.arg(company_id);

-- name: UpsertVehicleAxleConfig :one
INSERT INTO vehicle_axle_config (
    vehicle_id, template_id, display_name
)
SELECT sqlc.arg(parent_id), sqlc.arg(template_id), sqlc.arg(display_name)
WHERE EXISTS (SELECT 1 FROM asset WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
ON CONFLICT (vehicle_id) DO UPDATE SET template_id = EXCLUDED.template_id, display_name = EXCLUDED.display_name
RETURNING *;

-- name: DeleteVehicleAxleConfig :exec
DELETE FROM vehicle_axle_config c USING asset a
WHERE c.vehicle_id = sqlc.arg(parent_id) AND a.id = c.vehicle_id AND a.company_id = sqlc.arg(company_id);
