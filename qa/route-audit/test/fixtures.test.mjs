// qa/route-audit/test/fixtures.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { topoSort, FIXTURE_PLAN, UNTAGGABLE } from '../lib/fixtures.mjs'

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
