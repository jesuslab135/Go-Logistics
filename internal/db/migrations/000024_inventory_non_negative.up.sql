-- 000024_inventory_non_negative.up.sql
-- applyAdjustment locks the stock row and now refuses an adjustment that would
-- take it below zero, but that is one code path's discipline. This is the
-- invariant itself: available_quantity is a physical count, and no write of any
-- kind may leave it negative. ApplyPartInventoryAdjustment is a bare
-- `available_quantity = available_quantity + $1` with no floor, so the column
-- has never had anything defending it.
--
-- This is the schema's first CHECK constraint. It is deliberately narrow: it
-- states a fact about what the column means, not a business rule about how
-- stock may move, which belongs in the handler where it can explain itself.
--
-- Rows already negative would make the constraint un-addable, so they are
-- floored first. That is a visible data correction rather than a silent one: a
-- row that reached a negative count was already reporting a quantity that
-- cannot exist, and its true value was not recorded anywhere this migration
-- could recover it from.
UPDATE part_inventory SET available_quantity = 0 WHERE available_quantity < 0;

ALTER TABLE part_inventory
    ADD CONSTRAINT ck_part_inventory_available_non_negative
    CHECK (available_quantity >= 0);
