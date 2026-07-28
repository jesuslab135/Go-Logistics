-- Join-table ("link") queries for the m2m relationships Django exposed through
-- nested serializers. Each set follows the same four-query shape:
--
--   List/Count  read the linked entities, gated on the parent's company.
--   Find...Candidate resolves the target and doubles as the authorization check:
--     it returns a row only when BOTH the parent and the target belong to the
--     caller's company, so linking across tenants is a 404, not a silent write.
--   Add/Remove  perform the write; Add is idempotent via the unique constraint.

-- ---------------------------------------------------------------------------
-- issue_assigned_to
-- ---------------------------------------------------------------------------

-- name: ListIssueAssignees :many
SELECT e.* FROM employee e
JOIN issue_assigned_to ia ON ia.employee_id = e.id
JOIN issue i ON i.id = ia.issue_id
WHERE ia.issue_id = sqlc.arg(issue_id) AND i.company_id = sqlc.arg(company_id)
ORDER BY e.last_name, e.first_name
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountIssueAssignees :one
SELECT count(*) FROM issue_assigned_to ia
JOIN issue i ON i.id = ia.issue_id
WHERE ia.issue_id = sqlc.arg(issue_id) AND i.company_id = sqlc.arg(company_id);

-- name: FindIssueAssigneeCandidate :one
SELECT e.* FROM employee e
WHERE e.id = sqlc.arg(employee_id)
  AND EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.employee_id = e.id AND ec.company_id = sqlc.arg(company_id))
  AND EXISTS (SELECT 1 FROM issue i WHERE i.id = sqlc.arg(issue_id) AND i.company_id = sqlc.arg(company_id));

-- name: AddIssueAssignee :exec
INSERT INTO issue_assigned_to (issue_id, employee_id)
VALUES (sqlc.arg(issue_id), sqlc.arg(employee_id))
ON CONFLICT (issue_id, employee_id) DO NOTHING;

-- name: RemoveIssueAssignee :exec
DELETE FROM issue_assigned_to ia
USING issue i
WHERE i.id = ia.issue_id
  AND ia.issue_id = sqlc.arg(issue_id)
  AND ia.employee_id = sqlc.arg(employee_id)
  AND i.company_id = sqlc.arg(company_id);

-- ---------------------------------------------------------------------------
-- issue_watchers
-- ---------------------------------------------------------------------------

-- name: ListIssueWatchers :many
SELECT e.* FROM employee e
JOIN issue_watchers iw ON iw.employee_id = e.id
JOIN issue i ON i.id = iw.issue_id
WHERE iw.issue_id = sqlc.arg(issue_id) AND i.company_id = sqlc.arg(company_id)
ORDER BY e.last_name, e.first_name
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountIssueWatchers :one
SELECT count(*) FROM issue_watchers iw
JOIN issue i ON i.id = iw.issue_id
WHERE iw.issue_id = sqlc.arg(issue_id) AND i.company_id = sqlc.arg(company_id);

-- name: FindIssueWatcherCandidate :one
SELECT e.* FROM employee e
WHERE e.id = sqlc.arg(employee_id)
  AND EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.employee_id = e.id AND ec.company_id = sqlc.arg(company_id))
  AND EXISTS (SELECT 1 FROM issue i WHERE i.id = sqlc.arg(issue_id) AND i.company_id = sqlc.arg(company_id));

-- name: AddIssueWatcher :exec
INSERT INTO issue_watchers (issue_id, employee_id)
VALUES (sqlc.arg(issue_id), sqlc.arg(employee_id))
ON CONFLICT (issue_id, employee_id) DO NOTHING;

-- name: RemoveIssueWatcher :exec
DELETE FROM issue_watchers iw
USING issue i
WHERE i.id = iw.issue_id
  AND iw.issue_id = sqlc.arg(issue_id)
  AND iw.employee_id = sqlc.arg(employee_id)
  AND i.company_id = sqlc.arg(company_id);

