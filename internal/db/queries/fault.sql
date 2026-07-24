-- name: GetFault :one
SELECT * FROM fault WHERE id = $1;

-- name: ListFaultsByCompany :many
SELECT * FROM fault WHERE company_id = $1 ORDER BY code;

-- name: CreateFault :one
INSERT INTO fault (
    company_id, family, code, name, description, applies_to_asset_types
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;
