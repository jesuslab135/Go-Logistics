-- name: ListFuelComments :many
SELECT c.* FROM fuel_comment c JOIN fuel_entry p0 ON p0.id = c.entry_id JOIN asset p1 ON p1.id = p0.asset_id
WHERE c.entry_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id)
ORDER BY c.created_at DESC LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountFuelComments :one
SELECT count(*) FROM fuel_comment c JOIN fuel_entry p0 ON p0.id = c.entry_id JOIN asset p1 ON p1.id = p0.asset_id WHERE c.entry_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id);

-- name: GetFuelComment :one
SELECT c.* FROM fuel_comment c JOIN fuel_entry p0 ON p0.id = c.entry_id JOIN asset p1 ON p1.id = p0.asset_id
WHERE c.id = sqlc.arg(id) AND c.entry_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id);

-- name: CreateFuelComment :one
INSERT INTO fuel_comment (
    entry_id, employee_id, text, created_at, updated_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(employee_id), sqlc.arg(text), sqlc.arg(created_at), sqlc.arg(updated_at)
WHERE EXISTS (SELECT 1 FROM fuel_entry p0 JOIN asset p1 ON p1.id = p0.asset_id WHERE p0.id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateFuelComment :one
-- Authorship is stamped once at creation and never reassigned by an edit.
UPDATE fuel_comment AS c SET text = sqlc.arg(text), updated_at = sqlc.arg(updated_at)
FROM fuel_entry p0, asset p1
WHERE c.id = sqlc.arg(id) AND c.entry_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id) AND p0.id = c.entry_id AND p1.id = p0.asset_id
RETURNING c.*;

-- name: DeleteFuelComment :exec
DELETE FROM fuel_comment AS c USING fuel_entry p0, asset p1
WHERE c.id = sqlc.arg(id) AND c.entry_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id) AND p0.id = c.entry_id AND p1.id = p0.asset_id;
