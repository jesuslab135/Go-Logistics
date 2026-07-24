-- name: GetTire :one
SELECT * FROM tire WHERE id = $1 AND company_id = $2;

-- name: ListTires :many
SELECT * FROM tire WHERE company_id = $1 ORDER BY tire_identification_number LIMIT $2 OFFSET $3;

-- name: CountTires :one
SELECT count(*) FROM tire WHERE company_id = $1;

-- name: CreateTire :one
INSERT INTO tire (
    company_id, tire_identification_number, tire_model_id, status, current_tread_depth_32nds, current_psi, total_miles, current_vehicle_id, current_position_code, purchase_date, purchase_cost, vendor_id, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: UpdateTire :one
UPDATE tire SET tire_identification_number = $3, tire_model_id = $4, status = $5, current_tread_depth_32nds = $6, current_psi = $7, total_miles = $8, current_vehicle_id = $9, current_position_code = $10, purchase_date = $11, purchase_cost = $12, vendor_id = $13
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteTire :exec
DELETE FROM tire WHERE id = $1 AND company_id = $2;
