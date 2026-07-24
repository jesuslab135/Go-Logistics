-- name: GetIssue :one
SELECT * FROM issue WHERE id = $1 AND company_id = $2;

-- name: ListIssues :many
SELECT * FROM issue WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountIssues :one
SELECT count(*) FROM issue WHERE company_id = $1;

-- name: CreateIssue :one
INSERT INTO issue (
    company_id, number, asset_id, asset_type, name, summary, description, state, priority_id, fault_id, source_type, inspection_submission_id, reported_at, reported_by_id, due_date, due_meter_value, due_secondary_meter_value, overdue, resolved_at, resolved_by_id, resolution_note, reopened_at, reopened_by_id, resolvable_type, resolvable_id, closed_at, closed_by_id, closed_note, external_id, created_by_workflow, comments_count, images_count, documents_count, labels, custom_fields, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37
)
RETURNING *;

-- name: UpdateIssue :one
UPDATE issue SET number = $3, asset_id = $4, asset_type = $5, name = $6, summary = $7, description = $8, state = $9, priority_id = $10, fault_id = $11, source_type = $12, inspection_submission_id = $13, reported_at = $14, reported_by_id = $15, due_date = $16, due_meter_value = $17, due_secondary_meter_value = $18, overdue = $19, resolved_at = $20, resolved_by_id = $21, resolution_note = $22, reopened_at = $23, reopened_by_id = $24, resolvable_type = $25, resolvable_id = $26, closed_at = $27, closed_by_id = $28, closed_note = $29, external_id = $30, created_by_workflow = $31, comments_count = $32, images_count = $33, documents_count = $34, labels = $35, custom_fields = $36, updated_at = $37
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteIssue :exec
DELETE FROM issue WHERE id = $1 AND company_id = $2;
