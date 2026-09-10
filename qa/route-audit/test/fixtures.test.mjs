// qa/route-audit/test/fixtures.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { topoSort, FIXTURE_PLAN } from '../lib/fixtures.mjs'

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

test('every fixture body carries the run tag so teardown can find it', () => {
  const ids = Object.fromEntries(FIXTURE_PLAN.map(s => [s.key, 1]))
  for (const step of FIXTURE_PLAN) {
    const body = step.body(ids, 'ZZ-TEST-42')
    const serialised = JSON.stringify(body)
    assert.ok(
      serialised.includes('ZZ-TEST-42'),
      `${step.key} body does not carry the run tag: ${serialised}`
    )
  }
})
