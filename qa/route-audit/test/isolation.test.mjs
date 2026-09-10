import { test } from 'node:test'
import assert from 'node:assert/strict'
import { ISOLATION_TARGETS, verdictFor } from '../lib/isolation.mjs'
import { FIXTURE_PLAN } from '../lib/fixtures.mjs'

test('every isolation target names a real fixture key', () => {
  const keys = new Set(FIXTURE_PLAN.map(s => s.key))
  for (const t of ISOLATION_TARGETS) {
    assert.ok(keys.has(t.key), `${t.key} is not a fixture`)
  }
})

test('verdictFor accepts 403 and 404 as correct refusals', () => {
  assert.equal(verdictFor(403).ok, true)
  assert.equal(verdictFor(404).ok, true)
})

test('verdictFor treats a 200 as a leak at high severity', () => {
  const v = verdictFor(200)
  assert.equal(v.ok, false)
  assert.equal(v.severity, 'high')
})

test('verdictFor treats a 500 as a failure but not a leak', () => {
  const v = verdictFor(500)
  assert.equal(v.ok, false)
  assert.equal(v.severity, 'medium')
})
