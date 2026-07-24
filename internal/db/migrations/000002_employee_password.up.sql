-- 000002_employee_password.up.sql
-- Adds local credential storage to employee. Django's auth_user was not ported,
-- so passwords live here as a bcrypt hash. Empty string = no password set, which
-- can never match (bcrypt rejects it), so such accounts cannot log in.

ALTER TABLE employee ADD COLUMN password_hash varchar(255) NOT NULL DEFAULT '';
