// qa/route-audit/lib/fixtures.mjs

// Resources that answer POST .../{id}/archive when a delete is refused.
const ARCHIVABLE = new Set(['vendor', 'part', 'asset', 'serviceTask', 'inspectionForm'])

// Fixtures whose Create DTO has no free-text field, so there is nowhere to put
// the run tag. Teardown still tracks them by id in `created`, and all three
// are grandchildren that die with their parent, so cleanup is unaffected —
// the tag is a human-recovery aid, not the deletion mechanism.
//
// wheelPosition (dto/wheel_position_definition.go) has only code (varchar(10),
// NOT NULL), side (varchar(1), NOT NULL) and slot (int) — none of them free
// text, and code is far too short to hold `ZZ-TEST-<runId>` for most runIds
// without truncating (a truncated tag is not a usable tag: it defeats the
// human recovery sweep exactly as no tag would). code/side instead carry
// plausible domain values ('L1', 'L') so the row is realistic.
export const UNTAGGABLE = new Set(['laborEntry', 'purchaseOrderLineItem', 'wheelPosition'])

// Rows on an append-only ledger cannot be deleted; reversing is the system's
// own answer, exactly as archiving is for a referenced catalog row. The
// reversal is itself a new row, so the ledger ends with a balanced pair
// rather than one unbalanced adjustment.
//
// Why partInventory (neither ARCHIVABLE nor REVERSIBLE) needs no fallback
// of its own: journalEntry is reversed, not deleted, so both the original
// entry and its reversal row are still present — by design — when teardown
// reaches partInventory later in the reverse walk. partInventory's DELETE
// still succeeds anyway, because inventory_journal_entry.part_location_detail_id
// is declared ON DELETE CASCADE against part_inventory(id)
// (internal/db/migrations/000001_init.up.sql:1281) — deleting the
// part_inventory row takes both ledger rows with it. The reversal is not
// made redundant by that cascade: it neutralises the stock movement before
// the cascade fires, so the ledger reads as balanced at every instant, not
// only after teardown finishes removing its parent.
export const REVERSIBLE = new Set(['journalEntry'])

