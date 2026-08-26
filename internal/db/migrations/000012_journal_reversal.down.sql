-- 000012_journal_reversal.down.sql
-- Reversal entries themselves are left in place: they are real movements that
-- really happened, and deleting them would misstate the stock they moved.
DROP INDEX IF EXISTS uq_inventory_journal_entry_reversal;
ALTER TABLE inventory_journal_entry DROP CONSTRAINT IF EXISTS fk_inventory_journal_entry_reversal_of;
ALTER TABLE inventory_journal_entry DROP COLUMN IF EXISTS reversal_of_id;
