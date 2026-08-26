-- name: GetTrailer :one
SELECT c.* FROM trailer c JOIN asset a ON a.id = c.asset_id
WHERE c.asset_id = sqlc.arg(parent_id) AND a.company_id = sqlc.arg(company_id);

-- GetTrailerClassificationNames resolves both catalog names for one trailer, so
-- a client renders words rather than an id it would have to look up itself.
--
-- A second statement rather than a join on GetTrailer: sqlc flattens a joined
-- row into a new struct, which would mean a second copy of the trailer field
-- mapping - and two mappings of forty columns is how one of them goes stale.
-- name: GetTrailerClassificationNames :one
SELECT
    COALESCE(tc1.name, '')::varchar AS classification_name,
    COALESCE(tc2.name, '')::varchar AS classification_2_name
FROM trailer c
LEFT JOIN trailer_classification tc1 ON tc1.id = c.classification_id
LEFT JOIN trailer_classification tc2 ON tc2.id = c.classification_2_id
WHERE c.asset_id = sqlc.arg(parent_id);

-- name: UpsertTrailer :one
INSERT INTO trailer (
    asset_id, trailer_type, classification_id, classification_2_id, size, suspension, owner_name, financing, supplier, license_plate_us, license_plate_us_state, license_plate_mx, doors, walls, rail_post, skylight, floor_type, roof_type, hazmat_type, aero_kit_type, gps_provider, gps_serial, gps_signal_status, gps_contract_start, gps_contract_end, gps_contract_reference, contract_start, contract_end, contract_period, contract_reference, fumigation_date, fumigation_cert, registration_date, waterproofing_date, decommission_reason, decommission_date, operational_use, operation_zone
)
SELECT sqlc.arg(parent_id), sqlc.arg(trailer_type), sqlc.narg(classification_id), sqlc.narg(classification_2_id), sqlc.arg(size), sqlc.arg(suspension), sqlc.arg(owner_name), sqlc.arg(financing), sqlc.arg(supplier), sqlc.arg(license_plate_us), sqlc.arg(license_plate_us_state), sqlc.arg(license_plate_mx), sqlc.arg(doors), sqlc.arg(walls), sqlc.arg(rail_post), sqlc.arg(skylight), sqlc.arg(floor_type), sqlc.arg(roof_type), sqlc.arg(hazmat_type), sqlc.arg(aero_kit_type), sqlc.arg(gps_provider), sqlc.arg(gps_serial), sqlc.arg(gps_signal_status), sqlc.arg(gps_contract_start), sqlc.arg(gps_contract_end), sqlc.arg(gps_contract_reference), sqlc.arg(contract_start), sqlc.arg(contract_end), sqlc.arg(contract_period), sqlc.arg(contract_reference), sqlc.arg(fumigation_date), sqlc.arg(fumigation_cert), sqlc.arg(registration_date), sqlc.arg(waterproofing_date), sqlc.arg(decommission_reason), sqlc.arg(decommission_date), sqlc.arg(operational_use), sqlc.arg(operation_zone)
WHERE EXISTS (SELECT 1 FROM asset WHERE id = sqlc.arg(parent_id) AND company_id = sqlc.arg(company_id))
ON CONFLICT (asset_id) DO UPDATE SET trailer_type = EXCLUDED.trailer_type, classification_id = EXCLUDED.classification_id, classification_2_id = EXCLUDED.classification_2_id, size = EXCLUDED.size, suspension = EXCLUDED.suspension, owner_name = EXCLUDED.owner_name, financing = EXCLUDED.financing, supplier = EXCLUDED.supplier, license_plate_us = EXCLUDED.license_plate_us, license_plate_us_state = EXCLUDED.license_plate_us_state, license_plate_mx = EXCLUDED.license_plate_mx, doors = EXCLUDED.doors, walls = EXCLUDED.walls, rail_post = EXCLUDED.rail_post, skylight = EXCLUDED.skylight, floor_type = EXCLUDED.floor_type, roof_type = EXCLUDED.roof_type, hazmat_type = EXCLUDED.hazmat_type, aero_kit_type = EXCLUDED.aero_kit_type, gps_provider = EXCLUDED.gps_provider, gps_serial = EXCLUDED.gps_serial, gps_signal_status = EXCLUDED.gps_signal_status, gps_contract_start = EXCLUDED.gps_contract_start, gps_contract_end = EXCLUDED.gps_contract_end, gps_contract_reference = EXCLUDED.gps_contract_reference, contract_start = EXCLUDED.contract_start, contract_end = EXCLUDED.contract_end, contract_period = EXCLUDED.contract_period, contract_reference = EXCLUDED.contract_reference, fumigation_date = EXCLUDED.fumigation_date, fumigation_cert = EXCLUDED.fumigation_cert, registration_date = EXCLUDED.registration_date, waterproofing_date = EXCLUDED.waterproofing_date, decommission_reason = EXCLUDED.decommission_reason, decommission_date = EXCLUDED.decommission_date, operational_use = EXCLUDED.operational_use, operation_zone = EXCLUDED.operation_zone
RETURNING *;

-- name: DeleteTrailer :exec
DELETE FROM trailer c USING asset a
WHERE c.asset_id = sqlc.arg(parent_id) AND a.id = c.asset_id AND a.company_id = sqlc.arg(company_id);
