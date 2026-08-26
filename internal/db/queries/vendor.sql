-- name: GetVendor :one
SELECT * FROM vendor WHERE id = $1 AND company_id = $2;

-- Archived rows are excluded unless include_archived is true. An archived
-- record is one somebody retired; showing it in the default list, and in the
-- pickers built from that list, is how it gets referenced again.
-- name: ListVendors :many
SELECT * FROM vendor
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.arg(include_archived)::boolean OR archived_at IS NULL)
ORDER BY name, id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountVendors :one
SELECT count(*) FROM vendor
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.arg(include_archived)::boolean OR archived_at IS NULL);

-- name: CreateVendor :one
INSERT INTO vendor (
    company_id, name, is_mobile_service, street_address, street_address_line_2, city, region, postal_code, country, phone, website, contact_name, contact_phone, contact_email, external_id, latitude, longitude, is_fuel_vendor, is_service_vendor, is_parts_vendor, labels, custom_fields, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
)
RETURNING *;

-- name: UpdateVendor :one
UPDATE vendor SET name = $3, is_mobile_service = $4, street_address = $5, street_address_line_2 = $6, city = $7, region = $8, postal_code = $9, country = $10, phone = $11, website = $12, contact_name = $13, contact_phone = $14, contact_email = $15, external_id = $16, latitude = $17, longitude = $18, is_fuel_vendor = $19, is_service_vendor = $20, is_parts_vendor = $21, labels = $22, custom_fields = $23, updated_at = $24
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteVendor :exec
DELETE FROM vendor WHERE id = $1 AND company_id = $2;
