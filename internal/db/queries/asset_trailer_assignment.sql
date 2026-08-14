-- The trailer's name is joined in so a registry row does not need a second
-- request per assignment just to render which trailer is attached.

-- name: ListAssetTrailerAssignments :many
SELECT c.*, t.name AS trailer_name FROM asset_trailer_assignment c
JOIN asset p ON p.id = c.asset_id
JOIN asset t ON t.id = c.trailer_id
WHERE c.asset_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.position, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountAssetTrailerAssignments :one
SELECT count(*) FROM asset_trailer_assignment c JOIN asset p ON p.id = c.asset_id WHERE c.asset_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetAssetTrailerAssignment :one
SELECT c.*, t.name AS trailer_name FROM asset_trailer_assignment c
JOIN asset p ON p.id = c.asset_id
JOIN asset t ON t.id = c.trailer_id
WHERE c.id = sqlc.arg(id) AND c.asset_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateAssetTrailerAssignment :one
INSERT INTO asset_trailer_assignment (
    asset_id, trailer_id, position, assigned_date, unassigned_date, assigned_by_id, is_active, notes
)
SELECT sqlc.arg(parent_id), sqlc.arg(trailer_id), sqlc.arg(position), sqlc.arg(assigned_date), sqlc.arg(unassigned_date), sqlc.arg(assigned_by_id), sqlc.arg(is_active), sqlc.arg(notes)
WHERE EXISTS (SELECT 1 FROM asset WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateAssetTrailerAssignment :one
UPDATE asset_trailer_assignment AS c SET trailer_id = sqlc.arg(trailer_id), position = sqlc.arg(position), assigned_date = sqlc.arg(assigned_date), unassigned_date = sqlc.arg(unassigned_date), assigned_by_id = sqlc.arg(assigned_by_id), is_active = sqlc.arg(is_active), notes = sqlc.arg(notes)
FROM asset p
WHERE c.id = sqlc.arg(id) AND c.asset_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.asset_id
RETURNING c.*;

-- name: DeleteAssetTrailerAssignment :exec
DELETE FROM asset_trailer_assignment AS c USING asset p
WHERE c.id = sqlc.arg(id) AND c.asset_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.asset_id;
