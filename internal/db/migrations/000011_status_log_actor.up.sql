-- 000011_status_log_actor.up.sql
-- Django documented this table as "Historial inmutable de transiciones de estado
-- de una orden de trabajo. Se alimenta automáticamente mediante signals", and
-- signals.py wrote a row on post_save whenever the status differed from the last
-- logged one. The Go port registered it as a writable nested CRUD resource
-- instead, so status changes and log rows became unrelated writes: a work order
-- could change status with no row, and past rows could be edited or deleted.
--
-- The route is now read-only and the row is written inside the transaction that
-- changes the status. These columns add what neither codebase ever recorded:
-- who did it.
ALTER TABLE work_order_status_log
    ADD COLUMN actor_employee_id bigint,
    ADD COLUMN actor_type        varchar(20);

-- Existing rows were written by the CRUD route with no actor captured. Marking
-- them 'system' says "not attributable to a person", which is true; naming a
-- person would be a fabrication.
UPDATE work_order_status_log SET actor_type = 'system' WHERE actor_type IS NULL;

ALTER TABLE work_order_status_log ALTER COLUMN actor_type SET NOT NULL;
ALTER TABLE work_order_status_log ALTER COLUMN actor_type SET DEFAULT 'system';

-- No FK on actor_employee_id: an employee may be deleted, and the record of what
-- they did must outlive them. It is nullable for the same reason, and because a
-- system-initiated transition has no actor at all.

-- The history tab reads one work order's log, newest first.
CREATE INDEX idx_work_order_status_log_history ON work_order_status_log (work_order_id, changed_at DESC);
