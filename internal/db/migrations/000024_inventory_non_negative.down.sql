-- 000024_inventory_non_negative.down.sql
-- The floored rows are not restored: their pre-migration values were not
-- recorded, and re-deriving them from inventory_journal_entry is not work a
-- down migration should attempt.
ALTER TABLE part_inventory
    DROP CONSTRAINT IF EXISTS ck_part_inventory_available_non_negative;
