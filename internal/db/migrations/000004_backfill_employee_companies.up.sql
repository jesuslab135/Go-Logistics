-- 000004_backfill_employee_companies.up.sql
-- Company access is now scoped by the employee_companies m2m. Employees created
-- before that (including any seeded admin) carry a default_company_id but no
-- membership row, which would lock them out of their own company. Django's
-- CompanyViewSet.perform_create always did both; backfill the rows it implies.

INSERT INTO employee_companies (employee_id, company_id)
SELECT e.id, e.default_company_id
FROM employee e
WHERE e.default_company_id IS NOT NULL
ON CONFLICT (employee_id, company_id) DO NOTHING;
