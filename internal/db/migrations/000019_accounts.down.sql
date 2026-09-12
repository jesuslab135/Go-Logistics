-- 000019_accounts.down.sql
ALTER TABLE employee_companies DROP CONSTRAINT fk_ec_company_account;
ALTER TABLE employee_companies DROP CONSTRAINT fk_ec_employee_account;

ALTER TABLE employee_companies ADD CONSTRAINT fk_employee_companies_employee
    FOREIGN KEY (employee_id) REFERENCES employee(id) ON DELETE CASCADE;
ALTER TABLE employee_companies ADD CONSTRAINT fk_employee_companies_company
    FOREIGN KEY (company_id) REFERENCES company(id) ON DELETE CASCADE;

ALTER TABLE company  DROP CONSTRAINT uq_company_account;
ALTER TABLE employee DROP CONSTRAINT uq_employee_account;

ALTER TABLE account  DROP CONSTRAINT fk_account_owner;
ALTER TABLE employee DROP CONSTRAINT fk_employee_account;
ALTER TABLE company  DROP CONSTRAINT fk_company_account;

DROP INDEX IF EXISTS idx_employee_account;
DROP INDEX IF EXISTS idx_company_account;

ALTER TABLE employee_companies DROP COLUMN account_id;
ALTER TABLE employee           DROP COLUMN account_id;
ALTER TABLE company            DROP COLUMN account_id;

DROP TABLE account;
