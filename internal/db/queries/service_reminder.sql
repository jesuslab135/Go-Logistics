-- name: GetServiceReminder :one
SELECT * FROM service_reminder WHERE id = $1 AND company_id = $2;

-- name: ListServiceReminders :many
SELECT * FROM service_reminder WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: CountServiceReminders :one
SELECT count(*) FROM service_reminder WHERE company_id = $1;

-- name: CreateServiceReminder :one
INSERT INTO service_reminder (
    company_id, asset_id, service_task_id, is_active, status, time_interval, time_frequency, next_due_at, due_soon_at, due_soon_time_threshold, meter_interval, next_due_meter_value, due_soon_meter_value, due_soon_meter_threshold, snooze_until, last_service_entry_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
)
RETURNING *;

-- name: UpdateServiceReminder :one
UPDATE service_reminder SET asset_id = $3, service_task_id = $4, is_active = $5, status = $6, time_interval = $7, time_frequency = $8, next_due_at = $9, due_soon_at = $10, due_soon_time_threshold = $11, meter_interval = $12, next_due_meter_value = $13, due_soon_meter_value = $14, due_soon_meter_threshold = $15, snooze_until = $16, last_service_entry_id = $17, updated_at = $18
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteServiceReminder :exec
DELETE FROM service_reminder WHERE id = $1 AND company_id = $2;
