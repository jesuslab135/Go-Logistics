-- 000021_role_on_membership.down.sql
ALTER TABLE employee ADD COLUMN role_id bigint;

-- Restore a single global role per employee from their memberships. Which one
-- is arbitrary when they differ, because the old model could not express more
-- than one — that information is genuinely lost on the way down.
UPDATE employee e SET role_id = (
    SELECT ec.role_id FROM employee_companies ec
    WHERE ec.employee_id = e.id AND ec.role_id IS NOT NULL
    ORDER BY ec.id LIMIT 1
);

ALTER TABLE employee ADD CONSTRAINT fk_employee_role
    FOREIGN KEY (role_id) REFERENCES role(id) ON DELETE RESTRICT;

DROP INDEX IF EXISTS idx_employee_companies_role;
ALTER TABLE employee_companies DROP CONSTRAINT fk_ec_role_company;
ALTER TABLE role DROP CONSTRAINT uq_role_company;
ALTER TABLE employee_companies DROP COLUMN role_id;
