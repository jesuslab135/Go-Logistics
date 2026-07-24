-- name: GetCompany :one
SELECT * FROM company WHERE id = $1;

-- name: ListCompanies :many
SELECT * FROM company ORDER BY name;

-- name: CreateCompany :one
INSERT INTO company (
    name, tax_id, address, created_at, phone, email, website, logo,
    city, region, postal_code, country, timezone, currency, system_of_measurement
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: UpdateCompany :one
UPDATE company SET
    name = $2, tax_id = $3, address = $4, phone = $5, email = $6, website = $7,
    logo = $8, city = $9, region = $10, postal_code = $11, country = $12,
    timezone = $13, currency = $14, system_of_measurement = $15
WHERE id = $1
RETURNING *;

-- name: DeleteCompany :exec
DELETE FROM company WHERE id = $1;
