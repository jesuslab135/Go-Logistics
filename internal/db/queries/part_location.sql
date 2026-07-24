-- name: GetPartLocation :one
SELECT * FROM part_location WHERE id = $1 AND company_id = $2;

-- name: ListPartLocations :many
SELECT * FROM part_location WHERE company_id = $1 ORDER BY name LIMIT $2 OFFSET $3;

-- name: CountPartLocations :one
SELECT count(*) FROM part_location WHERE company_id = $1;

-- name: CreatePartLocation :one
INSERT INTO part_location (
    company_id, name, address, city, region, location_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdatePartLocation :one
UPDATE part_location SET name = $3, address = $4, city = $5, region = $6, location_id = $7, updated_at = $8
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeletePartLocation :exec
DELETE FROM part_location WHERE id = $1 AND company_id = $2;
