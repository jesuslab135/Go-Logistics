-- 000010_platform_admin.up.sql
-- A platform administrator is internal staff who may act across tenants. Until
-- now there was no such concept: /admin/* was gated on employee.is_admin, which
-- is a *tenant's own* administrator, so any company admin could read and write
-- cross-tenant employee data. employee.role_id is a single global FK, so an
-- admin of one company is an admin wherever they hold membership, which is what
-- made that gate reach further than it looks.
--
-- A boolean column rather than a table: the flag is per person and has no
-- attributes. A table would buy granularity nothing asks for yet, and can
-- replace this later without changing how the gate reads.
--
-- Nobody is a platform admin after this migration. The flag is granted from the
-- CLI (cmd/cli platform-admin), deliberately: the population is a handful of
-- internal staff, and a self-service surface would be an attack surface with no
-- user. Grant the first one before deploying the gate, or /admin/* is closed to
-- everyone.
ALTER TABLE employee ADD COLUMN is_platform_admin boolean NOT NULL DEFAULT false;

-- Membership changes are the highest-privilege write in this system: granting a
-- company confers admin there, because role_id is global. They were previously
-- unrecorded — a full replace overwrote employee_companies with no trace of who
-- changed what. This is append-only and never updated.
--
-- actor_employee_id is nullable and carries no FK to employee: an actor may be
-- deleted, and the record of what they did must outlive them.
CREATE TABLE membership_audit (
    id                  bigserial   PRIMARY KEY,
    actor_employee_id   bigint,
    subject_employee_id bigint      NOT NULL,
    company_id          bigint      NOT NULL,
    action              varchar(10) NOT NULL,
    occurred_at         timestamptz NOT NULL DEFAULT now()
);

-- The audit is read per subject ("what happened to this employee's access")
-- and per company ("who was added to this tenant"), both newest first.
CREATE INDEX idx_membership_audit_subject ON membership_audit (subject_employee_id, occurred_at DESC);
CREATE INDEX idx_membership_audit_company ON membership_audit (company_id, occurred_at DESC);
