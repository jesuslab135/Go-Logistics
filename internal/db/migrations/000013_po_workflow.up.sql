-- 000013_po_workflow.up.sql
-- purchase_order carried a full state machine's worth of columns - state, plus
-- submitted/rejected/approved/purchased/received/closed timestamps and three
-- actor ids - and the only way to write any of them was PUT, a whole-record
-- replace. A client could set state to CLOSED without ever passing through
-- approval, stamp approved_by_id with somebody else's id, or move backwards.
-- Nothing recorded the transitions either way.
--
-- The dedicated transition routes added alongside this migration stamp the
-- state, its timestamp and the actor together, from the authenticated caller.

-- Rejection is not terminal: REJECTED returns to DRAFT so the order can be
-- revised and resubmitted, which is how procurement rejection actually works.
-- A reason is required at rejection, because "no" without one sends the buyer
-- back to guess.
ALTER TABLE purchase_order ADD COLUMN rejection_reason text NOT NULL DEFAULT '';

-- Same shape and the same reasoning as work_order_status_log: append-only,
-- written inside the transaction that moves the state, with a nullable actor
-- and no FK so the record outlives the employee.
--
-- from_state is recorded as well as to_state. A work order's history can be read
-- as a sequence because every change is logged; a purchase order's cannot be
-- assumed to be, since rows written before this migration do not exist at all.
CREATE TABLE purchase_order_status_log (
    id                bigserial   PRIMARY KEY,
    purchase_order_id bigint      NOT NULL,
    from_state        varchar(20) NOT NULL,
    to_state          varchar(20) NOT NULL,
    actor_employee_id bigint,
    actor_type        varchar(20) NOT NULL DEFAULT 'employee',
    reason            text        NOT NULL DEFAULT '',
    changed_at        timestamptz NOT NULL
);

ALTER TABLE purchase_order_status_log
    ADD CONSTRAINT fk_purchase_order_status_log_order
    FOREIGN KEY (purchase_order_id) REFERENCES purchase_order(id) ON DELETE CASCADE;

CREATE INDEX idx_purchase_order_status_log_history ON purchase_order_status_log (purchase_order_id, changed_at DESC);
