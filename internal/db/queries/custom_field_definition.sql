-- What a company has declared may go in a resource's custom_fields document.

-- name: GetCustomFieldDefinition :one
SELECT * FROM custom_field_definition WHERE id = $1 AND company_id = $2;

-- name: ListCustomFieldDefinitions :many
SELECT * FROM custom_field_definition
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.narg(resource)::varchar IS NULL OR resource = sqlc.narg(resource))
ORDER BY resource, position, id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountCustomFieldDefinitions :one
SELECT count(*) FROM custom_field_definition
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.narg(resource)::varchar IS NULL OR resource = sqlc.narg(resource));

-- DefinitionsForResource backs validation on every write to a resource that
-- carries custom_fields. It is unpaginated on purpose: validation must see
-- every definition, and a page boundary would silently stop enforcing the ones
-- past it.
-- name: DefinitionsForResource :many
SELECT * FROM custom_field_definition
WHERE company_id = sqlc.arg(company_id) AND resource = sqlc.arg(resource)
ORDER BY position, id;

-- name: CreateCustomFieldDefinition :one
INSERT INTO custom_field_definition (company_id, resource, key, label, field_type, required, options, position, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
RETURNING *;

-- The key is immutable: it is what every stored document is keyed by, so
-- renaming it would orphan the values already written under the old one.
-- name: UpdateCustomFieldDefinition :one
UPDATE custom_field_definition SET
    label = sqlc.arg(label), field_type = sqlc.arg(field_type), required = sqlc.arg(required),
    options = sqlc.arg(options), position = sqlc.arg(position), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id)
RETURNING *;

-- Deleting a definition stops validating that key; the values already stored
-- under it stay in their documents, because they are the tenant's data and
-- nothing else records them.
-- name: DeleteCustomFieldDefinition :exec
DELETE FROM custom_field_definition WHERE id = $1 AND company_id = $2;
