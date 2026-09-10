-- 000019_accounts.up.sql
-- An account is a client organisation: the thing that owns companies. Until now
-- the model was flat — employee <-> employee_companies <-> company — so nothing
-- recorded which companies belonged to which client, and a client could not be
-- scoped, billed or isolated as a unit.
--
-- owner_employee_id is nullable because an account and its owner reference each
-- other: creation inserts the account, then the owner employee carrying
-- account_id, then sets this. It stays nullable so ownership can be transferred
-- without a window where the column has no legal value.
CREATE TABLE account (
    id                bigserial    PRIMARY KEY,
    name              varchar(200) NOT NULL,
    owner_employee_id bigint,
    is_active         boolean      NOT NULL DEFAULT true,
    created_at        timestamptz  NOT NULL DEFAULT now()
);

ALTER TABLE company            ADD COLUMN account_id bigint;
ALTER TABLE employee           ADD COLUMN account_id bigint;
ALTER TABLE employee_companies ADD COLUMN account_id bigint;

-- Everything that exists today belongs to one client: there has never been a
-- second. Create that account and adopt the existing rows into it. Guarded on
-- there being data at all, so a fresh development database is not given a
-- phantom account.
DO $$
DECLARE acct_id bigint;
BEGIN
    IF EXISTS (SELECT 1 FROM company) OR EXISTS (SELECT 1 FROM employee) THEN
        INSERT INTO account (name, created_at)
        VALUES ('Go Logistics', now())
        RETURNING id INTO acct_id;

        UPDATE company            SET account_id = acct_id WHERE account_id IS NULL;
        UPDATE employee           SET account_id = acct_id WHERE account_id IS NULL;
        UPDATE employee_companies SET account_id = acct_id WHERE account_id IS NULL;

        UPDATE account
           SET owner_employee_id = (
               SELECT id FROM employee WHERE is_account_owner ORDER BY id LIMIT 1
           )
         WHERE id = acct_id;
    END IF;
END $$;

ALTER TABLE company            ALTER COLUMN account_id SET NOT NULL;
ALTER TABLE employee_companies ALTER COLUMN account_id SET NOT NULL;
-- employee.account_id stays nullable: NULL means platform staff, who belong to
-- no client and reach tenant data through /api/v1/admin/*.

ALTER TABLE company  ADD CONSTRAINT fk_company_account
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE RESTRICT;
ALTER TABLE employee ADD CONSTRAINT fk_employee_account
    FOREIGN KEY (account_id) REFERENCES account(id) ON DELETE RESTRICT;
ALTER TABLE account  ADD CONSTRAINT fk_account_owner
    FOREIGN KEY (owner_employee_id) REFERENCES employee(id) ON DELETE SET NULL;

-- The composite foreign keys below need these as their referenced unique keys.
ALTER TABLE employee ADD CONSTRAINT uq_employee_account UNIQUE (id, account_id);
ALTER TABLE company  ADD CONSTRAINT uq_company_account  UNIQUE (id, account_id);

-- The isolation invariant, enforced by the database rather than by middleware:
-- a membership may only link an employee and a company in the SAME account.
-- A cross-account membership becomes impossible to insert — from the API, from
-- the CLI, from a hand-written INSERT during an incident. A middleware check
-- cannot give that guarantee, because every future endpoint that writes a
-- membership would have to remember it.
--
-- The single-column FKs are dropped: the composite ones subsume them, and
-- keeping both would mean two cascade paths for the same relationship.
ALTER TABLE employee_companies DROP CONSTRAINT fk_employee_companies_employee;
ALTER TABLE employee_companies DROP CONSTRAINT fk_employee_companies_company;

ALTER TABLE employee_companies ADD CONSTRAINT fk_ec_employee_account
    FOREIGN KEY (employee_id, account_id) REFERENCES employee(id, account_id) ON DELETE CASCADE;
ALTER TABLE employee_companies ADD CONSTRAINT fk_ec_company_account
    FOREIGN KEY (company_id, account_id) REFERENCES company(id, account_id) ON DELETE CASCADE;

CREATE INDEX idx_company_account  ON company (account_id);
CREATE INDEX idx_employee_account ON employee (account_id);
