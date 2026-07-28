-- 000003_widen_upload_urls.down.sql
-- Narrowing back rewrites the table and fails if any stored value exceeds 100
-- chars, which is the situation this migration exists to fix.

ALTER TABLE company                    ALTER COLUMN logo      TYPE varchar(100);
ALTER TABLE asset                      ALTER COLUMN photo     TYPE varchar(100);
ALTER TABLE fuel_photo                 ALTER COLUMN file      TYPE varchar(100);
ALTER TABLE inspection_submission      ALTER COLUMN signature TYPE varchar(100);
ALTER TABLE inspection_submission_item ALTER COLUMN photo     TYPE varchar(100);
ALTER TABLE media                      ALTER COLUMN file      TYPE varchar(100);
