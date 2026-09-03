-- 000018_trailer_classification_default.up.sql
-- trailer.classification / classification_2 are deprecated (superseded by
-- classification_id in 000016) and no longer written by UpsertTrailer, but they
-- were left NOT NULL with no default - so every new trailer INSERT fails 23502.
-- Give them an empty-string default so writes that omit them succeed. The
-- columns stay NOT NULL and varchar, so sqlc's generated model is unchanged.
ALTER TABLE trailer ALTER COLUMN classification   SET DEFAULT '';
ALTER TABLE trailer ALTER COLUMN classification_2 SET DEFAULT '';
