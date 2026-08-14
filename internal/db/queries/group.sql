-- "group" is a reserved word in SQL, so every reference is quoted.
--
-- parent_id carries no foreign key in the schema, so nothing at the database
-- level stops a client naming another tenant's group — or an id that does not
-- exist at all. The writes below resolve the parent within the caller's company
-- and match no row when they cannot, which surfaces as 404 exactly like every
-- other cross-tenant reference in the API.
--
-- ancestry is derived from that parent rather than accepted from the client: it
-- is the materialised path of parent_id, so letting a caller set it
-- independently only creates a tree whose two representations disagree.
-- (Known limitation: re-parenting a group does not rewrite its descendants'
-- ancestry — that needs a subtree update this API does not yet expose.)

-- name: GetGroup :one
SELECT * FROM "group" WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id);

-- name: ListGroups :many
SELECT * FROM "group" WHERE company_id = sqlc.arg(company_id)
ORDER BY ancestry, name, id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountGroups :one
SELECT count(*) FROM "group" WHERE company_id = sqlc.arg(company_id);

-- name: CreateGroup :one
INSERT INTO "group" (
    company_id, name, parent_id, ancestry, is_default, created_at, updated_at
)
SELECT
    sqlc.arg(company_id),
    sqlc.arg(name),
    sqlc.narg(parent_id),
    COALESCE(
        (SELECT CASE WHEN p.ancestry = '' THEN p.id::text ELSE p.ancestry || '/' || p.id END
           FROM "group" p
          WHERE p.id = sqlc.narg(parent_id) AND p.company_id = sqlc.arg(company_id)),
        ''),
    sqlc.arg(is_default),
    sqlc.arg(created_at),
    sqlc.arg(updated_at)
WHERE sqlc.narg(parent_id)::bigint IS NULL
   OR EXISTS (
       SELECT 1 FROM "group" p
       WHERE p.id = sqlc.narg(parent_id) AND p.company_id = sqlc.arg(company_id)
   )
RETURNING *;

-- name: UpdateGroup :one
-- The parent must also not be the group itself nor one of its descendants,
-- which would detach the subtree into a cycle.
UPDATE "group" g SET
    name = sqlc.arg(name),
    parent_id = sqlc.narg(parent_id),
    ancestry = COALESCE(
        (SELECT CASE WHEN p.ancestry = '' THEN p.id::text ELSE p.ancestry || '/' || p.id END
           FROM "group" p
          WHERE p.id = sqlc.narg(parent_id) AND p.company_id = sqlc.arg(company_id)),
        ''),
    is_default = sqlc.arg(is_default),
    updated_at = sqlc.arg(updated_at)
WHERE g.id = sqlc.arg(id)
  AND g.company_id = sqlc.arg(company_id)
  AND (
      sqlc.narg(parent_id)::bigint IS NULL
      OR EXISTS (
          SELECT 1 FROM "group" p
          WHERE p.id = sqlc.narg(parent_id)
            AND p.company_id = sqlc.arg(company_id)
            AND p.id <> g.id
            AND ('/' || p.ancestry || '/') NOT LIKE ('%/' || g.id || '/%')
      )
  )
RETURNING *;

-- name: DeleteGroup :exec
DELETE FROM "group" WHERE id = sqlc.arg(id) AND company_id = sqlc.arg(company_id);
