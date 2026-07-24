-- name: GetFault :one
SELECT * FROM fault WHERE id = $1 AND company_id = $2;

-- name: ListFaults :many
SELECT * FROM fault WHERE company_id = $1 ORDER BY code LIMIT $2 OFFSET $3;

-- name: CountFaults :one
SELECT count(*) FROM fault WHERE company_id = $1;

-- name: CreateFault :one
INSERT INTO fault (
    company_id, family, code, name, description, applies_to_asset_types
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateFault :one
UPDATE fault SET family = $3, code = $4, name = $5, description = $6, applies_to_asset_types = $7
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteFault :exec
DELETE FROM fault WHERE id = $1 AND company_id = $2;
