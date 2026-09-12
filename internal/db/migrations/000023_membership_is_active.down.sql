-- 000023_membership_is_active.down.sql
-- A person whose memberships all exist but none is active was deactivated
-- everywhere, which is what the global flag expressed.
UPDATE employee e SET is_active = false
WHERE EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.employee_id = e.id)
  AND NOT EXISTS (SELECT 1 FROM employee_companies ec WHERE ec.employee_id = e.id AND ec.is_active);

ALTER TABLE employee_companies DROP COLUMN is_active;
