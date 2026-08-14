-- name: GetComment :one
SELECT * FROM comment WHERE id = $1 AND company_id = $2;

-- name: ListComments :many
SELECT * FROM comment WHERE company_id = $1 ORDER BY created_at DESC, id LIMIT $2 OFFSET $3;

-- name: CountComments :one
SELECT count(*) FROM comment WHERE company_id = $1;

-- name: CreateComment :one
INSERT INTO comment (
    company_id, content_type_id, object_id, body, author_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateComment :one
UPDATE comment SET content_type_id = $3, object_id = $4, body = $5, author_id = $6, updated_at = $7
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comment WHERE id = $1 AND company_id = $2;
