// qa/route-audit/test/fixtures.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { topoSort, FIXTURE_PLAN, UNTAGGABLE, teardownFixtures } from '../lib/fixtures.mjs'

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
  // '2') and ISO-8601 timestamps (e.g. started_at). Those are format-
  // constrained by the API and cannot hold a human-readable tag no matter
  // what — they are not the "hole" this test guards against. A field that
  // holds arbitrary text (name, notes, description, label, terms, ...)
  // is the hole: it would mean the DTO grew a place for the tag and the
  // exemption should be removed. So this test fails on any string field
  // that is NOT a recognised constrained format, rather than on every
  // string field.
  const ISO_TIMESTAMP = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/
  const DECIMAL_STRING = /^\d+(\.\d+)?$/
  const ids = Object.fromEntries(FIXTURE_PLAN.map(s => [s.key, 1]))
  const byKey = new Map(FIXTURE_PLAN.map(s => [s.key, s]))
  for (const key of UNTAGGABLE) {
    const step = byKey.get(key)
    assert.ok(step, `UNTAGGABLE names unknown step ${key}`)
    const body = step.body(ids, 'ZZ-TEST-42')
    for (const [field, value] of Object.entries(body)) {
      if (typeof value !== 'string') continue
      assert.ok(
        ISO_TIMESTAMP.test(value) || DECIMAL_STRING.test(value),
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
  // location is in neither ARCHIVABLE nor REVERSIBLE -> DELETE refused -> failed
  assert.deepEqual(result.failed.map(f => f.key), ['location#11'])
  assert.deepEqual(result.deleted, [])

  const reverseCall = calls.find(c => c.path === '/api/v1/inventory-journal-entries/3/reverse')
  assert.ok(reverseCall, 'expected a POST to the reverse endpoint')
  assert.deepEqual(reverseCall.body, { notes: 'ZZ-TEST-99 teardown reversal' })

  const archiveCall = calls.find(c => c.path === '/api/v1/vendors/7/archive')
  assert.ok(archiveCall, 'expected a POST to the archive endpoint')

  // location must not have triggered a reverse or archive call at all.
  assert.ok(!calls.some(c => c.path.startsWith('/api/v1/locations/11/')))
})
