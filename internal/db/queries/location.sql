-- name: GetLocation :one
SELECT * FROM location WHERE id = $1 AND company_id = $2;

-- name: ListLocations :many
SELECT * FROM location WHERE company_id = $1 ORDER BY name LIMIT $2 OFFSET $3;

-- name: CountLocations :one
SELECT count(*) FROM location WHERE company_id = $1;

-- name: CreateLocation :one
INSERT INTO location (
    company_id, name, is_active
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdateLocation :one
UPDATE location SET name = $3, is_active = $4
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteLocation :exec
DELETE FROM location WHERE id = $1 AND company_id = $2;
