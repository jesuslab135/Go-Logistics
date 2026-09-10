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
