-- name: ListNotifications :many
SELECT * FROM notification
WHERE employee_id = $1 AND company_id = $2
ORDER BY created_at DESC, id DESC
LIMIT $3 OFFSET $4;

-- name: CountUnreadNotifications :one
SELECT count(*) FROM notification
WHERE employee_id = $1 AND company_id = $2 AND read_at IS NULL;

-- name: MarkNotificationRead :one
-- COALESCE keeps the transition idempotent: re-reading preserves the original
-- timestamp. The recipient predicate means a caller can only mark their own.
UPDATE notification SET read_at = COALESCE(read_at, sqlc.arg(read_at))
WHERE id = sqlc.arg(id) AND employee_id = sqlc.arg(employee_id) AND company_id = sqlc.arg(company_id)
RETURNING *;

-- name: CreateNotification :exec
INSERT INTO notification (company_id, employee_id, kind, title, body, url, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: NotifyTireApprovers :exec
-- Fans out to everyone who could act on the request: an admin, or a role that
-- grants tire_approvals/approve. An empty module object is a whole-module
-- grant, matching middleware.DecodePermissions.
INSERT INTO notification (company_id, employee_id, kind, title, body, url, created_at)
SELECT sqlc.arg(company_id), e.id, sqlc.arg(kind), sqlc.arg(title), sqlc.arg(body), sqlc.arg(url), sqlc.arg(created_at)
FROM employee e
JOIN employee_companies ec ON ec.employee_id = e.id AND ec.company_id = sqlc.arg(company_id)
LEFT JOIN role r ON r.id = ec.role_id
WHERE e.is_active
  AND (
    COALESCE(r.is_admin, false)
    OR COALESCE(r.permissions -> 'tire_approvals' ->> 'approve', 'false') = 'true'
    OR (r.permissions ? 'tire_approvals' AND r.permissions -> 'tire_approvals' = '{}'::jsonb)
  );

-- PruneNotifications drops what nobody will read again.
--
-- It is called opportunistically from the list endpoint rather than by a
-- scheduler: this deployment is a single API container with nothing to run a
-- cron in, and adding one to delete rows would be disproportionate. The bound
-- caps the work any single request does, so a long-neglected inbox is trimmed
-- over several visits instead of stalling one.
-- name: PruneNotifications :exec
DELETE FROM notification
WHERE id IN (
    SELECT n.id FROM notification n
    WHERE n.created_at < sqlc.arg(before)
    LIMIT sqlc.arg(lim)
);

-- NotifyWorkOrderAssignee tells one person that work is now theirs.
--
-- Only the assignee, not the watchers: on a busy asset, watcher fan-out turns
-- the bell into noise, and the assignment is news to exactly one person.
-- Nothing is sent when a work order is assigned to nobody, or to the person who
-- assigned it - being told about your own action is not a notification.
-- name: NotifyWorkOrderAssignee :exec
INSERT INTO notification (company_id, employee_id, kind, title, body, url, created_at)
SELECT sqlc.arg(company_id), sqlc.arg(employee_id), sqlc.arg(kind), sqlc.arg(title), sqlc.arg(body), sqlc.arg(url), sqlc.arg(created_at)
WHERE EXISTS (
    SELECT 1 FROM employee_companies ec
    WHERE ec.employee_id = sqlc.arg(employee_id) AND ec.company_id = sqlc.arg(company_id)
);
