import {
  checkUnauthenticated, checkNotFound, checkHappyPath,
  checkConformance, checkListContract, checkRetiredVerb, fillPath,
} from './checks.mjs'

// Collection segment -> fixture key. The sweep resolves {id} from the segment
// immediately preceding it, and {child_id} from the segment after it.
export const PATH_FIXTURE_MAP = {
  'vendors': 'vendor',
  'locations': 'location',
  'asset-types': 'assetType',
  'asset-statuses': 'assetStatus',
  'part-categories': 'partCategory',
  'part-manufacturers': 'partManufacturer',
  'measurement-units': 'measurementUnit',
  'part-locations': 'partLocation',
  'inventory-adjustment-reasons': 'adjustmentReason',
  'fuel-types': 'fuelType',
  'trailer-classifications': 'trailerClassification',
  'issue-priorities': 'issuePriority',
  'work-order-statuses': 'workOrderStatus',
  'vehicle-makes': 'vehicleMake',
  'vehicle-models': 'vehicleModel',
  'axle-templates': 'axleTemplate',
  'axle-definitions': 'axleDefinition',
  'service-tasks': 'serviceTask',
  'roles': 'role',
  'groups': 'group',
  'faults': 'fault',
  'inspection-forms': 'inspectionForm',
  'tire-models': 'tireModel',
  'tires': 'tire',
  'parts': 'part',
  'assets': 'asset',
  'employees': 'employee',
  'work-orders': 'workOrder',
  'work-order-line-items': 'workOrderLineItem',
  'work-order-sub-line-items': 'workOrderSubLineItem',
  'issues': 'issue',
  'purchase-orders': 'purchaseOrder',
  'service-entries': 'serviceEntry',
  'service-entry-line-items': 'serviceEntryLineItem',
  'warranties': 'warranty',
  'weekly-mileage-goals': 'weeklyMileageGoal',
  'inventory-journal-entries': 'journalEntry',
  'line-items': 'workOrderLineItem',
  'sub-line-items': 'workOrderSubLineItem',
  'labor-entries': 'laborEntry',
  'wheel-positions': 'wheelPosition',
  'definitions': 'axleDefinition',
  'inventory': 'partInventory',
  'fuel-entries': 'fuelEntry',
  'trailer-assignments': 'trailerAssignment',
  'items': 'inspectionFormItem',
}

export function idsForOperation(op, graph) {
  const segments = op.path.split('/').filter(Boolean)
  const ids = {}

  for (let i = 0; i < segments.length; i++) {
    const seg = segments[i]
    if (seg !== '{id}' && seg !== '{child_id}') continue

    // {id} belongs to the collection before it; {child_id} to the collection
    // that follows {id}.
    const collection = segments[i - 1]
    const fixtureKey = PATH_FIXTURE_MAP[collection]
    if (!fixtureKey) return null
    const value = graph.ids[fixtureKey]
    if (value == null) return null
    ids[seg === '{id}' ? 'id' : 'child_id'] = value
  }

  return ids
}

export async function sweep(client, spec, ops, graph, { onProgress } = {}) {
  const results = []
  const coverage = []
  const byOp = new Map()

  const push = r => {
    if (!r) return
    results.push(r)
    if (!byOp.has(r.opKey)) byOp.set(r.opKey, [])
    byOp.get(r.opKey).push(r)
  }

  for (const op of ops) {
    onProgress?.(op)

    push(await checkUnauthenticated(client, op))
    push(await checkRetiredVerb(client, op))
    push(await checkNotFound(client, op))
    push(await checkListContract(client, op))

    const ids = idsForOperation(op, graph)
    if (ids === null) {
      coverage.push({
        opKey: op.opKey,
        tested: false,
        reason: 'no fixture row satisfies this path',
      })
      continue
    }

    if (op.method === 'GET') {
      const happy = await checkHappyPath(client, op, ids, undefined)
      push(happy)
      push(checkConformance(spec, op, happy.response))
      coverage.push({ opKey: op.opKey, tested: true, reason: null })
      continue
    }

    if (op.method === 'PUT') {
      // A round-trip, not a blind write: read the row back first, then hand
      // the exact same body straight back. This proves the write path end
      // to end without inventing field values that could blank a fixture
      // row mid-sweep. A 4xx on the write-back is a genuine finding, so it
      // is still recorded as a normal happy-path CheckResult, not swallowed.
      const path = fillPath(op.path, ids)
      const getRes = await client.request('GET', path, { opKey: op.opKey })
      if (getRes.status < 200 || getRes.status >= 300) {
        coverage.push({
          opKey: op.opKey,
          tested: false,
          reason: 'could not read the row to round-trip it',
        })
        continue
      }
      const happy = await checkHappyPath(client, op, ids, getRes.body)
      push(happy)
      push(checkConformance(spec, op, happy.response))
      coverage.push({ opKey: op.opKey, tested: true, reason: null })
      continue
    }

    if (op.method === 'POST') {
      // Collection POST is proved by the fixture build succeeding; action
      // POST (approve/void/etc.) belongs to the workflow suites, which know
      // the legal states. Neither is safe to fire blind against a live DB.
      coverage.push({
        opKey: op.opKey,
        tested: false,
        reason: 'collection POST is proved by the fixture build; action POST is owned by the workflow suites',
      })
      continue
    }

    if (op.method === 'DELETE') {
      // Destructive verbs are exercised by teardown and the workflow suites,
      // never by the generic sweep: a DELETE here would dissolve the fixture
      // graph the remaining operations depend on.
      coverage.push({ opKey: op.opKey, tested: false, reason: 'covered by teardown' })
      continue
    }

    // Defensive fallback: every operation must land in results or coverage,
    // never fall through silently. The live spec has no verb beyond
    // GET/PUT/POST/DELETE today, but if one appears it must be accounted
    // for rather than disappear from the report.
    coverage.push({ opKey: op.opKey, tested: false, reason: `unhandled verb ${op.method}` })
  }

  return { results, byOp, coverage }
}
