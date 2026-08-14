-- Lossy: the Django content-type ids were never recoverable, so rows come back
-- pointing at content type 0.

DROP INDEX IF EXISTS idx_comment_content_type_object;

ALTER TABLE comment ALTER COLUMN object_id TYPE integer;

ALTER TABLE comment ADD COLUMN content_type_id bigint NOT NULL DEFAULT 0;
ALTER TABLE comment DROP COLUMN content_type;

CREATE INDEX idx_comment_content_type_object ON comment (content_type_id, object_id);
