-- name: GetIssuePriority :one
SELECT * FROM issue_priority WHERE id = $1 AND company_id = $2;

-- name: ListIssuePriorities :many
SELECT * FROM issue_priority WHERE company_id = $1 ORDER BY position LIMIT $2 OFFSET $3;

-- name: CountIssuePriorities :one
SELECT count(*) FROM issue_priority WHERE company_id = $1;

-- name: CreateIssuePriority :one
INSERT INTO issue_priority (
    company_id, name, color, position
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateIssuePriority :one
UPDATE issue_priority SET name = $3, color = $4, position = $5
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteIssuePriority :exec
DELETE FROM issue_priority WHERE id = $1 AND company_id = $2;
