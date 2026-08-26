-- is_primary is not settable through create or update. At most one photo per
-- entry may carry it (a partial unique index enforces that), so setting it is
-- an action that also clears the others - two writes that have to happen
-- together, which a field on a whole-record update cannot express.

-- name: ListFuelPhotos :many
SELECT c.* FROM fuel_photo c JOIN fuel_entry p0 ON p0.id = c.entry_id JOIN asset p1 ON p1.id = p0.asset_id
WHERE c.entry_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id)
ORDER BY c.uploaded_at DESC, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountFuelPhotos :one
SELECT count(*) FROM fuel_photo c JOIN fuel_entry p0 ON p0.id = c.entry_id JOIN asset p1 ON p1.id = p0.asset_id WHERE c.entry_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id);

-- name: GetFuelPhoto :one
SELECT c.* FROM fuel_photo c JOIN fuel_entry p0 ON p0.id = c.entry_id JOIN asset p1 ON p1.id = p0.asset_id
WHERE c.id = sqlc.arg(id) AND c.entry_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id);

-- name: CreateFuelPhoto :one
INSERT INTO fuel_photo (
    entry_id, uploaded_by_id, file, file_name, file_size, mime_type, description, uploaded_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(uploaded_by_id), sqlc.arg(file), sqlc.arg(file_name), sqlc.arg(file_size), sqlc.arg(mime_type), sqlc.arg(description), sqlc.arg(uploaded_at)
WHERE EXISTS (SELECT 1 FROM fuel_entry p0 JOIN asset p1 ON p1.id = p0.asset_id WHERE p0.id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateFuelPhoto :one
-- Upload attribution is stamped once and never reassigned by an edit.
UPDATE fuel_photo AS c SET file = sqlc.arg(file), file_name = sqlc.arg(file_name), file_size = sqlc.arg(file_size), mime_type = sqlc.arg(mime_type), description = sqlc.arg(description), uploaded_at = sqlc.arg(uploaded_at)
FROM fuel_entry p0, asset p1
WHERE c.id = sqlc.arg(id) AND c.entry_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id) AND p0.id = c.entry_id AND p1.id = p0.asset_id
RETURNING c.*;

-- name: DeleteFuelPhoto :exec
DELETE FROM fuel_photo AS c USING fuel_entry p0, asset p1
WHERE c.id = sqlc.arg(id) AND c.entry_id = sqlc.arg(parent_id) AND p1.company_id = sqlc.arg(company_id) AND p0.id = c.entry_id AND p1.id = p0.asset_id;

-- ClearFuelEntryPrimaryPhoto demotes whatever currently holds the flag, so the
-- following promotion cannot collide with the unique index.
-- name: ClearFuelEntryPrimaryPhoto :exec
UPDATE fuel_photo c SET is_primary = false
FROM fuel_entry p0, asset p1
WHERE c.entry_id = sqlc.arg(parent_id) AND c.is_primary
  AND p0.id = c.entry_id AND p1.id = p0.asset_id AND p1.company_id = sqlc.arg(company_id);

-- The tenant scope runs through the entry's asset, as every other fuel_photo
-- query here does: fuel_entry has no company of its own.
-- name: SetFuelPhotoPrimary :one
UPDATE fuel_photo AS c SET is_primary = true
FROM fuel_entry p0, asset p1
WHERE c.id = sqlc.arg(id) AND c.entry_id = sqlc.arg(parent_id)
  AND p0.id = c.entry_id AND p1.id = p0.asset_id AND p1.company_id = sqlc.arg(company_id)
RETURNING c.*;
