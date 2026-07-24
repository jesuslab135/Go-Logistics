-- name: GetWeeklyMileageGoal :one
SELECT * FROM weekly_mileage_goal WHERE id = $1 AND company_id = $2;

-- name: ListWeeklyMileageGoals :many
SELECT * FROM weekly_mileage_goal WHERE company_id = $1 ORDER BY sort_order LIMIT $2 OFFSET $3;

-- name: CountWeeklyMileageGoals :one
SELECT count(*) FROM weekly_mileage_goal WHERE company_id = $1;

-- name: CreateWeeklyMileageGoal :one
INSERT INTO weekly_mileage_goal (
    company_id, service_type, rate_per_mile, weekly_mileage_goal, mpg_goal, units_per_service, machines_in_workshop, missing_miles, sort_order, is_active, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: UpdateWeeklyMileageGoal :one
UPDATE weekly_mileage_goal SET service_type = $3, rate_per_mile = $4, weekly_mileage_goal = $5, mpg_goal = $6, units_per_service = $7, machines_in_workshop = $8, missing_miles = $9, sort_order = $10, is_active = $11, updated_at = $12
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteWeeklyMileageGoal :exec
DELETE FROM weekly_mileage_goal WHERE id = $1 AND company_id = $2;
