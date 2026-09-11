-- 000023_membership_is_active.up.sql
-- Deactivation belongs to a membership, not a person. employee.is_active was
-- a single global flag, so deactivating someone in one company locked them out
-- of every company in the account, including ones the administrator doing it
-- cannot see.
ALTER TABLE employee_companies ADD COLUMN is_active boolean NOT NULL DEFAULT true;

-- Carry today's deactivations onto every membership, so nobody gains access.
-- employee.is_active itself is left as it is: it stays a hard switch the API
-- no longer turns off, and reactivating a membership turns it back on.
UPDATE employee_companies ec SET is_active = false
FROM employee e
WHERE e.id = ec.employee_id AND NOT e.is_active;
