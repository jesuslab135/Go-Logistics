-- 000014_money.down.sql
-- Dropping the override columns discards the reasons recorded for figures that
-- deliberately departed from the formula; the totals themselves remain.
ALTER TABLE service_entry  DROP COLUMN IF EXISTS total_override_at, DROP COLUMN IF EXISTS total_override_by_id, DROP COLUMN IF EXISTS total_override_reason, DROP COLUMN IF EXISTS total_override;
ALTER TABLE purchase_order DROP COLUMN IF EXISTS total_override_at, DROP COLUMN IF EXISTS total_override_by_id, DROP COLUMN IF EXISTS total_override_reason, DROP COLUMN IF EXISTS total_override;
ALTER TABLE work_order     DROP COLUMN IF EXISTS total_override_at, DROP COLUMN IF EXISTS total_override_by_id, DROP COLUMN IF EXISTS total_override_reason, DROP COLUMN IF EXISTS total_override;
ALTER TABLE service_entry DROP COLUMN IF EXISTS discount_percentage;
ALTER TABLE work_order    DROP COLUMN IF EXISTS discount_percentage;
