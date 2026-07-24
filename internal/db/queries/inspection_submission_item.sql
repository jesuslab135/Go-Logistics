-- name: ListInspectionSubmissionItems :many
SELECT c.* FROM inspection_submission_item c JOIN inspection_submission p ON p.id = c.submission_id
WHERE c.submission_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountInspectionSubmissionItems :one
SELECT count(*) FROM inspection_submission_item c JOIN inspection_submission p ON p.id = c.submission_id WHERE c.submission_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetInspectionSubmissionItem :one
SELECT c.* FROM inspection_submission_item c JOIN inspection_submission p ON p.id = c.submission_id
WHERE c.id = sqlc.arg(id) AND c.submission_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: CreateInspectionSubmissionItem :one
INSERT INTO inspection_submission_item (
    submission_id, form_item_id, result_status, result_value, remark, photo, latitude, longitude, generated_issue_id
)
SELECT sqlc.arg(parent_id), sqlc.arg(form_item_id), sqlc.arg(result_status), sqlc.arg(result_value), sqlc.arg(remark), sqlc.arg(photo), sqlc.arg(latitude), sqlc.arg(longitude), sqlc.arg(generated_issue_id)
WHERE EXISTS (SELECT 1 FROM inspection_submission WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateInspectionSubmissionItem :one
UPDATE inspection_submission_item AS c SET form_item_id = sqlc.arg(form_item_id), result_status = sqlc.arg(result_status), result_value = sqlc.arg(result_value), remark = sqlc.arg(remark), photo = sqlc.arg(photo), latitude = sqlc.arg(latitude), longitude = sqlc.arg(longitude), generated_issue_id = sqlc.arg(generated_issue_id)
FROM inspection_submission p
WHERE c.id = sqlc.arg(id) AND c.submission_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.submission_id
RETURNING c.*;

-- name: DeleteInspectionSubmissionItem :exec
DELETE FROM inspection_submission_item AS c USING inspection_submission p
WHERE c.id = sqlc.arg(id) AND c.submission_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.submission_id;
