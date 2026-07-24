-- name: GetFuelType :one
SELECT * FROM fuel_type WHERE id = $1 AND company_id = $2;

-- name: ListFuelTypes :many
SELECT * FROM fuel_type WHERE company_id = $1 ORDER BY name LIMIT $2 OFFSET $3;

-- name: CountFuelTypes :one
SELECT count(*) FROM fuel_type WHERE company_id = $1;

-- name: CreateFuelType :one
INSERT INTO fuel_type (
    company_id, name, created_at
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdateFuelType :one
UPDATE fuel_type SET name = $3
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteFuelType :exec
DELETE FROM fuel_type WHERE id = $1 AND company_id = $2;
