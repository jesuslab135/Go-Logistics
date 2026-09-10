// qa/route-audit/test/fixtures.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { topoSort, FIXTURE_PLAN, UNTAGGABLE, buildFixtures, teardownFixtures } from '../lib/fixtures.mjs'

test('topoSort places dependencies before dependents', () => {
  const plan = [
    { key: 'c', dependsOn: ['b'] },
    { key: 'a', dependsOn: [] },
    { key: 'b', dependsOn: ['a'] },
  ]
  assert.deepEqual(topoSort(plan).map(s => s.key), ['a', 'b', 'c'])
})

test('topoSort throws on a cycle rather than looping forever', () => {
  const plan = [
    { key: 'a', dependsOn: ['b'] },
    { key: 'b', dependsOn: ['a'] },
  ]
  assert.throws(() => topoSort(plan), /cycle/i)
})

test('topoSort throws when a dependency is not in the plan', () => {
  const plan = [{ key: 'a', dependsOn: ['ghost'] }]
  assert.throws(() => topoSort(plan), /unknown dependency/i)
})

test('every FIXTURE_PLAN dependency refers to a real step', () => {
  const keys = new Set(FIXTURE_PLAN.map(s => s.key))
  for (const step of FIXTURE_PLAN) {
    for (const dep of step.dependsOn) {
      assert.ok(keys.has(dep), `${step.key} depends on unknown ${dep}`)
    }
  }
})

test('FIXTURE_PLAN sorts without a cycle', () => {
  assert.equal(topoSort(FIXTURE_PLAN).length, FIXTURE_PLAN.length)
})

test('every taggable fixture body carries the run tag so teardown can find it', () => {
  const ids = Object.fromEntries(FIXTURE_PLAN.map(s => [s.key, 1]))
  for (const step of FIXTURE_PLAN) {
    if (UNTAGGABLE.has(step.key)) continue
    const body = step.body(ids, 'ZZ-TEST-42')
    const serialised = JSON.stringify(body)
    assert.ok(
      serialised.includes('ZZ-TEST-42'),
      `${step.key} body does not carry the run tag: ${serialised}`
    )
  }
})

test('UNTAGGABLE fixtures have no free-text field to tag', () => {
  // Some UNTAGGABLE bodies legitimately carry string-typed values that are
  // NOT free text: decimal amounts serialised as strings (e.g. quantity:
  // '2'), ISO-8601 timestamps (e.g. started_at), and short domain codes
  // whose DB column is only a few characters wide (e.g. wheelPosition's
  // code/side, which top out at 10 and 1 characters — nowhere near enough
  // to hold `ZZ-TEST-<runId>` even if the DTO allowed free text there).
  // Those are format-constrained by the API and cannot hold a human-
  // readable tag no matter what — they are not the "hole" this test
  // guards against. A field that holds arbitrary text (name, notes,
  // description, label, terms, ...) is the hole: it would mean the DTO
  // grew a place for the tag and the exemption should be removed. So this
  // test fails on any string field that is NOT a recognised constrained
  // format, rather than on every string field. Free text in this codebase
  // is either much longer than a short code or contains whitespace/
  // punctuation a code never has, so a short token with no whitespace is
  // a reasonable proxy for "constrained", not "free text".
  const ISO_TIMESTAMP = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/
  const DECIMAL_STRING = /^\d+(\.\d+)?$/
  const SHORT_CODE = /^[A-Za-z0-9_-]{1,20}$/
  const ids = Object.fromEntries(FIXTURE_PLAN.map(s => [s.key, 1]))
  const byKey = new Map(FIXTURE_PLAN.map(s => [s.key, s]))
  for (const key of UNTAGGABLE) {
    const step = byKey.get(key)
    assert.ok(step, `UNTAGGABLE names unknown step ${key}`)
    const body = step.body(ids, 'ZZ-TEST-42')
    for (const [field, value] of Object.entries(body)) {
      if (typeof value !== 'string') continue
      assert.ok(
        ISO_TIMESTAMP.test(value) || DECIMAL_STRING.test(value) || SHORT_CODE.test(value),
        `${key}.${field} ("${value}") looks like free text — UNTAGGABLE exemption may be stale, tag it instead`
      )
    }
  }
})

