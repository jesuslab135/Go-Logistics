-- name: GetAsset :one
SELECT * FROM asset WHERE id = $1 AND company_id = $2;

-- name: ListAssets :many
SELECT * FROM asset WHERE company_id = $1 ORDER BY name LIMIT $2 OFFSET $3;

-- name: CountAssets :one
SELECT count(*) FROM asset WHERE company_id = $1;

-- name: CreateAsset :one
INSERT INTO asset (
    company_id, name, vin_sn, msrp, generate_expenses, asset_type_id, status_id, lease_vendor_id, vehicle_type, ownership_type, labels, linked_vehicles, loan_start_date, loan_end_date, monthly_payment, number_of_payments, lease_number, lease_start_date, lease_end_date, excess_mileage_charge, owner_company_id, year, make, model, trim, color, license_plate, "group", photo, meter_unit, current_meter, secondary_meter_unit, secondary_meter_value, fuel_type, body_type, body_subtype, registration_state, purchase_date, purchase_price, purchase_vendor, purchase_meter, in_service_date, in_service_meter, out_of_service_date, out_of_service_meter, estimated_service_months, estimated_replacement_mileage, estimated_resale_price, acquisition_type, monthly_cost, acquisition_date, loan_amount, capitalized_cost, down_payment, annual_percentage_rate, first_payment_date, residual_value, mileage_cap, notes, archived_at, external_id, custom_fields, fuel_volume_units, current_meter_date, loan_account_number, loan_notes, loan_vendor_id, loan_started_at, loan_ended_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58, $59, $60, $61, $62, $63, $64, $65, $66, $67, $68, $69, $70
)
RETURNING *;

-- name: UpdateAsset :one
UPDATE asset SET name = $3, vin_sn = $4, msrp = $5, generate_expenses = $6, asset_type_id = $7, status_id = $8, lease_vendor_id = $9, vehicle_type = $10, ownership_type = $11, labels = $12, linked_vehicles = $13, loan_start_date = $14, loan_end_date = $15, monthly_payment = $16, number_of_payments = $17, lease_number = $18, lease_start_date = $19, lease_end_date = $20, excess_mileage_charge = $21, owner_company_id = $22, year = $23, make = $24, model = $25, trim = $26, color = $27, license_plate = $28, "group" = $29, photo = $30, meter_unit = $31, current_meter = $32, secondary_meter_unit = $33, secondary_meter_value = $34, fuel_type = $35, body_type = $36, body_subtype = $37, registration_state = $38, purchase_date = $39, purchase_price = $40, purchase_vendor = $41, purchase_meter = $42, in_service_date = $43, in_service_meter = $44, out_of_service_date = $45, out_of_service_meter = $46, estimated_service_months = $47, estimated_replacement_mileage = $48, estimated_resale_price = $49, acquisition_type = $50, monthly_cost = $51, acquisition_date = $52, loan_amount = $53, capitalized_cost = $54, down_payment = $55, annual_percentage_rate = $56, first_payment_date = $57, residual_value = $58, mileage_cap = $59, notes = $60, archived_at = $61, external_id = $62, custom_fields = $63, fuel_volume_units = $64, current_meter_date = $65, loan_account_number = $66, loan_notes = $67, loan_vendor_id = $68, loan_started_at = $69, loan_ended_at = $70, updated_at = $71
WHERE id = $1 AND company_id = $2
RETURNING *;

-- name: DeleteAsset :exec
DELETE FROM asset WHERE id = $1 AND company_id = $2;
