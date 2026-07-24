-- name: GetTireAssignmentRequest :one
SELECT * FROM tire_assignment_request WHERE id = $1 AND company_id = $2;

-- name: ListTireAssignmentRequests :many
SELECT * FROM tire_assignment_request WHERE company_id = $1 ORDER BY requested_at DESC LIMIT $2 OFFSET $3;

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
UPDATE tire_assignment_request SET tire_id = $3, vehicle_id = $4, position_code = $5, state = $6, requested_by_id = $7, requested_at = $8, approved_by_id = $9, resolved_at = $10, rejection_reason = $11, notes = $12
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteTireAssignmentRequest :exec
DELETE FROM tire_assignment_request WHERE id = $1 AND company_id = $2;
