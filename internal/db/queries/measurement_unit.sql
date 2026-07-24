-- name: GetMeasurementUnit :one
SELECT * FROM measurement_unit WHERE id = $1 AND company_id = $2;

-- name: ListMeasurementUnits :many
SELECT * FROM measurement_unit WHERE company_id = $1 ORDER BY name LIMIT $2 OFFSET $3;

-- name: CountMeasurementUnits :one
SELECT count(*) FROM measurement_unit WHERE company_id = $1;

-- name: CreateMeasurementUnit :one
INSERT INTO measurement_unit (
    company_id, name, abbreviation, created_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateMeasurementUnit :one
UPDATE measurement_unit SET name = $3, abbreviation = $4
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteMeasurementUnit :exec
DELETE FROM measurement_unit WHERE id = $1 AND company_id = $2;
