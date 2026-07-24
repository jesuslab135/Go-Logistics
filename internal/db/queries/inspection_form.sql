-- name: GetInspectionForm :one
SELECT * FROM inspection_form WHERE id = $1 AND company_id = $2;

-- name: ListInspectionForms :many
SELECT * FROM inspection_form WHERE company_id = $1 ORDER BY title LIMIT $2 OFFSET $3;

-- name: CountInspectionForms :one
SELECT count(*) FROM inspection_form WHERE company_id = $1;

-- name: CreateInspectionForm :one
INSERT INTO inspection_form (
    company_id, title, description, version, require_live_photo, auto_create_issues, color, archived_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateInspectionForm :one
UPDATE inspection_form SET title = $3, description = $4, version = $5, require_live_photo = $6, auto_create_issues = $7, color = $8, archived_at = $9, updated_at = $10
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteInspectionForm :exec
DELETE FROM inspection_form WHERE id = $1 AND company_id = $2;
