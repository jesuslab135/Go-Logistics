-- 000013_po_workflow.down.sql
-- Rolling back discards the transition history, which cannot be reconstructed
-- from the timestamps alone: those record when a state was reached, not who
-- moved it or what it was rejected for.
DROP TABLE IF EXISTS purchase_order_status_log;
ALTER TABLE purchase_order DROP COLUMN IF EXISTS rejection_reason;