export const FIXTURE_PLAN = [
  // Level 0 — depend on nothing but the company.
  { key: 'vendor', path: '/api/v1/vendors', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-vendor`, external_id: `${tag}-v` }) },
  { key: 'location', path: '/api/v1/locations', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-location` }) },
  { key: 'assetType', path: '/api/v1/asset-types', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-asset-type` }) },
  { key: 'assetStatus', path: '/api/v1/asset-statuses', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-asset-status` }) },
  { key: 'partCategory', path: '/api/v1/part-categories', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-part-category` }) },
  { key: 'partManufacturer', path: '/api/v1/part-manufacturers', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-part-manufacturer` }) },
  { key: 'measurementUnit', path: '/api/v1/measurement-units', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-unit`, abbreviation: 'ZZ' }) },
  { key: 'partLocation', path: '/api/v1/part-locations', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-part-location` }) },
  { key: 'adjustmentReason', path: '/api/v1/inventory-adjustment-reasons', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-adjustment-reason` }) },
  { key: 'fuelType', path: '/api/v1/fuel-types', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-fuel-type` }) },
  { key: 'trailerClassification', path: '/api/v1/trailer-classifications', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-trailer-class` }) },
  { key: 'issuePriority', path: '/api/v1/issue-priorities', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-priority` }) },
  { key: 'workOrderStatus', path: '/api/v1/work-order-statuses', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-wo-status` }) },
  { key: 'vehicleMake', path: '/api/v1/vehicle-makes', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-make` }) },
  { key: 'axleTemplate', path: '/api/v1/axle-templates', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-axle-template` }) },
  { key: 'serviceTask', path: '/api/v1/service-tasks', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-service-task` }) },
  { key: 'role', path: '/api/v1/roles', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-role`, is_admin: false }) },
  { key: 'group', path: '/api/v1/groups', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-group` }) },
  { key: 'fault', path: '/api/v1/faults', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-fault` }) },
  { key: 'inspectionForm', path: '/api/v1/inspection-forms', dependsOn: [],
    body: (_, tag) => ({ title: `${tag}-inspection-form`, description: `${tag} audit form`, version: 1 }) },
  { key: 'tireModel', path: '/api/v1/tire-models', dependsOn: [],
    body: (_, tag) => ({ brand: `${tag}-brand`, model_name: `${tag}-tire-model`, size: '295/75R22.5' }) },

  // Level 1 — one hop.
  { key: 'vehicleModel', path: '/api/v1/vehicle-models', dependsOn: ['vehicleMake'],
    body: (ids, tag) => ({ name: `${tag}-model`, make_id: ids.vehicleMake }) },
  { key: 'part', path: '/api/v1/parts', dependsOn: ['partCategory', 'partManufacturer', 'measurementUnit'],
    body: (ids, tag) => ({
      part_number: `${tag}-PN`, description: `${tag}-part`,
      part_category_id: ids.partCategory, part_manufacturer_id: ids.partManufacturer,
      measurement_unit_id: ids.measurementUnit,
    }) },
  { key: 'asset', path: '/api/v1/assets', dependsOn: ['assetType', 'assetStatus'],
    body: (ids, tag) => ({
      name: `${tag}-asset`, vin_sn: `${tag}-VIN`,
      asset_type_id: ids.assetType, status_id: ids.assetStatus,
      vehicle_type: 'VEHICLE', ownership_type: 'OWNED',
    }) },
  { key: 'trailerAsset', path: '/api/v1/assets', dependsOn: ['assetType', 'assetStatus'],
    body: (ids, tag) => ({
      name: `${tag}-trailer-asset`, vin_sn: `${tag}-TVIN`,
      asset_type_id: ids.assetType, status_id: ids.assetStatus,
      vehicle_type: 'TRAILER', ownership_type: 'OWNED',
    }) },
  { key: 'tire', path: '/api/v1/tires', dependsOn: ['tireModel'],
    body: (ids, tag) => ({ tire_identification_number: `${tag}-tire`, tire_model_id: ids.tireModel, status: 'IN_STOCK' }) },
  { key: 'axleDefinition', path: '/api/v1/axle-templates/{axleTemplate}/definitions', dependsOn: ['axleTemplate'],
    body: (_, tag) => ({ label: `${tag}-axle-def`, position_index: 1, axle_role: 'DRIVE', positions_per_side: 2 }) },
  { key: 'employee', path: '/api/v1/employees', dependsOn: ['role', 'group'],
    body: (ids, tag) => ({
      first_name: 'ZZTEST', last_name: `${tag}-employee`,
      email: `${tag.toLowerCase()}-employee@example.invalid`,
      role_id: ids.role, group_id: ids.group, is_active: true,
    }) },

  // Level 2 — two hops.
  { key: 'workOrder', path: '/api/v1/work-orders', dependsOn: ['asset', 'workOrderStatus'],
    body: (ids, tag) => ({
      number: `${tag}-WO`, description: `${tag} work order`,
      asset_id: ids.asset, status_id: ids.workOrderStatus,
      issued_at: '2026-09-09T00:00:00Z',
    }) },
  { key: 'issue', path: '/api/v1/issues', dependsOn: ['asset', 'issuePriority', 'employee'],
    body: (ids, tag) => ({
      name: `${tag}-issue`, summary: `${tag}-issue`, asset_id: ids.asset, priority_id: ids.issuePriority,
      reported_by_id: ids.employee, reported_at: '2026-09-09T00:00:00Z',
    }) },
  { key: 'purchaseOrder', path: '/api/v1/purchase-orders', dependsOn: ['vendor', 'partLocation'],
    body: (ids, tag) => ({
      number: `${tag}-PO`, description: `${tag} purchase order`,
      vendor_id: ids.vendor, destination_id: ids.partLocation,
    }) },
  { key: 'serviceEntry', path: '/api/v1/service-entries', dependsOn: ['asset', 'vendor'],
    body: (ids, tag) => ({
      asset_id: ids.asset, vendor_id: ids.vendor,
      reference: `${tag}-SE`, started_at: '2026-09-09T00:00:00Z',
    }) },
  { key: 'fuelEntry', path: '/api/v1/assets/{asset}/fuel-entries', dependsOn: ['asset', 'employee', 'vendor'],
    body: (ids, tag) => ({
      employee_id: ids.employee, vendor_id: ids.vendor, date: '2026-09-09T00:00:00Z',
      fuel_type: 'DIESEL', quantity: '10', unit_cost: '10', total_cost: '100',
      odometer: '1000', reference: `${tag}-FE`,
    }) },
  { key: 'partInventory', path: '/api/v1/parts/{part}/inventory', dependsOn: ['part', 'partLocation'],
    body: (ids, tag) => ({
      location_id: ids.partLocation, available_quantity: '10',
      aisle: `${tag}-aisle`, row: 'R1', bin: 'B1',
    }) },
  { key: 'warranty', path: '/api/v1/warranties', dependsOn: ['vendor', 'asset'],
    body: (ids, tag) => ({
      provider_id: ids.vendor, asset_id: ids.asset,
      start_date: '2026-09-09T00:00:00Z', end_date: '2027-09-09T00:00:00Z',
      terms: `${tag}-warranty`,
    }) },
  { key: 'weeklyMileageGoal', path: '/api/v1/weekly-mileage-goals', dependsOn: [],
    body: (_, tag) => ({
      service_type: `${tag}-goal`, rate_per_mile: '1.5',
      weekly_mileage_goal: 100, units_per_service: 1,
    }) },
  { key: 'trailerAssignment', path: '/api/v1/assets/{asset}/trailer-assignments', dependsOn: ['asset', 'trailerAsset'],
    body: (ids, tag) => ({ trailer_id: ids.trailerAsset, notes: `${tag}-assignment` }) },
  { key: 'inspectionFormItem', path: '/api/v1/inspection-forms/{inspectionForm}/items', dependsOn: ['inspectionForm'],
    body: (_, tag) => ({ label: `${tag}-form-item`, position: 1, item_type: 'PASS_FAIL' }) },

  // Level 3 — line items and logs.
  { key: 'workOrderLineItem', path: '/api/v1/work-orders/{workOrder}/line-items', dependsOn: ['workOrder'],
    body: (_, tag) => ({ title: `${tag}-wo-line`, description: `${tag} work order line item` }) },
  { key: 'purchaseOrderLineItem', path: '/api/v1/purchase-orders/{purchaseOrder}/line-items', dependsOn: ['purchaseOrder', 'part'],
    body: (ids, tag) => ({ part_id: ids.part, quantity: '2', unit_cost: '5', position: 1 }) },
  { key: 'serviceEntryLineItem', path: '/api/v1/service-entries/{serviceEntry}/line-items', dependsOn: ['serviceEntry', 'serviceTask'],
    body: (ids, tag) => ({ service_task_id: ids.serviceTask, description: `${tag}-se-line` }) },
  { key: 'wheelPosition', path: '/api/v1/axle-definitions/{axleDefinition}/wheel-positions', dependsOn: ['axleDefinition'],
    body: (_, tag) => ({ code: 'L1', side: 'L', slot: 1 }) },
  { key: 'journalEntry', path: '/api/v1/inventory-journal-entries', dependsOn: ['part', 'partInventory', 'adjustmentReason'],
    body: (ids, tag) => ({
      part_id: ids.part, part_location_detail_id: ids.partInventory,
      adjustment_quantity: '5', unit_cost: '1',
      reason_id: ids.adjustmentReason, notes: `${tag}-journal`,
    }) },

  // Level 4 — grandchildren.
  { key: 'workOrderSubLineItem', path: '/api/v1/work-order-line-items/{workOrderLineItem}/sub-line-items', dependsOn: ['workOrderLineItem', 'part'],
    body: (ids, tag) => ({ part_id: ids.part, quantity: '1', unit_cost: '3', description: `${tag}-sub-line` }) },
  { key: 'laborEntry', path: '/api/v1/work-order-sub-line-items/{workOrderSubLineItem}/labor-entries', dependsOn: ['workOrderSubLineItem', 'employee'],
    body: (ids, tag) => ({ technician_id: ids.employee, started_at: '2026-09-09T00:00:00Z', duration_seconds: 3600 }) },
]

export function topoSort(plan) {
  const byKey = new Map(plan.map(s => [s.key, s]))
  for (const step of plan) {
    for (const dep of step.dependsOn) {
      if (!byKey.has(dep)) {
        throw new Error(`${step.key} has unknown dependency ${dep}`)
      }
    }
  }

  const sorted = []
  const state = new Map() // key -> 'visiting' | 'done'

  const visit = (step, trail) => {
    const mark = state.get(step.key)
    if (mark === 'done') return
    if (mark === 'visiting') {
      throw new Error(`cycle in fixture plan: ${[...trail, step.key].join(' -> ')}`)
    }
    state.set(step.key, 'visiting')
    for (const dep of step.dependsOn) visit(byKey.get(dep), [...trail, step.key])
    state.set(step.key, 'done')
    sorted.push(step)
  }

  for (const step of plan) visit(step, [])
  return sorted
}

// Fixture paths reference earlier fixtures by key, e.g.
// /api/v1/assets/{asset}/fuel-entries. Placeholders resolve from ids.
function resolveFixturePath(path, ids) {
  return path.replace(/\{(\w+)\}/g, (match, key) =>
    key in ids ? String(ids[key]) : match
  )
}

export async function buildFixtures(client, runId) {
  const tag = `ZZ-TEST-${runId}`
  const ids = {}
  const created = []
  const failed = []

  for (const step of topoSort(FIXTURE_PLAN)) {
    const missing = step.dependsOn.filter(d => !(d in ids))
    if (missing.length) {
      failed.push({ key: step.key, status: null, body: `skipped: unmet dependencies ${missing.join(', ')}` })
      continue
    }
    const path = resolveFixturePath(step.path, ids)
    const res = await client.request('POST', path, {
      body: step.body(ids, tag),
      opKey: `POST ${step.path} (fixture)`,
    })
    if (res.status >= 200 && res.status < 300 && res.body?.id != null) {
      // Gin silently ignores unknown JSON fields: if a fixture body uses a
      // field name the Create DTO does not declare, the server still
      // returns 2xx and a row still exists — it just stores nothing we
      // sent, so no ZZ-TEST tag is anywhere in the response. That row is
      // real but untaggable, so the tag-based human recovery sweep can
      // never find it. Checking the RESPONSE (not the request we sent) is
      // the only way to catch this: it's what actually landed in the DB.
      // The row still goes into `created` so teardown deletes it this run,
      // but it is not treated as a healthy fixture and nothing downstream
      // may build on it.
      const taggedInResponse = UNTAGGABLE.has(step.key) || JSON.stringify(res.body).includes(tag)
      created.push({ key: step.key, path, id: res.body.id })
      if (taggedInResponse) {
        ids[step.key] = res.body.id
      } else {
        failed.push({
          key: step.key, status: res.status,
          body: 'created but the response carries no run tag — a field name is probably absent from the Create DTO and was silently ignored; this row cannot be recovered by tag',
        })
      }
    } else {
      failed.push({ key: step.key, status: res.status, body: res.body })
    }
  }

  return { runId, tag, ids, created, failed }
}

export async function teardownFixtures(client, graph) {
  const deleted = []
  const archived = []
  const reversed = []
  const failed = []

  for (const row of [...graph.created].reverse()) {
    const target = `${row.path}/${row.id}`
    const res = await client.request('DELETE', target, { opKey: `DELETE ${row.key} (teardown)` })
    if (res.status >= 200 && res.status < 300) {
      deleted.push(`${row.key}#${row.id}`)
      continue
    }
    if (ARCHIVABLE.has(row.key)) {
      const arc = await client.request('POST', `${target}/archive`, { opKey: `ARCHIVE ${row.key} (teardown)` })
      if (arc.status >= 200 && arc.status < 300) {
        archived.push(`${row.key}#${row.id}`)
        continue
      }
    }
    if (REVERSIBLE.has(row.key)) {
      const rev = await client.request('POST', `${target}/reverse`, {
        body: { notes: `${graph.tag} teardown reversal` },
        opKey: `REVERSE ${row.key} (teardown)`,
      })
      if (rev.status >= 200 && rev.status < 300) {
        reversed.push(`${row.key}#${row.id}`)
        continue
      }
    }
    failed.push({ key: `${row.key}#${row.id}`, status: res.status, body: res.body })
  }

  return { deleted, archived, reversed, failed }
}
