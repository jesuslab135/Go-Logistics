-- name: GetPartCategory :one
SELECT * FROM part_category WHERE id = $1 AND company_id = $2;

-- name: ListPartCategories :many
SELECT * FROM part_category WHERE company_id = $1 ORDER BY name, id LIMIT $2 OFFSET $3;

-- name: CountPartCategories :one
SELECT count(*) FROM part_category WHERE company_id = $1;

-- name: CreatePartCategory :one
INSERT INTO part_category (
    company_id, name, description, created_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdatePartCategory :one
UPDATE part_category SET name = $3, description = $4
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeletePartCategory :exec
DELETE FROM part_category WHERE id = $1 AND company_id = $2;
