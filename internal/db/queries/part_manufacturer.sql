-- name: GetPartManufacturer :one
SELECT * FROM part_manufacturer WHERE id = $1 AND company_id = $2;

-- name: ListPartManufacturers :many
SELECT * FROM part_manufacturer WHERE company_id = $1 ORDER BY name LIMIT $2 OFFSET $3;

-- name: CountPartManufacturers :one
SELECT count(*) FROM part_manufacturer WHERE company_id = $1;

-- name: CreatePartManufacturer :one
INSERT INTO part_manufacturer (
    company_id, name, website, created_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdatePartManufacturer :one
UPDATE part_manufacturer SET name = $3, website = $4
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeletePartManufacturer :exec
DELETE FROM part_manufacturer WHERE id = $1 AND company_id = $2;
