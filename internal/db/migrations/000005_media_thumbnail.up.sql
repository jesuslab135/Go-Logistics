-- 000005_media_thumbnail.up.sql
-- The gallery grid needs a small image per item. The thumbnail is generated at
-- upload time and stored under a key derived from the original, but the column
-- records the URL that was actually written rather than leaving clients to
-- assume a thumbnail exists: media uploaded before this migration, and formats
-- the server cannot decode (WebP, PDF), legitimately have none.
ALTER TABLE media ADD COLUMN thumbnail text NOT NULL DEFAULT '';
