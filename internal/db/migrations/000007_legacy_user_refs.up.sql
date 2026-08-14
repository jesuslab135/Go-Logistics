-- BR-04: migrate the last Django auth_user id columns to employee ids.
--
-- employee.user_id preserves the original auth_user id, so it is the mapping
-- key. Rows whose author cannot be mapped become NULL — explicitly
-- unattributed rather than silently credited to the wrong employee — which is
-- why the two NOT NULL columns are relaxed first.
--
-- inventory_journal_entry.user_id already references employee(id) and needs no
-- work here; employee.user_id itself stays as the legacy mapping key.

ALTER TABLE fuel_comment ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE fuel_photo ALTER COLUMN uploaded_by_id DROP NOT NULL;

UPDATE fuel_comment
SET user_id = (SELECT e.id FROM employee e WHERE e.user_id = fuel_comment.user_id);

UPDATE fuel_photo
SET uploaded_by_id = (SELECT e.id FROM employee e WHERE e.user_id = fuel_photo.uploaded_by_id);

UPDATE media
SET uploaded_by_id = (SELECT e.id FROM employee e WHERE e.user_id = media.uploaded_by_id)
WHERE uploaded_by_id IS NOT NULL;

-- The column now names what it actually holds.
ALTER TABLE fuel_comment RENAME COLUMN user_id TO employee_id;

ALTER TABLE fuel_comment ADD CONSTRAINT fk_fuel_comment_employee FOREIGN KEY (employee_id) REFERENCES employee(id) ON DELETE SET NULL;
ALTER TABLE fuel_photo ADD CONSTRAINT fk_fuel_photo_uploaded_by FOREIGN KEY (uploaded_by_id) REFERENCES employee(id) ON DELETE SET NULL;
ALTER TABLE media ADD CONSTRAINT fk_media_uploaded_by FOREIGN KEY (uploaded_by_id) REFERENCES employee(id) ON DELETE SET NULL;
