-- name: GetTireModel :one
SELECT * FROM tire_model WHERE id = $1 AND company_id = $2;

-- name: ListTireModels :many
SELECT * FROM tire_model WHERE company_id = $1 ORDER BY brand, id LIMIT $2 OFFSET $3;

-- name: CountTireModels :one
SELECT count(*) FROM tire_model WHERE company_id = $1;

-- name: CreateTireModel :one
INSERT INTO tire_model (
    company_id, brand, model_name, size, factory_tread_depth_32nds, minimum_tread_depth_32nds, life_expectancy_miles, recommended_psi
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateTireModel :one
UPDATE tire_model SET brand = $3, model_name = $4, size = $5, factory_tread_depth_32nds = $6, minimum_tread_depth_32nds = $7, life_expectancy_miles = $8, recommended_psi = $9
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteTireModel :exec
DELETE FROM tire_model WHERE id = $1 AND company_id = $2;