-- ---------------------------------------------------------------------------
-- work_order_issues
-- ---------------------------------------------------------------------------

-- name: ListWorkOrderIssues :many
SELECT i.* FROM issue i
JOIN work_order_issues woi ON woi.issue_id = i.id
JOIN work_order wo ON wo.id = woi.work_order_id
WHERE woi.work_order_id = sqlc.arg(work_order_id) AND wo.company_id = sqlc.arg(company_id)
ORDER BY i.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountWorkOrderIssues :one
SELECT count(*) FROM work_order_issues woi
JOIN work_order wo ON wo.id = woi.work_order_id
WHERE woi.work_order_id = sqlc.arg(work_order_id) AND wo.company_id = sqlc.arg(company_id);

-- name: FindWorkOrderIssueCandidate :one
SELECT i.* FROM issue i
WHERE i.id = sqlc.arg(issue_id) AND i.company_id = sqlc.arg(company_id)
  AND EXISTS (SELECT 1 FROM work_order wo WHERE wo.id = sqlc.arg(work_order_id) AND wo.company_id = sqlc.arg(company_id));

-- name: AddWorkOrderIssue :exec
INSERT INTO work_order_issues (work_order_id, issue_id)
VALUES (sqlc.arg(work_order_id), sqlc.arg(issue_id))
ON CONFLICT (work_order_id, issue_id) DO NOTHING;

-- name: RemoveWorkOrderIssue :exec
DELETE FROM work_order_issues woi
USING work_order wo
WHERE wo.id = woi.work_order_id
  AND woi.work_order_id = sqlc.arg(work_order_id)
  AND woi.issue_id = sqlc.arg(issue_id)
  AND wo.company_id = sqlc.arg(company_id);

-- ---------------------------------------------------------------------------
-- work_order_faults
-- ---------------------------------------------------------------------------

-- name: ListWorkOrderFaults :many
SELECT f.* FROM fault f
JOIN work_order_faults wof ON wof.fault_id = f.id
JOIN work_order wo ON wo.id = wof.work_order_id
WHERE wof.work_order_id = sqlc.arg(work_order_id) AND wo.company_id = sqlc.arg(company_id)
ORDER BY f.code, f.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountWorkOrderFaults :one
SELECT count(*) FROM work_order_faults wof
JOIN work_order wo ON wo.id = wof.work_order_id
WHERE wof.work_order_id = sqlc.arg(work_order_id) AND wo.company_id = sqlc.arg(company_id);

-- name: FindWorkOrderFaultCandidate :one
SELECT f.* FROM fault f
WHERE f.id = sqlc.arg(fault_id) AND f.company_id = sqlc.arg(company_id)
  AND EXISTS (SELECT 1 FROM work_order wo WHERE wo.id = sqlc.arg(work_order_id) AND wo.company_id = sqlc.arg(company_id));

-- name: AddWorkOrderFault :exec
INSERT INTO work_order_faults (work_order_id, fault_id)
VALUES (sqlc.arg(work_order_id), sqlc.arg(fault_id))
ON CONFLICT (work_order_id, fault_id) DO NOTHING;

-- name: RemoveWorkOrderFault :exec
DELETE FROM work_order_faults wof
USING work_order wo
WHERE wo.id = wof.work_order_id
  AND wof.work_order_id = sqlc.arg(work_order_id)
  AND wof.fault_id = sqlc.arg(fault_id)
  AND wo.company_id = sqlc.arg(company_id);

-- ---------------------------------------------------------------------------
-- service_entry_line_item_issues (parent is two hops from the company)
-- ---------------------------------------------------------------------------