test('teardown fallback chain routes each resource to the correct mechanism', async () => {
  const calls = []
  const stubClient = {
    async request(method, path, opts) {
      calls.push({ method, path, body: opts?.body })
      if (method === 'DELETE') {
        return { status: 405, body: { error: { code: 'route_retired' } } }
      }
      if (method === 'POST' && path.endsWith('/reverse')) {
        return { status: 201, body: { id: 999 } }
      }
      if (method === 'POST' && path.endsWith('/archive')) {
        return { status: 200, body: { id: 998 } }
      }
      if (method === 'GET') {
        // location has no fallback and genuinely never gets removed by
        // anything else in this stub's walk, so the verify-before-failing
        // GET must find it still there.
        return { status: 200, body: { id: 11 } }
      }
      throw new Error(`stub client received unexpected call: ${method} ${path}`)
    },
  }

  const graph = {
    runId: 'stub',
    tag: 'ZZ-TEST-99',
    ids: {},
    created: [
      { key: 'journalEntry', path: '/api/v1/inventory-journal-entries', id: 3 },
      { key: 'vendor', path: '/api/v1/vendors', id: 7 },
      { key: 'location', path: '/api/v1/locations', id: 11 },
    ],
    failed: [],
  }

  const result = await teardownFixtures(stubClient, graph)

  // journalEntry -> DELETE refused -> POST .../reverse -> reversed
  assert.deepEqual(result.reversed, ['journalEntry#3'])
  // vendor -> DELETE refused -> POST .../archive -> archived
  assert.deepEqual(result.archived, ['vendor#7'])
  // location is in neither ARCHIVABLE nor REVERSIBLE -> DELETE refused,
  // verify GET confirms it still exists (200) -> stays failed
  assert.deepEqual(result.failed.map(f => f.key), ['location#11'])
  assert.deepEqual(result.deleted, [])

  const reverseCall = calls.find(c => c.path === '/api/v1/inventory-journal-entries/3/reverse')
  assert.ok(reverseCall, 'expected a POST to the reverse endpoint')
  // The reverse endpoint binds no JSON body (see
  // internal/http/handler/inventory_journal_entry.go) — teardown must not
  // send one, or send it and have it silently ignored while a test claims
  // otherwise.
  assert.equal(reverseCall.body, undefined)

  const archiveCall = calls.find(c => c.path === '/api/v1/vendors/7/archive')
  assert.ok(archiveCall, 'expected a POST to the archive endpoint')

  // location must not have triggered a reverse or archive call, only the
  // DELETE and the final verification GET.
  assert.ok(!calls.some(c => c.method === 'POST' && c.path.startsWith('/api/v1/locations/11/')))
  const locationGet = calls.find(c => c.method === 'GET' && c.path === '/api/v1/locations/11')
  assert.ok(locationGet, 'expected a verification GET for the row that stayed failed')
})

test('teardown promotes a row to deleted when a post-walk GET proves it is 404 gone', async () => {
  // journalEntry: DELETE refused, then reverse also refused (409 - already
  // reversed by something else this run), so it would land in `failed`.
  // But the verification GET afterwards returns 404 - something later in
  // the same reverse walk (a cascade from deleting partInventory, in the
  // real system) removed it anyway. It must be reported as deleted, not
  // failed, or the audit report states something false.
  const stubClient = {
    async request(method, path) {
      if (method === 'DELETE') return { status: 405, body: { error: { code: 'route_retired' } } }
      if (method === 'POST' && path.endsWith('/reverse')) {
        return { status: 409, body: { error: { code: 'already_reversed' } } }
      }
      if (method === 'GET') return { status: 404, body: { error: { code: 'not_found' } } }
      throw new Error(`stub client received unexpected call: ${method} ${path}`)
    },
  }

  const graph = {
    runId: 'stub', tag: 'ZZ-TEST-99', ids: {},
    created: [{ key: 'journalEntry', path: '/api/v1/inventory-journal-entries', id: 4 }],
    failed: [],
  }

  const result = await teardownFixtures(stubClient, graph)

  assert.deepEqual(result.deleted, ['journalEntry#4'])
  assert.deepEqual(result.failed, [])
  assert.deepEqual(result.reversed, [])
})

