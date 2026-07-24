-- name: GetAsset :one
SELECT * FROM asset WHERE id = $1;

-- name: ListAssetsByCompany :many
SELECT * FROM asset WHERE company_id = $1 ORDER BY name;

-- name: ArchiveAsset :exec
UPDATE asset SET archived_at = $2, updated_at = $3 WHERE id = $1;

-- name: SetAssetGroup :exec
UPDATE asset SET "group" = $2, updated_at = $3 WHERE id = $1;

-- name: DeleteAsset :exec
DELETE FROM asset WHERE id = $1;
