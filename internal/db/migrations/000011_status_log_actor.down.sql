-- 000011_status_log_actor.down.sql
-- Rolling back discards every recorded actor; the information cannot be
-- reconstructed from anywhere else.
DROP INDEX IF EXISTS idx_work_order_status_log_history;
ALTER TABLE work_order_status_log
    DROP COLUMN IF EXISTS actor_type,
    DROP COLUMN IF EXISTS actor_employee_id;
