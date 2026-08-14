-- name: ListWheelPositionDefinitions :many
SELECT c.* FROM wheel_position_definition c JOIN axle_definition p0 ON p0.id = c.axle_id JOIN axle_template p1 ON p1.id = p0.template_id
WHERE c.axle_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id)
ORDER BY c.slot, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountWheelPositionDefinitions :one
SELECT count(*) FROM wheel_position_definition c JOIN axle_definition p0 ON p0.id = c.axle_id JOIN axle_template p1 ON p1.id = p0.template_id WHERE c.axle_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id);

-- name: GetWheelPositionDefinition :one
SELECT c.* FROM wheel_position_definition c JOIN axle_definition p0 ON p0.id = c.axle_id JOIN axle_template p1 ON p1.id = p0.template_id
WHERE c.id = sqlc.arg(id) AND c.axle_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id);

-- name: CreateWheelPositionDefinition :one
INSERT INTO wheel_position_definition (
    axle_id, code, side, slot
)
SELECT sqlc.arg(parent_id), sqlc.arg(code), sqlc.arg(side), sqlc.arg(slot)
WHERE EXISTS (SELECT 1 FROM axle_definition p0 JOIN axle_template p1 ON p1.id = p0.template_id WHERE p0.id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateWheelPositionDefinition :one
UPDATE wheel_position_definition AS c SET code = sqlc.arg(code), side = sqlc.arg(side), slot = sqlc.arg(slot)
FROM axle_definition p0, axle_template p1
WHERE c.id = sqlc.arg(id) AND c.axle_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id) AND p0.id = c.axle_id AND p1.id = p0.template_id
RETURNING c.*;

-- name: DeleteWheelPositionDefinition :exec
DELETE FROM wheel_position_definition AS c USING axle_definition p0, axle_template p1
WHERE c.id = sqlc.arg(id) AND c.axle_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id) AND p0.id = c.axle_id AND p1.id = p0.template_id;