test('teardown leaves a row failed when the post-walk GET proves it still exists', async () => {
  // Same shape as above, but the verification GET returns 200: the row
  // really is still there. It must stay in `failed` - a promotion here
  // would hide a genuine leftover row from the audit report.
  const stubClient = {
    async request(method, path) {
      if (method === 'DELETE') return { status: 405, body: { error: { code: 'route_retired' } } }
      if (method === 'POST' && path.endsWith('/reverse')) {
        return { status: 409, body: { error: { code: 'already_reversed' } } }
      }
      if (method === 'GET') return { status: 200, body: { id: 4 } }
      throw new Error(`stub client received unexpected call: ${method} ${path}`)
    },
  }

  const graph = {
    runId: 'stub', tag: 'ZZ-TEST-99', ids: {},
    created: [{ key: 'journalEntry', path: '/api/v1/inventory-journal-entries', id: 4 }],
    failed: [],
  }

  const result = await teardownFixtures(stubClient, graph)

  assert.deepEqual(result.failed.map(f => f.key), ['journalEntry#4'])
  assert.deepEqual(result.deleted, [])
})

test('buildFixtures treats a created-but-untagged response as failed, not healthy', async () => {
  // Simulates the Gin silent-unknown-field-drop failure mode: the server
  // accepts the POST (2xx, real id) but the row it actually stored carries
  // none of the fields we sent, so the run tag never appears in the
  // response body. Every step gets this same stub response.
  let nextId = 1
  const stubClient = {
    async request(method, _path, _opts) {
      if (method !== 'POST') throw new Error(`unexpected ${method} call`)
      return { status: 201, body: { id: nextId++, note: 'server stored nothing we sent' } }
    },
  }

  const graph = await buildFixtures(stubClient, 'untagged-stub')

  // vendor has no dependencies, so it is always attempted first.
  const vendorCreated = graph.created.find(c => c.key === 'vendor')
  assert.ok(vendorCreated, 'vendor should still be recorded in created so teardown can delete it')

  const vendorFailure = graph.failed.find(f => f.key === 'vendor')
  assert.ok(vendorFailure, 'vendor should also be recorded in failed')
  assert.match(String(vendorFailure.body), /cannot be recovered by tag/)

  // Its id must not be usable by dependents — buildFixtures must not have
  // published it into `ids`, or a downstream step could silently build on
  // an unrecoverable row.
  assert.equal(graph.ids.vendor, undefined)
})

test('buildFixtures treats singleton PUT steps as healthy without an id, and DOES track them in created for teardown', async () => {
  // A singleton upsert (vehicle/trailer/axle-config) returns a 2xx body with
  // no `id` field at all. buildFixtures must not treat that as a failure,
  // must still publish the step into `ids` for dependents (as `true`, not an
  // id), and — LEAK 1 fix — must still add a `created` row for it (marked
  // `singleton: true`, `id: null`) so teardownFixtures can find and delete
  // it. A generic stub that echoes every request body back (so the run tag
  // always round-trips) lets the real FIXTURE_PLAN run end to end.
  let nextId = 1
  const stubClient = {
    async request(method, _path, opts) {
      if (method === 'POST') return { status: 201, body: { id: nextId++, ...opts.body } }
      if (method === 'PUT') return { status: 200, body: { ...opts.body } }
      throw new Error(`unexpected ${method} call`)
    },
  }

  const graph = await buildFixtures(stubClient, 'singleton-stub')

  for (const key of ['vehicle', 'trailer', 'axleConfig']) {
    assert.equal(graph.ids[key], true, `${key} should be marked built via its singleton marker`)
    assert.ok(!graph.failed.some(f => f.key === key), `${key} must not be recorded as failed`)
    const row = graph.created.find(c => c.key === key)
    assert.ok(row, `${key} must be tracked in created so teardown can find it`)
    assert.equal(row.singleton, true)
    assert.equal(row.id, null)
  }

  // The three singletons all depend on `asset`, which is created earlier —
  // topoSort must place them after it in `created` so the reverse teardown
  // walk removes them BEFORE the asset they're attached to.
  const assetIndex = graph.created.findIndex(c => c.key === 'asset')
  for (const key of ['vehicle', 'trailer', 'axleConfig']) {
    const idx = graph.created.findIndex(c => c.key === key)
    assert.ok(idx > assetIndex, `${key} must be created after asset`)
  }

  const serviceTaskPart = graph.created.find(c => c.key === 'serviceTaskPart')
  assert.ok(serviceTaskPart, 'serviceTaskPart should be a normal created row with an id')
  assert.notEqual(serviceTaskPart.id, null)
  assert.ok(!serviceTaskPart.singleton)
})

