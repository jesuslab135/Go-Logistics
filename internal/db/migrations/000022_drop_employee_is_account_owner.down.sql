-- 000022_drop_employee_is_account_owner.down.sql
ALTER TABLE employee ADD COLUMN is_account_owner boolean NOT NULL DEFAULT false;

-- Restore the flag for account owners, the one meaning it still carried.
UPDATE employee e SET is_account_owner = true
FROM account a
WHERE a.owner_employee_id = e.id;
