-- name: GetComment :one
SELECT * FROM comment WHERE id = $1 AND company_id = $2;

-- name: ListComments :many
SELECT * FROM comment WHERE company_id = $1 ORDER BY created_at DESC, id LIMIT $2 OFFSET $3;

-- name: CountComments :one
SELECT count(*) FROM comment WHERE company_id = $1;

-- name: CommentParentExists :one
-- Confirms the parent exists AND belongs to the caller's company, so a comment
-- cannot be attached to another tenant's record. An unknown content type
-- yields false rather than an error, and is reported as a missing parent.
SELECT CASE sqlc.arg(content_type)::text
    WHEN 'asset'         THEN EXISTS(SELECT 1 FROM asset a          WHERE a.id  = sqlc.arg(object_id)::bigint AND a.company_id  = sqlc.arg(scope_company_id)::bigint)
    WHEN 'issue'         THEN EXISTS(SELECT 1 FROM issue i          WHERE i.id  = sqlc.arg(object_id)::bigint AND i.company_id  = sqlc.arg(scope_company_id)::bigint)
    WHEN 'work_order'    THEN EXISTS(SELECT 1 FROM work_order wo    WHERE wo.id = sqlc.arg(object_id)::bigint AND wo.company_id = sqlc.arg(scope_company_id)::bigint)
    WHEN 'service_entry' THEN EXISTS(SELECT 1 FROM service_entry se WHERE se.id = sqlc.arg(object_id)::bigint AND se.company_id = sqlc.arg(scope_company_id)::bigint)
    ELSE false
END::boolean AS ok;

-- name: CreateComment :one
INSERT INTO comment (
    company_id, content_type, object_id, body, author_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateComment :one
-- The parent and the author are fixed at creation; an edit changes only the body.
UPDATE comment SET body = $3, updated_at = $4
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comment WHERE id = $1 AND company_id = $2;
