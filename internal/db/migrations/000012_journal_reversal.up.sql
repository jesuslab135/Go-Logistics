-- 000012_journal_reversal.up.sql
-- The journal is the record of stock movements, and from this migration on it is
-- also what performs them: creating an entry moves
-- part_inventory.available_quantity in the same transaction.
--
-- That is what previous_quantity / adjustment_quantity / current_quantity always
-- meant. Django's Excel importer wrote them exactly that way — read the current
-- quantity, add the adjustment, save the row, then record the three values
-- around the move — but its ordinary endpoint, and this port's, let a client
-- supply all three and moved nothing. The columns recorded a movement the
-- inventory never made.
--
-- Correcting an entry is a new, offsetting entry rather than an edit, so the
-- ledger stays append-only and the correction is itself visible.
ALTER TABLE inventory_journal_entry ADD COLUMN reversal_of_id bigint;

ALTER TABLE inventory_journal_entry
    ADD CONSTRAINT fk_inventory_journal_entry_reversal_of
    FOREIGN KEY (reversal_of_id) REFERENCES inventory_journal_entry(id) ON DELETE RESTRICT;

-- One reversal per entry. The uniqueness is the rule: without it a client could
-- reverse the same entry repeatedly and drive the stock arbitrarily far from
-- what the ledger sums to.
CREATE UNIQUE INDEX uq_inventory_journal_entry_reversal ON inventory_journal_entry (reversal_of_id)
    WHERE reversal_of_id IS NOT NULL;

-- Existing rows are NOT reconciled against part_inventory. They were written
-- while nothing moved stock, so the two have drifted by an unknown amount, and
-- rewriting history to match a ledger that was never authoritative would destroy
-- the only evidence of what actually happened. Run
-- `fleet-cli inventory-drift` to see the divergence per part.
