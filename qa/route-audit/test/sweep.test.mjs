import { test } from 'node:test'
import assert from 'node:assert/strict'
import { sweep, idsForOperation, PATH_FIXTURE_MAP } from '../lib/sweep.mjs'

const graph = { ids: { workOrder: 10, workOrderLineItem: 20, vendor: 30, asset: 40 } }

test('idsForOperation maps a top-level detail path to its fixture', () => {
  const op = { path: '/api/v1/vendors/{id}' }
  assert.deepEqual(idsForOperation(op, graph), { id: 30 })
})

test('idsForOperation maps a nested path to parent and child fixtures', () => {
  const op = { path: '/api/v1/work-orders/{id}/line-items/{child_id}' }
  assert.deepEqual(idsForOperation(op, graph), { id: 10, child_id: 20 })
})

test('idsForOperation returns an empty object for a collection path', () => {
  const op = { path: '/api/v1/vendors' }
  assert.deepEqual(idsForOperation(op, graph), {})
})

test('idsForOperation returns null when no fixture matches the collection', () => {
  const op = { path: '/api/v1/unknown-thing/{id}' }
  assert.equal(idsForOperation(op, graph), null)
})

test('idsForOperation returns null when the mapped fixture failed to build', () => {
  const op = { path: '/api/v1/assets/{id}/fuel-entries/{child_id}' }
  assert.equal(idsForOperation(op, { ids: { asset: 40 } }), null)
})

test('every PATH_FIXTURE_MAP value names a key the fixture plan can produce', async () => {
  const { FIXTURE_PLAN } = await import('../lib/fixtures.mjs')
  const keys = new Set(FIXTURE_PLAN.map(s => s.key))
  for (const [collection, fixtureKey] of Object.entries(PATH_FIXTURE_MAP)) {
    assert.ok(keys.has(fixtureKey), `${collection} maps to unknown fixture ${fixtureKey}`)
  }
})

// ---------------------------------------------------------------------------
// Verb-policy tests: a stub client records every call so we can prove the
// sweep never sends a request that could create or blank a row it did not
// already own via the fixture graph.
// ---------------------------------------------------------------------------

function makeStubClient(handlers, fallback = { status: 200, body: {}, headers: {}, durationMs: 1 }) {
  const calls = []
  return {
    calls,
    async request(method, path, opts = {}) {
      calls.push({ method, path, opts })
      const handler = handlers[`${method} ${path}`]
      if (typeof handler === 'function') return handler(opts)
      return handler ?? fallback
    },
  }
}

const emptySpec = {}

test('sweep: GET operation produces a happy-path result and no coverage exclusion', async () => {
  const op = {
    opKey: 'GET /api/v1/vendors/{id}',
    method: 'GET',
    path: '/api/v1/vendors/{id}',
    tags: [],
    params: [],
    successStatus: 200,
    successSchema: null,
    requestSchema: null,
  }
  const client = makeStubClient({
    'GET /api/v1/vendors/42': { status: 200, body: { id: 42, name: 'Acme' }, headers: {}, durationMs: 1 },
  })
  const g = { ids: { vendor: 42 } }

  const { results, coverage } = await sweep(client, emptySpec, [op], g)

  const happy = results.find(r => r.opKey === op.opKey && r.check === 'happy-path')
  assert.ok(happy, 'expected a happy-path result')
  assert.equal(happy.ok, true)

  const cov = coverage.find(c => c.opKey === op.opKey)
  assert.deepEqual(cov, { opKey: op.opKey, tested: true, reason: null })
})

test('sweep: POST operation produces a coverage exclusion and sends no real request', async () => {
  const op = {
    opKey: 'POST /api/v1/vendors',
    method: 'POST',
    path: '/api/v1/vendors',
    tags: [],
    params: [],
    successStatus: 201,
    successSchema: null,
    requestSchema: { type: 'object' },
  }
  const client = makeStubClient({})
  const g = { ids: { vendor: 42 } }

  const { results, coverage } = await sweep(client, emptySpec, [op], g)

  // The only calls a POST operation may generate are the generic pre-checks
  // (checkUnauthenticated), which are anonymous and expected to be rejected.
  // No authenticated POST — the one that would actually create a row — may
  // ever be sent by the sweep's own verb dispatch.
  const authenticatedPosts = client.calls.filter(c => c.method === 'POST' && c.opts.anonymous !== true)
  assert.deepEqual(authenticatedPosts, [])

  assert.equal(results.some(r => r.opKey === op.opKey && r.check === 'happy-path'), false)

  const cov = coverage.find(c => c.opKey === op.opKey)
  assert.deepEqual(cov, {
    opKey: op.opKey,
    tested: false,
    reason: 'collection POST is proved by the fixture build; action POST is owned by the workflow suites',
  })
})

