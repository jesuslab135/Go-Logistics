-- Lossy by nature: the ids stay employee ids. This only restores the old
-- column names, nullability and the absence of foreign keys.

ALTER TABLE media DROP CONSTRAINT IF EXISTS fk_media_uploaded_by;
ALTER TABLE fuel_photo DROP CONSTRAINT IF EXISTS fk_fuel_photo_uploaded_by;
ALTER TABLE fuel_comment DROP CONSTRAINT IF EXISTS fk_fuel_comment_employee;

ALTER TABLE fuel_comment RENAME COLUMN employee_id TO user_id;

UPDATE fuel_comment SET user_id = 0 WHERE user_id IS NULL;
UPDATE fuel_photo SET uploaded_by_id = 0 WHERE uploaded_by_id IS NULL;

ALTER TABLE fuel_comment ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE fuel_photo ALTER COLUMN uploaded_by_id SET NOT NULL;
