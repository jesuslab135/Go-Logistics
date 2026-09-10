import { test } from 'node:test'
import assert from 'node:assert/strict'
import { fillPath, MISSING_ID, classify } from '../lib/checks.mjs'

test('fillPath substitutes a single id', () => {
  assert.equal(fillPath('/api/v1/vendors/{id}', { id: 7 }), '/api/v1/vendors/7')
})

test('fillPath substitutes parent and child ids', () => {
  assert.equal(
    fillPath('/api/v1/work-orders/{id}/line-items/{child_id}', { id: 3, child_id: 9 }),
    '/api/v1/work-orders/3/line-items/9'
  )
})

test('fillPath leaves an unknown placeholder in place so it fails loudly', () => {
  assert.equal(fillPath('/api/v1/x/{other}', { id: 1 }), '/api/v1/x/{other}')
})

test('MISSING_ID is far outside any real sequence', () => {
  assert.ok(MISSING_ID > 100000000)
})

test('classify marks an auth gap as high severity', () => {
  assert.equal(classify('unauthenticated', false), 'high')
  assert.equal(classify('tenant-isolation', false), 'high')
})

test('classify marks an undocumented-field drift as low severity', () => {
  assert.equal(classify('conformance', false), 'low')
})

test('classify returns info for a passing check', () => {
  assert.equal(classify('unauthenticated', true), 'info')
})