-- name: ListServiceEntryLineItemIssues :many
SELECT i.* FROM issue i
JOIN service_entry_line_item_issues slii ON slii.issue_id = i.id
JOIN service_entry_line_item li ON li.id = slii.service_entry_line_item_id
JOIN service_entry se ON se.id = li.service_entry_id
WHERE slii.service_entry_line_item_id = sqlc.arg(line_item_id) AND se.company_id = sqlc.arg(company_id)
ORDER BY i.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountServiceEntryLineItemIssues :one
SELECT count(*) FROM service_entry_line_item_issues slii
JOIN service_entry_line_item li ON li.id = slii.service_entry_line_item_id
JOIN service_entry se ON se.id = li.service_entry_id
WHERE slii.service_entry_line_item_id = sqlc.arg(line_item_id) AND se.company_id = sqlc.arg(company_id);

-- name: FindServiceEntryLineItemIssueCandidate :one
SELECT i.* FROM issue i
WHERE i.id = sqlc.arg(issue_id) AND i.company_id = sqlc.arg(company_id)
  AND EXISTS (
      SELECT 1 FROM service_entry_line_item li
      JOIN service_entry se ON se.id = li.service_entry_id
      WHERE li.id = sqlc.arg(line_item_id) AND se.company_id = sqlc.arg(company_id)
  );

-- name: AddServiceEntryLineItemIssue :exec
INSERT INTO service_entry_line_item_issues (service_entry_line_item_id, issue_id)
VALUES (sqlc.arg(line_item_id), sqlc.arg(issue_id))
ON CONFLICT (service_entry_line_item_id, issue_id) DO NOTHING;

-- name: RemoveServiceEntryLineItemIssue :exec
DELETE FROM service_entry_line_item_issues slii
USING service_entry_line_item li, service_entry se
WHERE li.id = slii.service_entry_line_item_id
  AND se.id = li.service_entry_id
  AND slii.service_entry_line_item_id = sqlc.arg(line_item_id)
  AND slii.issue_id = sqlc.arg(issue_id)
  AND se.company_id = sqlc.arg(company_id);

-- ---------------------------------------------------------------------------
-- work_order_line_item_issues (parent is two hops from the company)
-- ---------------------------------------------------------------------------

-- name: ListWorkOrderLineItemIssues :many
SELECT i.* FROM issue i
JOIN work_order_line_item_issues wlii ON wlii.issue_id = i.id
JOIN work_order_line_item li ON li.id = wlii.work_order_line_item_id
JOIN work_order wo ON wo.id = li.work_order_id
WHERE wlii.work_order_line_item_id = sqlc.arg(line_item_id) AND wo.company_id = sqlc.arg(company_id)
ORDER BY i.id
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountWorkOrderLineItemIssues :one
SELECT count(*) FROM work_order_line_item_issues wlii
JOIN work_order_line_item li ON li.id = wlii.work_order_line_item_id
JOIN work_order wo ON wo.id = li.work_order_id
WHERE wlii.work_order_line_item_id = sqlc.arg(line_item_id) AND wo.company_id = sqlc.arg(company_id);

-- name: FindWorkOrderLineItemIssueCandidate :one
SELECT i.* FROM issue i
WHERE i.id = sqlc.arg(issue_id) AND i.company_id = sqlc.arg(company_id)
  AND EXISTS (
      SELECT 1 FROM work_order_line_item li
      JOIN work_order wo ON wo.id = li.work_order_id
      WHERE li.id = sqlc.arg(line_item_id) AND wo.company_id = sqlc.arg(company_id)
  );

-- name: AddWorkOrderLineItemIssue :exec
INSERT INTO work_order_line_item_issues (work_order_line_item_id, issue_id)
VALUES (sqlc.arg(line_item_id), sqlc.arg(issue_id))
ON CONFLICT (work_order_line_item_id, issue_id) DO NOTHING;

-- name: RemoveWorkOrderLineItemIssue :exec
DELETE FROM work_order_line_item_issues wlii
USING work_order_line_item li, work_order wo
WHERE li.id = wlii.work_order_line_item_id
  AND wo.id = li.work_order_id
  AND wlii.work_order_line_item_id = sqlc.arg(line_item_id)
  AND wlii.issue_id = sqlc.arg(issue_id)
  AND wo.company_id = sqlc.arg(company_id);
