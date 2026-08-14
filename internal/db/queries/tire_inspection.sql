-- name: ListTireInspections :many
SELECT c.* FROM tire_inspection c JOIN tire p ON p.id = c.tire_id
WHERE c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.inspection_date DESC, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountTireInspections :one
SELECT count(*) FROM tire_inspection c JOIN tire p ON p.id = c.tire_id WHERE c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetTireInspection :one
SELECT c.* FROM tire_inspection c JOIN tire p ON p.id = c.tire_id
WHERE c.id = sqlc.arg(id) AND c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateTireInspection :one
INSERT INTO tire_inspection (
    tire_id, inspection_date, odometer, tread_depth_32nds, psi, measured_by_id, notes
)
SELECT sqlc.arg(parent_id), sqlc.arg(inspection_date), sqlc.arg(odometer), sqlc.arg(tread_depth_32nds), sqlc.arg(psi), sqlc.arg(measured_by_id), sqlc.arg(notes)
WHERE EXISTS (SELECT 1 FROM tire WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateTireInspection :one
UPDATE tire_inspection AS c SET inspection_date = sqlc.arg(inspection_date), odometer = sqlc.arg(odometer), tread_depth_32nds = sqlc.arg(tread_depth_32nds), psi = sqlc.arg(psi), measured_by_id = sqlc.arg(measured_by_id), notes = sqlc.arg(notes)
FROM tire p
WHERE c.id = sqlc.arg(id) AND c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.tire_id
RETURNING c.*;

-- name: DeleteTireInspection :exec
DELETE FROM tire_inspection AS c USING tire p
WHERE c.id = sqlc.arg(id) AND c.tire_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.tire_id;