test('sweep: DELETE operation produces a coverage exclusion and sends no real request', async () => {
  const op = {
    opKey: 'DELETE /api/v1/vendors/{id}',
    method: 'DELETE',
    path: '/api/v1/vendors/{id}',
    tags: [],
    params: [],
    successStatus: 204,
    successSchema: null,
    requestSchema: null,
  }
  const client = makeStubClient({})
  const g = { ids: { vendor: 777 } }

  const { results, coverage } = await sweep(client, emptySpec, [op], g)

  // No call may ever reach the real fixture row (id 777) — only the
  // pre-check probes against the MISSING_ID placeholder are allowed.
  const callsAgainstRealRow = client.calls.filter(c => c.path.includes('777'))
  assert.deepEqual(callsAgainstRealRow, [])

  assert.equal(results.some(r => r.opKey === op.opKey && r.check === 'happy-path'), false)

  const cov = coverage.find(c => c.opKey === op.opKey)
  assert.deepEqual(cov, { opKey: op.opKey, tested: false, reason: 'covered by teardown' })
})

test('sweep: PUT operation whose preparatory GET succeeds sends the fetched body back unchanged', async () => {
  const op = {
    opKey: 'PUT /api/v1/vendors/{id}',
    method: 'PUT',
    path: '/api/v1/vendors/{id}',
    tags: [],
    params: [],
    successStatus: 200,
    successSchema: null,
    requestSchema: { type: 'object' },
  }
  const fetchedBody = { id: 55, name: 'Acme', rating: 4 }
  let putBodyReceived = null
  const client = makeStubClient({
    'GET /api/v1/vendors/55': { status: 200, body: fetchedBody, headers: {}, durationMs: 1 },
    'PUT /api/v1/vendors/55': (opts) => {
      putBodyReceived = opts.body
      return { status: 200, body: fetchedBody, headers: {}, durationMs: 1 }
    },
  })
  const g = { ids: { vendor: 55 } }

  const { results, coverage } = await sweep(client, emptySpec, [op], g)

  assert.deepEqual(putBodyReceived, fetchedBody)

  const happy = results.find(r => r.opKey === op.opKey && r.check === 'happy-path')
  assert.ok(happy, 'expected a happy-path result for the round-trip PUT')
  assert.equal(happy.ok, true)

  const cov = coverage.find(c => c.opKey === op.opKey)
  assert.deepEqual(cov, { opKey: op.opKey, tested: true, reason: null })
})

test('sweep: PUT operation whose preparatory GET fails produces a coverage exclusion and sends no PUT', async () => {
  const op = {
    opKey: 'PUT /api/v1/vendors/{id}',
    method: 'PUT',
    path: '/api/v1/vendors/{id}',
    tags: [],
    params: [],
    successStatus: 200,
    successSchema: null,
    requestSchema: { type: 'object' },
  }
  const client = makeStubClient({
    'GET /api/v1/vendors/66': { status: 404, body: { error: 'not found' }, headers: {}, durationMs: 1 },
  })
  const g = { ids: { vendor: 66 } }

  const { results, coverage } = await sweep(client, emptySpec, [op], g)

  const putsToRealRow = client.calls.filter(c => c.method === 'PUT' && c.path.includes('66'))
  assert.deepEqual(putsToRealRow, [])

  assert.equal(results.some(r => r.opKey === op.opKey && r.check === 'happy-path'), false)

  const cov = coverage.find(c => c.opKey === op.opKey)
  assert.deepEqual(cov, {
    opKey: op.opKey,
    tested: false,
    reason: 'could not read the row to round-trip it',
  })
})
