-- The company-scoped vocabulary for trailer classification, following the
-- fuel_type / measurement_unit pattern every other catalog in this system uses.

-- name: GetTrailerClassification :one
SELECT * FROM trailer_classification WHERE id = $1 AND company_id = $2;

-- name: ListTrailerClassifications :many
SELECT * FROM trailer_classification WHERE company_id = $1 ORDER BY position, name, id LIMIT $2 OFFSET $3;

-- name: CountTrailerClassifications :one
SELECT count(*) FROM trailer_classification WHERE company_id = $1;

-- name: CreateTrailerClassification :one
INSERT INTO trailer_classification (company_id, name, position)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateTrailerClassification :one
UPDATE trailer_classification SET name = sqlc.arg(name), position = sqlc.arg(position)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id)
RETURNING *;

-- name: DeleteTrailerClassification :exec
DELETE FROM trailer_classification WHERE id = $1 AND company_id = $2;

-- CountTrailerClassificationsByIDs is how a trailer write checks the values it
-- was given belong to the caller's company. Naming a classification from
-- another tenant would otherwise link a trailer to a vocabulary its own fleet
-- cannot see or edit.
-- name: CountTrailerClassificationsByIDs :one
SELECT count(*) FROM trailer_classification
WHERE company_id = sqlc.arg(company_id) AND id = ANY(sqlc.arg(ids)::bigint[]);
