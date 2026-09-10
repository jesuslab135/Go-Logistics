// qa/route-audit/lib/fixtures.mjs

// Resources that answer POST .../{id}/archive when a delete is refused.
const ARCHIVABLE = new Set(['vendor', 'part', 'asset', 'serviceTask', 'inspectionForm'])

// Fixtures whose Create DTO has no free-text field, so there is nowhere to put
// the run tag. Teardown still tracks them by id in `created`, and both are
// grandchildren that die with their parent, so cleanup is unaffected — the tag
// is a human-recovery aid, not the deletion mechanism.
export const UNTAGGABLE = new Set(['laborEntry', 'purchaseOrderLineItem'])

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
    body: (_, tag) => ({ name: `${tag}-inspection-form` }) },
  { key: 'tireModel', path: '/api/v1/tire-models', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-tire-model` }) },

  // Level 1 — one hop.
  { key: 'vehicleModel', path: '/api/v1/vehicle-models', dependsOn: ['vehicleMake'],
    body: (ids, tag) => ({ name: `${tag}-model`, make_id: ids.vehicleMake }) },
  { key: 'part', path: '/api/v1/parts', dependsOn: ['partCategory', 'partManufacturer', 'measurementUnit'],
    body: (ids, tag) => ({
      name: `${tag}-part`, part_number: `${tag}-PN`,
      category_id: ids.partCategory, manufacturer_id: ids.partManufacturer,
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
    body: (ids, tag) => ({ serial_number: `${tag}-tire`, tire_model_id: ids.tireModel }) },
  { key: 'axleDefinition', path: '/api/v1/axle-templates/{axleTemplate}/definitions', dependsOn: ['axleTemplate'],
    body: (_, tag) => ({ name: `${tag}-axle-def`, position: 1, wheel_count: 2 }) },
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
  { key: 'purchaseOrder', path: '/api/v1/purchase-orders', dependsOn: ['vendor', 'location'],
    body: (ids, tag) => ({
      number: `${tag}-PO`, description: `${tag} purchase order`,
      vendor_id: ids.vendor, destination_id: ids.location,
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
    body: (ids, tag) => ({ location_id: ids.partLocation, quantity: '10', notes: `${tag}-inv` }) },
  { key: 'warranty', path: '/api/v1/warranties', dependsOn: ['vendor', 'asset'],
    body: (ids, tag) => ({
      provider_id: ids.vendor, asset_id: ids.asset,
      start_date: '2026-09-09T00:00:00Z', end_date: '2027-09-09T00:00:00Z',
      terms: `${tag}-warranty`,
    }) },
  { key: 'weeklyMileageGoal', path: '/api/v1/weekly-mileage-goals', dependsOn: ['asset'],
    body: (ids, tag) => ({ asset_id: ids.asset, target_miles: '100', notes: `${tag}-goal` }) },
  { key: 'trailerAssignment', path: '/api/v1/assets/{asset}/trailer-assignments', dependsOn: ['asset', 'trailerAsset'],
    body: (ids, tag) => ({ trailer_id: ids.trailerAsset, notes: `${tag}-assignment` }) },
  { key: 'inspectionFormItem', path: '/api/v1/inspection-forms/{inspectionForm}/items', dependsOn: ['inspectionForm'],
    body: (_, tag) => ({ label: `${tag}-form-item`, position: 1, item_type: 'PASS_FAIL' }) },

  // Level 3 — line items and logs.
  { key: 'workOrderLineItem', path: '/api/v1/work-orders/{workOrder}/line-items', dependsOn: ['workOrder', 'serviceTask'],
    body: (ids, tag) => ({ service_task_id: ids.serviceTask, description: `${tag}-wo-line` }) },
  { key: 'purchaseOrderLineItem', path: '/api/v1/purchase-orders/{purchaseOrder}/line-items', dependsOn: ['purchaseOrder', 'part'],
    body: (ids, tag) => ({ part_id: ids.part, quantity: '2', unit_cost: '5', position: 1 }) },
  { key: 'serviceEntryLineItem', path: '/api/v1/service-entries/{serviceEntry}/line-items', dependsOn: ['serviceEntry', 'serviceTask'],
    body: (ids, tag) => ({ service_task_id: ids.serviceTask, description: `${tag}-se-line` }) },
  { key: 'wheelPosition', path: '/api/v1/axle-definitions/{axleDefinition}/wheel-positions', dependsOn: ['axleDefinition'],
    body: (_, tag) => ({ label: `${tag}-wheel`, position: 1 }) },
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
      ids[step.key] = res.body.id
      created.push({ key: step.key, path, id: res.body.id })
    } else {
      failed.push({ key: step.key, status: res.status, body: res.body })
    }
  }

  return { runId, tag, ids, created, failed }
}

export async function teardownFixtures(client, graph) {
  const deleted = []
  const archived = []
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
    failed.push({ key: `${row.key}#${row.id}`, status: res.status, body: res.body })
  }

  return { deleted, archived, failed }
}
