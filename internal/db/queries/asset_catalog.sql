-- name: GetAssetType :one
SELECT * FROM asset_type WHERE id = $1 AND company_id = $2;

-- name: ListAssetTypes :many
SELECT * FROM asset_type WHERE company_id = $1 ORDER BY name, id LIMIT $2 OFFSET $3;

-- name: CountAssetTypes :one
SELECT count(*) FROM asset_type WHERE company_id = $1;

-- name: CreateAssetType :one
INSERT INTO asset_type (company_id, name, category, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateAssetType :one
UPDATE asset_type SET name = $3, category = $4, description = $5
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteAssetType :exec
DELETE FROM asset_type WHERE id = $1 AND company_id = $2;

-- name: GetAssetStatus :one
SELECT * FROM asset_status WHERE id = $1 AND company_id = $2;

-- name: ListAssetStatuses :many
SELECT * FROM asset_status WHERE company_id = $1 ORDER BY name, id LIMIT $2 OFFSET $3;

-- name: CountAssetStatuses :one
SELECT count(*) FROM asset_status WHERE company_id = $1;

-- name: CreateAssetStatus :one
INSERT INTO asset_status (company_id, name, color_code)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateAssetStatus :one
UPDATE asset_status SET name = $3, color_code = $4
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteAssetStatus :exec
DELETE FROM asset_status WHERE id = $1 AND company_id = $2;

-- name: GetCatalogOption :one
SELECT * FROM catalog_option WHERE id = $1 AND company_id = $2;

-- name: ListCatalogOptions :many
SELECT * FROM catalog_option WHERE company_id = $1 ORDER BY category, value, id LIMIT $2 OFFSET $3;

-- name: CountCatalogOptions :one
SELECT count(*) FROM catalog_option WHERE company_id = $1;

-- name: CreateCatalogOption :one
INSERT INTO catalog_option (company_id, category, value)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateCatalogOption :one
UPDATE catalog_option SET category = $3, value = $4
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteCatalogOption :exec
DELETE FROM catalog_option WHERE id = $1 AND company_id = $2;
