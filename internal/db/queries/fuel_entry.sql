-- name: ListFuelEntries :many
SELECT c.* FROM fuel_entry c JOIN asset p ON p.id = c.asset_id
WHERE c.asset_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id)
ORDER BY c.date DESC, c.id LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountFuelEntries :one
SELECT count(*) FROM fuel_entry c JOIN asset p ON p.id = c.asset_id WHERE c.asset_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- name: GetFuelEntry :one
SELECT c.* FROM fuel_entry c JOIN asset p ON p.id = c.asset_id
WHERE c.id = sqlc.arg(id) AND c.asset_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id);

-- miles_traveled and fuel_efficiency are omitted on purpose: they are derived
-- by recalcFuelSeries after the write, never taken from the request.
-- name: CreateFuelEntry :one
INSERT INTO fuel_entry (
    asset_id, employee_id, date, fuel_type, quantity, unit_cost, total_cost, odometer, vendor_id, full_tank, state, reference, personal, reset, latitude, longitude, external_id, no_semana, estado_prov, operator_name, updated_at
)
SELECT sqlc.arg(parent_id), sqlc.arg(employee_id), sqlc.arg(date), sqlc.arg(fuel_type), sqlc.arg(quantity), sqlc.arg(unit_cost), sqlc.arg(total_cost), sqlc.arg(odometer), sqlc.arg(vendor_id), sqlc.arg(full_tank), sqlc.arg(state), sqlc.arg(reference), sqlc.arg(personal), sqlc.arg(reset), sqlc.arg(latitude), sqlc.arg(longitude), sqlc.arg(external_id), sqlc.arg(no_semana), sqlc.arg(estado_prov), sqlc.arg(operator_name), sqlc.arg(updated_at)
WHERE EXISTS (SELECT 1 FROM asset WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
RETURNING *;

-- name: UpdateFuelEntry :one
UPDATE fuel_entry AS c SET employee_id = sqlc.arg(employee_id), date = sqlc.arg(date), fuel_type = sqlc.arg(fuel_type), quantity = sqlc.arg(quantity), unit_cost = sqlc.arg(unit_cost), total_cost = sqlc.arg(total_cost), odometer = sqlc.arg(odometer), vendor_id = sqlc.arg(vendor_id), full_tank = sqlc.arg(full_tank), state = sqlc.arg(state), reference = sqlc.arg(reference), personal = sqlc.arg(personal), reset = sqlc.arg(reset), latitude = sqlc.arg(latitude), longitude = sqlc.arg(longitude), external_id = sqlc.arg(external_id), updated_at = sqlc.arg(updated_at), no_semana = sqlc.arg(no_semana), estado_prov = sqlc.arg(estado_prov), operator_name = sqlc.arg(operator_name)
FROM asset p
WHERE c.id = sqlc.arg(id) AND c.asset_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.asset_id
RETURNING c.*;

-- name: DeleteFuelEntry :exec
DELETE FROM fuel_entry AS c USING asset p
WHERE c.id = sqlc.arg(id) AND c.asset_id = sqlc.arg(parent_id) AND p.company_id = sqlc.arg(company_id) AND p.id = c.asset_id;

-- Fuel derivation works along a series: entries for one asset and one fuel
-- type, ordered by (date, id). The row tuple comparison makes that ordering
-- the identity of a position in the series, so an entry inserted between two
-- others is handled the same way as one appended at the end.

-- name: GetPreviousFuelEntry :one
SELECT id, date, odometer, quantity, full_tank, reset
FROM fuel_entry
WHERE asset_id = sqlc.arg(asset_id) AND fuel_type = sqlc.arg(fuel_type)
  AND (date, id) < (sqlc.arg(date)::timestamptz, sqlc.arg(id)::bigint)
ORDER BY date DESC, id DESC
LIMIT 1;

-- name: ListFuelEntriesFrom :many
SELECT id, date, odometer, quantity, full_tank, reset
FROM fuel_entry
WHERE asset_id = sqlc.arg(asset_id) AND fuel_type = sqlc.arg(fuel_type)
  AND (date, id) >= (sqlc.arg(date)::timestamptz, sqlc.arg(id)::bigint)
ORDER BY date, id
FOR UPDATE;

-- name: SetFuelEntryDerived :exec
UPDATE fuel_entry SET miles_traveled = sqlc.arg(miles_traveled), fuel_efficiency = sqlc.arg(fuel_efficiency)
WHERE id = sqlc.arg(id);

-- name: GetFuelEntrySeries :one
-- The series an entry belongs to, for recomputing after an edit or delete.
SELECT c.id, c.asset_id, c.fuel_type, c.date
FROM fuel_entry c JOIN asset p ON p.id = c.asset_id
WHERE c.id = sqlc.arg(id) AND p.company_id = sqlc.arg(company_id);
