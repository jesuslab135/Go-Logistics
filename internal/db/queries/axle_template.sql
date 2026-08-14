-- name: GetAxleTemplate :one
SELECT * FROM axle_template WHERE id = $1 AND company_id = $2;

-- name: ListAxleTemplates :many
SELECT * FROM axle_template WHERE company_id = $1 ORDER BY name, id LIMIT $2 OFFSET $3;

-- name: CountAxleTemplates :one
SELECT count(*) FROM axle_template WHERE company_id = $1;

-- name: CreateAxleTemplate :one
INSERT INTO axle_template (
    company_id, name, description, total_positions
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateAxleTemplate :one
UPDATE axle_template SET name = $3, description = $4, total_positions = $5
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteAxleTemplate :exec
DELETE FROM axle_template WHERE id = $1 AND company_id = $2;
