-- 000022_drop_employee_is_account_owner.up.sql
-- employee.is_account_owner was the old global "may create companies" flag.
-- Since 000019 ownership is account.owner_employee_id, and no permission check
-- reads this column. The employee form and /admin/companies/{id}/set-owner
-- still wrote it, which produced a toggle that looked like ownership and
-- granted nothing. Dropped so it cannot be mistaken for ownership again.
ALTER TABLE employee DROP COLUMN is_account_owner;
