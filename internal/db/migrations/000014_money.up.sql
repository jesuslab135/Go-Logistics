-- 000014_money.up.sql
-- Totals become server-authoritative. Until now every money column on a work
-- order, purchase order or service entry was whatever the client sent: the
-- handlers assigned in.TotalAmount straight through and nothing computed
-- anything. Three clients each doing their own arithmetic is three answers to
-- the same question, and the first disagreement lands on an invoice.

-- discount_percentage existed on purchase_order but not on work_order or
-- service_entry, even though both carry discount_type with a PERCENTAGE option.
-- A percentage discount was literally unrepresentable there - the type could say
-- PERCENTAGE while the only available number was a fixed amount.
ALTER TABLE work_order    ADD COLUMN discount_percentage numeric(5,2) NOT NULL DEFAULT 0;
ALTER TABLE service_entry ADD COLUMN discount_percentage numeric(5,2) NOT NULL DEFAULT 0;

-- An override is the escape hatch for the case the formula cannot express: a
-- negotiated figure, a rounding agreed with a vendor, a correction to a
-- historical document. It is audited because it overrides the only number the
-- system can justify, and null when the computed total stands.
--
-- It does not survive an edit to the line items. A stale override silently
-- misstating a total is exactly the failure this work exists to prevent, so a
-- line-item write clears it and the total returns to computed - visibly, rather
-- than drifting.
ALTER TABLE work_order
    ADD COLUMN total_override        numeric(12,2),
    ADD COLUMN total_override_reason text,
    ADD COLUMN total_override_by_id  bigint,
    ADD COLUMN total_override_at     timestamptz;

ALTER TABLE purchase_order
    ADD COLUMN total_override        numeric(12,2),
    ADD COLUMN total_override_reason text,
    ADD COLUMN total_override_by_id  bigint,
    ADD COLUMN total_override_at     timestamptz;

ALTER TABLE service_entry
    ADD COLUMN total_override        numeric(12,2),
    ADD COLUMN total_override_reason text,
    ADD COLUMN total_override_by_id  bigint,
    ADD COLUMN total_override_at     timestamptz;

-- No FK on total_override_by_id, for the reason the audit columns elsewhere
-- have none: the record of who overrode a total must outlive the employee.

-- Existing rows are left exactly as they are. Their totals were client-supplied
-- and may not satisfy the formula, but they are what the documents said at the
-- time, and recomputing historical invoices to match a formula written after
-- them would restate figures somebody has already acted on. They are recomputed
-- the next time the document is edited, which is a deliberate act.
