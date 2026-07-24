-- name: GetMedium :one
SELECT * FROM media WHERE id = $1 AND company_id = $2;

-- name: ListMediaItems :many
SELECT * FROM media WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountMediaItems :one
SELECT count(*) FROM media WHERE company_id = $1;

-- name: CreateMedium :one
INSERT INTO media (
    company_id, asset_id, file, title, description, file_type, file_size, uploaded_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateMedium :one
UPDATE media SET asset_id = $3, file = $4, title = $5, description = $6, file_type = $7, file_size = $8, uploaded_by_id = $9, updated_at = $10
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteMedium :exec
DELETE FROM media WHERE id = $1 AND company_id = $2;
