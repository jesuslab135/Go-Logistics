-- 000016_trailer_classification.down.sql
-- The varchar columns were never stopped being populated by this migration, so
-- rolling back loses only the catalog itself and the links to it.
ALTER TABLE trailer DROP CONSTRAINT IF EXISTS fk_trailer_classification_2;
ALTER TABLE trailer DROP CONSTRAINT IF EXISTS fk_trailer_classification;
ALTER TABLE trailer DROP COLUMN IF EXISTS classification_2_id, DROP COLUMN IF EXISTS classification_id;
DROP TABLE IF EXISTS trailer_classification;
