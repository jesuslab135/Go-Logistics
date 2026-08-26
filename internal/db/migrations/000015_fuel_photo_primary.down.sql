-- 000015_fuel_photo_primary.down.sql
-- The demoted duplicates are not restored: which of them had been primary was
-- never recorded, because nothing distinguished them.
DROP INDEX IF EXISTS uq_fuel_photo_primary;
