-- name: GetVehicleMake :one
SELECT * FROM vehicle_make WHERE id = $1 AND company_id = $2;

-- name: ListVehicleMakes :many
SELECT * FROM vehicle_make WHERE company_id = $1 ORDER BY name LIMIT $2 OFFSET $3;

-- name: CountVehicleMakes :one
SELECT count(*) FROM vehicle_make WHERE company_id = $1;

-- name: CreateVehicleMake :one
INSERT INTO vehicle_make (company_id, name, created_at, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateVehicleMake :one
UPDATE vehicle_make SET name = $3, updated_at = $4
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteVehicleMake :exec
DELETE FROM vehicle_make WHERE id = $1 AND company_id = $2;

-- name: GetVehicleModel :one
SELECT * FROM vehicle_model WHERE id = $1 AND company_id = $2;

-- name: ListVehicleModels :many
SELECT * FROM vehicle_model WHERE company_id = $1 ORDER BY name LIMIT $2 OFFSET $3;

-- name: CountVehicleModels :one
SELECT count(*) FROM vehicle_model WHERE company_id = $1;

-- name: CreateVehicleModel :one
INSERT INTO vehicle_model (company_id, name, make_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateVehicleModel :one
UPDATE vehicle_model SET name = $3, make_id = $4, updated_at = $5
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteVehicleModel :exec
DELETE FROM vehicle_model WHERE id = $1 AND company_id = $2;
