-- The provider's name is joined in so a warranty row renders without a second
-- request per row just to resolve the vendor.

-- name: GetWarranty :one
SELECT w.*, v.name AS provider_name FROM warranty w
JOIN vendor v ON v.id = w.provider_id
WHERE w.id = sqlc.arg(id) AND w.company_id = sqlc.arg(company_id);

-- name: ListWarranties :many
SELECT w.*, v.name AS provider_name FROM warranty w
JOIN vendor v ON v.id = w.provider_id
WHERE w.company_id = sqlc.arg(company_id)
ORDER BY w.end_date DESC, w.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountWarranties :one
SELECT count(*) FROM warranty WHERE company_id = $1;

-- name: CreateWarranty :one
INSERT INTO warranty (
    company_id, provider_id, asset_id, part_id, start_date, end_date, terms, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateWarranty :one
UPDATE warranty SET provider_id = $3, asset_id = $4, part_id = $5, start_date = $6, end_date = $7, terms = $8, is_active = $9
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteWarranty :exec
DELETE FROM warranty WHERE id = $1 AND company_id = $2;
