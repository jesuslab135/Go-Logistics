-- name: GetInspectionSubmission :one
SELECT * FROM inspection_submission WHERE id = $1 AND company_id = $2;

-- name: ListInspectionSubmissions :many
SELECT * FROM inspection_submission WHERE company_id = $1 ORDER BY submitted_at DESC, id LIMIT $2 OFFSET $3;

-- name: CountInspectionSubmissions :one
SELECT count(*) FROM inspection_submission WHERE company_id = $1;

-- name: CreateInspectionSubmission :one
INSERT INTO inspection_submission (
    company_id, form_id, asset_id, submitted_by_id, started_at, submitted_at, duration_seconds, starting_latitude, starting_longitude, submitted_latitude, submitted_longitude, signature, odometer, total_items, failed_items_count, passed_items_count, comments_count, images_count, general_notes, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
)
RETURNING *;

-- name: UpdateInspectionSubmission :one
UPDATE inspection_submission SET form_id = $3, asset_id = $4, submitted_by_id = $5, started_at = $6, submitted_at = $7, duration_seconds = $8, starting_latitude = $9, starting_longitude = $10, submitted_latitude = $11, submitted_longitude = $12, signature = $13, odometer = $14, total_items = $15, failed_items_count = $16, passed_items_count = $17, comments_count = $18, images_count = $19, general_notes = $20
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteInspectionSubmission :exec
DELETE FROM inspection_submission WHERE id = $1 AND company_id = $2;
