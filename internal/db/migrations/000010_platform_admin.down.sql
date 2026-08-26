-- 000010_platform_admin.down.sql
-- Rolling back drops the audit trail, which cannot be reconstructed. Export it
-- first if the grants it records still matter.
DROP TABLE IF EXISTS membership_audit;
ALTER TABLE employee DROP COLUMN IF EXISTS is_platform_admin;
