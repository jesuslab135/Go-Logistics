-- name: GetTireAssignmentRequest :one
SELECT * FROM tire_assignment_request WHERE id = $1 AND company_id = $2;

-- name: ListTireAssignmentRequests :many
SELECT * FROM tire_assignment_request WHERE company_id = $1 ORDER BY requested_at DESC, id LIMIT $2 OFFSET $3;

-- name: CountTireAssignmentRequests :one
SELECT count(*) FROM tire_assignment_request WHERE company_id = $1;

-- name: CreateTireAssignmentRequest :one
INSERT INTO tire_assignment_request (
    company_id, tire_id, vehicle_id, position_code, state, requested_by_id, requested_at, approved_by_id, resolved_at, rejection_reason, notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: UpdateTireAssignmentRequest :one
-- state, approved_by_id and resolved_at are deliberately absent from the SET
-- list: they belong to ResolveTireAssignmentRequest below, which guards on
-- state = 'PENDING' and is reached only through the approve route's
-- tire_approvals/approve gate. Leaving them here made the ordinary update a
-- way to forge an approval, so the restriction lives in the statement rather
-- than in whichever handler happens to call it.
UPDATE tire_assignment_request SET tire_id = $3, vehicle_id = $4, position_code = $5, requested_by_id = $6, requested_at = $7, rejection_reason = $8, notes = $9
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteTireAssignmentRequest :exec
DELETE FROM tire_assignment_request WHERE id = $1 AND company_id = $2;

-- name: GetTireAssignmentRequestForUpdate :one
-- Locks the request row so two concurrent approvals cannot both see PENDING.
SELECT * FROM tire_assignment_request WHERE id = $1 AND company_id = $2 FOR UPDATE;

-- name: ResolveTireAssignmentRequest :one
-- The state guard makes the resolution itself the idempotency barrier: a retry
-- matches no row and the caller learns the request was already resolved.
UPDATE tire_assignment_request
SET state = sqlc.arg(state), approved_by_id = sqlc.arg(approved_by_id), resolved_at = sqlc.arg(resolved_at)
WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id) AND state = 'PENDING'
RETURNING *;
