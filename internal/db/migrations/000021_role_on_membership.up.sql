-- 000021_role_on_membership.up.sql
-- A role is a property of a membership, not of a person.
--
-- employee.role_id was a single global foreign key while role is company-scoped,
-- so a person whose role pointed at one company's administrator role was an
-- administrator in EVERY company they belonged to. Migration 000010 exists
-- because that reached further than it looked once already.
ALTER TABLE employee_companies ADD COLUMN role_id bigint;

-- A role from another company must be impossible to attach — not merely
-- rejected by whichever endpoint remembers to check. Same construction 000019
-- used for cross-account membership, and for the same reason.
ALTER TABLE role ADD CONSTRAINT uq_role_company UNIQUE (id, company_id);
ALTER TABLE employee_companies ADD CONSTRAINT fk_ec_role_company
    FOREIGN KEY (role_id, company_id) REFERENCES role(id, company_id) ON DELETE SET NULL;

-- Preserve effective access exactly.
--
-- Today a global admin role confers admin in every company the person belongs
-- to. Moving role_id onto memberships naively would leave them an administrator
-- only where that particular role lives, silently stripping access everywhere
-- else with no error and no indication anything changed.
--
-- So each membership takes the role of the SAME NAME in its own company, and
-- where none exists the definition is copied. Role names are not unique per
-- company (there is no UNIQUE (company_id, name)), so the lowest id wins and the
-- result is deterministic.
DO $$
DECLARE m RECORD;
        src RECORD;
        target_role_id bigint;
BEGIN
    FOR m IN
        SELECT ec.id AS ec_id, ec.company_id, e.role_id AS global_role_id
        FROM employee_companies ec
        JOIN employee e ON e.id = ec.employee_id
        WHERE e.role_id IS NOT NULL
    LOOP
        SELECT r.name, r.is_admin, r.permissions INTO src
        FROM role r WHERE r.id = m.global_role_id;

        SELECT r.id INTO target_role_id
        FROM role r
        WHERE r.company_id = m.company_id AND r.name = src.name
        ORDER BY r.id
        LIMIT 1;

        IF target_role_id IS NULL THEN
            INSERT INTO role (company_id, name, is_admin, permissions)
            VALUES (m.company_id, src.name, src.is_admin, src.permissions)
            RETURNING id INTO target_role_id;
        END IF;

        UPDATE employee_companies SET role_id = target_role_id WHERE id = m.ec_id;
    END LOOP;
END $$;

-- An employee with no global role keeps role_id NULL on every membership, which
-- grants nothing — the same as today, since a NULL role already yields empty
-- permissions.

ALTER TABLE employee DROP CONSTRAINT fk_employee_role;
ALTER TABLE employee DROP COLUMN role_id;

CREATE INDEX idx_employee_companies_role ON employee_companies (role_id);
