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
// serviceTaskPart (dto/service_task_part.go CreateServiceTaskPartRequest) has
// only part_id (FK), quantity (decimal) and position (int) — none free
// text, same reasoning as wheelPosition above.
export const UNTAGGABLE = new Set(['laborEntry', 'purchaseOrderLineItem', 'wheelPosition', 'serviceTaskPart'])

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

  // Level 2b — asset singleton sub-types. These are upserts (PUT), not
  // creates: crud.SingletonHandler.Register wires GET/PUT/DELETE on
  // /{parent}/:id/{child} with no child id of its own (see
  // internal/platform/crud/singleton.go). The Upsert response is the full
  // *Response DTO for the row, which has no "id" field (VehicleResponse,
  // TrailerResponse and VehicleAxleConfigResponse key off asset_id /
  // vehicle_id instead) — buildFixtures below special-cases `singleton`
  // steps so that absence doesn't read as a failed create. Bodies are
  // built from the real request structs (internal/http/dto/vehicle.go,
  // trailer.go, vehicle_axle_config.go), not the OpenAPI spec, per the
  // Gin-silently-drops-unknown-fields hazard.
  //
  // All three depend on `asset`, not `trailerAsset`: the sweep's own
  // idsForOperation resolves the {id} in GET /assets/{id}/vehicle (and
  // /trailer, /axle-config) via PATH_FIXTURE_MAP['assets'] = 'asset'
  // unconditionally — it has no awareness of which singleton child
  // follows, and TrailerStore.Upsert (internal/http/handler/trailer.go)
  // has no vehicle_type gate. Putting the trailer fixture on `trailerAsset`
  // would create it under an id the sweep's GET never queries, so the
  // check it exists to fix would still 404.
  // singleton: true tells both buildFixtures and teardownFixtures that this
  // row has no id of its own — it is deleted at its own path
  // (DELETE /assets/{asset}/vehicle), never at `${path}/${id}`. Because
  // topoSort creates these after `asset` (they depend on it) they land
  // later in `created`, so the reverse teardown walk removes them BEFORE
  // asset — required, since a surviving singleton refuses the asset's own
  // DELETE and pins whatever catalog rows the asset references.
  { key: 'vehicle', path: '/api/v1/assets/{asset}/vehicle', method: 'PUT',
    singleton: true, dependsOn: ['asset'],
    body: (_, tag) => ({ engine_serial: `${tag}-vehicle` }) },
  { key: 'trailer', path: '/api/v1/assets/{asset}/trailer', method: 'PUT',
    singleton: true, dependsOn: ['asset'],
    body: (_, tag) => ({ owner_name: `${tag}-trailer` }) },
  { key: 'axleConfig', path: '/api/v1/assets/{asset}/axle-config', method: 'PUT',
    singleton: true, dependsOn: ['asset', 'axleTemplate'],
    body: (ids, tag) => ({ template_id: ids.axleTemplate, display_name: `${tag}-axle-config` }) },

  // service-tasks/{id}/parts/{child_id} needs its own row: 'parts' resolves
  // flatly to the catalog `part` fixture (PATH_FIXTURE_MAP), which 404s
  // under a service task. dto/service_task_part.go's CreateServiceTaskPartRequest
  // has no free-text field, so this is UNTAGGABLE (see above), same as
  // purchaseOrderLineItem.
  { key: 'serviceTaskPart', path: '/api/v1/service-tasks/{serviceTask}/parts', dependsOn: ['serviceTask', 'part'],
    body: (ids, tag) => ({ part_id: ids.part, quantity: '1', position: 1 }) },

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
    const method = step.method ?? 'POST'
    const res = await client.request(method, path, {
      body: step.body(ids, tag),
      opKey: `${method} ${step.path} (fixture)`,
    })
    // A singleton upsert (PUT .../vehicle, .../trailer, .../axle-config) has
    // no `id` field in its response at all — it is a 1:1 child addressed by
    // its parent id, not a row with its own identity — so success there is
    // just a 2xx, not a 2xx-with-id.
    const succeeded = res.status >= 200 && res.status < 300
    const identified = step.singleton ? succeeded : succeeded && res.body?.id != null
    if (identified) {
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
      // Singleton rows go into `created` too — they must be torn down, just
      // not at `${path}/${id}` (there is no id). teardownFixtures reads the
      // `singleton` flag on the row itself to know to DELETE `path` as-is.
      created.push({ key: step.key, path, id: step.singleton ? null : res.body.id, singleton: !!step.singleton })
      if (taggedInResponse) {
        ids[step.key] = step.singleton ? true : res.body.id
      } else {
        failed.push({
          key: step.key, status: res.status,
          body: 'created but the response carries no run tag — a field name is probably absent from the Create DTO and was silently ignored; this row cannot be recovered by tag',
        })
      }
    } else if (succeeded) {
      // A 2xx create with no `id` in the response (and not a declared
      // singleton) is not a validation failure — a row plausibly exists in
      // the database right now. There is no id to build `${path}/${id}`
      // from, so it cannot be tracked in `created` and teardown cannot
      // remove it automatically. Not reachable against today's DTOs (every
      // non-singleton Create*Response declares `id`), but this must never
      // be silent the way a generic `res.body` dump here would be: it has
      // to say in plain words that a row may have leaked and nobody but a
      // human, searching by this run's tag, can find it.
      failed.push({
        key: step.key, status: res.status,
        body: `created (${res.status}) but the response carries no id — this row cannot be tracked for teardown and may still exist in the database; it must be found and removed by hand, by searching for the tag ${tag}`,
      })
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
  const pending = [] // rows every fallback refused; verified before becoming `failed`

  for (const row of [...graph.created].reverse()) {
    // A singleton (vehicle/trailer/axle-config) has no id of its own — it is
    // addressed, and deleted, at its own path. Every other row is deleted at
    // `${path}/${id}` as before.
    const target = row.singleton ? row.path : `${row.path}/${row.id}`
    const label = row.singleton ? row.key : `${row.key}#${row.id}`
    const res = await client.request('DELETE', target, { opKey: `DELETE ${row.key} (teardown)` })
    if (res.status >= 200 && res.status < 300) {
      deleted.push(label)
      continue
    }
    if (ARCHIVABLE.has(row.key)) {
      const arc = await client.request('POST', `${target}/archive`, { opKey: `ARCHIVE ${row.key} (teardown)` })
      if (arc.status >= 200 && arc.status < 300) {
        archived.push(label)
        continue
      }
    }
    if (REVERSIBLE.has(row.key)) {
      // No body: the reverse endpoint binds no JSON request (see
      // internal/http/handler/inventory_journal_entry.go) — sending one
      // would be dropped silently, the same class of bug this whole task
      // exists to catch, so don't pin a false belief that it's accepted.
      const rev = await client.request('POST', `${target}/reverse`, { opKey: `REVERSE ${row.key} (teardown)` })
      if (rev.status >= 200 && rev.status < 300) {
        reversed.push(label)
        continue
      }
    }
    pending.push({ row, target, label, status: res.status, body: res.body })
  }

  // A row can land here as a false failure: something later in this same
  // reverse walk (e.g. deleting partInventory, which cascades onto
  // journalEntry via ON DELETE CASCADE) can remove a row whose own DELETE
  // and every fallback were refused earlier in the walk. Reporting it as
  // failed would be a false statement in the audit report, so verify with
  // a GET before finalising: 404 means it is genuinely gone (promote to
  // deleted); anything else (including network failure) means it is still
  // there, or its state is unknown, and it must stay failed rather than be
  // excused.
  const failed = []
  for (const { row, target, label, status, body } of pending) {
    const check = await client.request('GET', target, { opKey: `GET ${row.key} (teardown verify)` })
    if (check.status === 404) {
      deleted.push(label)
    } else {
      failed.push({ key: label, status, body })
    }
  }

  return { deleted, archived, reversed, failed }
}
