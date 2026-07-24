-- 000002_employee_password.down.sql

ALTER TABLE employee DROP COLUMN password_hash;