// MINOR 7 regression: a non-singleton 2xx create with no `id` in the
// response is not a validation failure — a row may exist in the database
// right now with no way to address it for teardown. It must never be
// silent: it lands in `failed` (never `created`, since there is no id to
// delete by) with a message that says in plain words a row may still exist.
test('buildFixtures records a 2xx create with no id as a loud, explicit failure, not silence', async () => {
  const stubClient = {
    async request(method, _path, _opts) {
      if (method !== 'POST') throw new Error(`unexpected ${method} call`)
      // 2xx, but the body carries no `id` at all — the shape MINOR 7 covers.
      return { status: 201, body: { note: 'created but no id echoed back' } }
    },
  }

  const graph = await buildFixtures(stubClient, 'no-id-stub')

  const vendorCreated = graph.created.find(c => c.key === 'vendor')
  assert.equal(vendorCreated, undefined, 'a row with no id cannot be tracked in created — there is nothing to delete by')

  const vendorFailure = graph.failed.find(f => f.key === 'vendor')
  assert.ok(vendorFailure, 'vendor must still be recorded, in failed')
  assert.match(String(vendorFailure.body), /created.*201/i)
  assert.match(String(vendorFailure.body), /may still exist/i)
  assert.match(String(vendorFailure.body), /no-id-stub/)

  // Nothing downstream may build on this row either.
  assert.equal(graph.ids.vendor, undefined)
})

test('the vehicle/trailer/axle-config/serviceTaskPart steps are wired as expected', () => {
  const byKey = new Map(FIXTURE_PLAN.map(s => [s.key, s]))

  for (const key of ['vehicle', 'trailer', 'axleConfig']) {
    const step = byKey.get(key)
    assert.ok(step, `expected a FIXTURE_PLAN step for ${key}`)
    assert.equal(step.method, 'PUT')
    assert.equal(step.singleton, true)
  }

  const serviceTaskPart = byKey.get('serviceTaskPart')
  assert.ok(serviceTaskPart)
  assert.equal(serviceTaskPart.method ?? 'POST', 'POST')
  assert.ok(!serviceTaskPart.singleton)
  assert.ok(UNTAGGABLE.has('serviceTaskPart'))
})

// LEAK 1 regression: teardownFixtures must delete a singleton row at its own
// path with NO id appended — DELETE /api/v1/assets/43/vehicle, never
// .../vehicle/undefined or .../vehicle/null.
test('teardown deletes a singleton row at its own path, with no id suffix', async () => {
  const calls = []
  const stubClient = {
    async request(method, path) {
      calls.push({ method, path })
      if (method === 'DELETE') return { status: 204, body: null }
      throw new Error(`stub client received unexpected call: ${method} ${path}`)
    },
  }

  const graph = {
    runId: 'stub', tag: 'ZZ-TEST-99', ids: {},
    created: [
      { key: 'asset', path: '/api/v1/assets', id: 43 },
      { key: 'vehicle', path: '/api/v1/assets/43/vehicle', id: null, singleton: true },
    ],
    failed: [],
  }

  const result = await teardownFixtures(stubClient, graph)

  const deleteCalls = calls.filter(c => c.method === 'DELETE')
  assert.equal(deleteCalls.length, 2)
  // Reverse walk: the singleton (created after asset) is deleted first.
  assert.equal(deleteCalls[0].path, '/api/v1/assets/43/vehicle')
  assert.equal(deleteCalls[1].path, '/api/v1/assets/43')
  // No id, undefined, or null ever appended to the singleton's own path.
  assert.ok(!deleteCalls.some(c => /\/vehicle\/(undefined|null|\d)/.test(c.path)))

  assert.deepEqual(result.deleted.sort(), ['asset#43', 'vehicle'])
})

test('every path placeholder in FIXTURE_PLAN names a declared dependency', () => {
  const PLACEHOLDER = /\{(\w+)\}/g
  for (const step of FIXTURE_PLAN) {
    for (const match of step.path.matchAll(PLACEHOLDER)) {
      const placeholder = match[1]
      assert.ok(
        step.dependsOn.includes(placeholder),
        `${step.key} path '${step.path}' references {${placeholder}}, but dependsOn is [${step.dependsOn.join(', ')}]`
      )
    }
  }
})
