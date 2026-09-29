-- 000025_scheduled_reports.down.sql
-- Dropping report_run forgets what was sent. If the migration is applied again
-- afterwards, the scheduler owes the latest period of every report to every
-- company and will send it once more.
DROP INDEX IF EXISTS idx_fuel_entry_asset_date;
DROP TABLE IF EXISTS report_run;
DROP TABLE IF EXISTS report_recipient;
