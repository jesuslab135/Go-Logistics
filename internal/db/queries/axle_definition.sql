-- name: ListAxleDefinitions :many
SELECT c.* FROM axle_definition c JOIN axle_template p ON p.id = c.template_id
WHERE c.template_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.position_index, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountAxleDefinitions :one
SELECT count(*) FROM axle_definition c JOIN axle_template p ON p.id = c.template_id WHERE c.template_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetAxleDefinition :one
SELECT c.* FROM axle_definition c JOIN axle_template p ON p.id = c.template_id
WHERE c.id = sqlc.arg(id) AND c.template_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateAxleDefinition :one
INSERT INTO axle_definition (
    template_id, position_index, label, axle_role, positions_per_side
)
SELECT sqlc.arg(parent_id), sqlc.arg(position_index), sqlc.arg(label), sqlc.arg(axle_role), sqlc.arg(positions_per_side)
WHERE EXISTS (SELECT 1 FROM axle_template WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateAxleDefinition :one
UPDATE axle_definition AS c SET position_index = sqlc.arg(position_index), label = sqlc.arg(label), axle_role = sqlc.arg(axle_role), positions_per_side = sqlc.arg(positions_per_side)
FROM axle_template p
WHERE c.id = sqlc.arg(id) AND c.template_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.template_id
RETURNING c.*;

-- name: DeleteAxleDefinition :exec
DELETE FROM axle_definition AS c USING axle_template p
WHERE c.id = sqlc.arg(id) AND c.template_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.template_id;
