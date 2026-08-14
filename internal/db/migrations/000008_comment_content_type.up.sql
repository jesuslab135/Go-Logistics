-- BR-14: replace Django's content_type_id with a discriminator this API owns.
--
-- content_type_id held a row id from django_content_type, a table assigned at
-- migrate time and never ported here. Clients had to know those ids to file or
-- read a comment, and the values are not reproducible in a fresh database.
--
-- Historic ids are therefore unmappable: existing rows get an empty
-- content_type, an explicit "unknown parent" rather than a guessed one.

ALTER TABLE comment ADD COLUMN content_type varchar(30) NOT NULL DEFAULT '';
ALTER TABLE comment DROP COLUMN content_type_id;

-- object_id points at bigserial primary keys, so integer was always too
-- narrow; Django's PositiveIntegerField is what made it so.
ALTER TABLE comment ALTER COLUMN object_id TYPE bigint;

DROP INDEX IF EXISTS idx_comment_content_type_object;
CREATE INDEX idx_comment_content_type_object ON comment (content_type, object_id);
