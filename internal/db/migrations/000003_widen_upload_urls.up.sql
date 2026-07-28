-- 000003_widen_upload_urls.up.sql
-- Django's FileField/ImageField stored a RELATIVE key ("media/assets/2026/07/x.png")
-- and django-storages composed the public URL at read time, so max_length=100 was
-- ample. This API stores the fully-qualified URL returned by POST /api/v1/uploads,
-- which routinely exceeds 100 chars once the CDN base and a real filename are in
-- play. Widen to text (same storage as varchar in Postgres, no table rewrite).

ALTER TABLE company                    ALTER COLUMN logo      TYPE text;
ALTER TABLE asset                      ALTER COLUMN photo     TYPE text;
ALTER TABLE fuel_photo                 ALTER COLUMN file      TYPE text;
ALTER TABLE inspection_submission      ALTER COLUMN signature TYPE text;
ALTER TABLE inspection_submission_item ALTER COLUMN photo     TYPE text;
ALTER TABLE media                      ALTER COLUMN file      TYPE text;
