import { test } from 'node:test'
import assert from 'node:assert/strict'
import { normalisePath, matchOperation, crossReference, unusedOperations } from '../lib/crossref.mjs'

test('normalisePath strips the origin', () => {
  assert.equal(
    normalisePath('https://go-logistics.jesuslab135.com/api/v1/assets'),
    '/api/v1/assets'
  )
})

test('normalisePath strips the query string', () => {
  assert.equal(normalisePath('/api/v1/assets?limit=25&offset=0'), '/api/v1/assets')
})

test('normalisePath replaces the first numeric segment with {id}', () => {
  assert.equal(normalisePath('/api/v1/assets/12'), '/api/v1/assets/{id}')
})

test('normalisePath replaces the second numeric segment with {child_id}', () => {
  assert.equal(
    normalisePath('/api/v1/work-orders/12/line-items/34'),
    '/api/v1/work-orders/{id}/line-items/{child_id}'
  )
})

test('normalisePath leaves non-numeric segments alone', () => {
  assert.equal(
    normalisePath('/api/v1/purchase-orders/12/approve'),
    '/api/v1/purchase-orders/{id}/approve'
  )
})

test('matchOperation finds the operation for a real frontend URL', () => {
  const ops = [
    { opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' },
    { opKey: 'GET /api/v1/assets/{id}', method: 'GET', path: '/api/v1/assets/{id}' },
  ]
  assert.equal(matchOperation(ops, 'GET', '/api/v1/assets/9').opKey, 'GET /api/v1/assets/{id}')
})

test('matchOperation returns null for an endpoint the backend does not define', () => {
  const ops = [{ opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' }]
  assert.equal(matchOperation(ops, 'GET', '/api/v1/ghosts'), null)
})

test('crossReference flags a frontend call to an undefined endpoint', () => {
  const ops = [{ opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' }]
  const captures = [{
    route: '/app/vehicles',
    rendered: 'error',
    consoleErrors: [],
    requests: [{ method: 'GET', url: '/api/v1/ghosts', status: 404 }],
  }]
  const findings = crossReference({ definitions: {} }, ops, captures)
  assert.equal(findings.length, 1)
  assert.equal(findings[0].layer, 'frontend')
  assert.match(findings[0].summary, /no such backend operation/i)
})

test('crossReference flags a failing request on a rendered screen', () => {
  const ops = [{ opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' }]
  const captures = [{
    route: '/app/vehicles',
    rendered: 'error',
    consoleErrors: [],
    requests: [{ method: 'GET', url: '/api/v1/assets', status: 500 }],
  }]
  const findings = crossReference({ definitions: {} }, ops, captures)
  assert.equal(findings[0].layer, 'backend')
  assert.equal(findings[0].severity, 'high')
})

test('crossReference flags a blank render even when every request succeeded', () => {
  const ops = [{ opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' }]
  const captures = [{
    route: '/app/vehicles',
    rendered: 'blank',
    consoleErrors: ['TypeError: x is undefined'],
    requests: [{ method: 'GET', url: '/api/v1/assets', status: 200 }],
  }]
  const findings = crossReference({ definitions: {} }, ops, captures)
  assert.equal(findings.length, 1)
  assert.equal(findings[0].layer, 'frontend')
})

test('unusedOperations lists backend operations no screen calls', () => {
  const ops = [
    { opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' },
    { opKey: 'GET /api/v1/tires', method: 'GET', path: '/api/v1/tires' },
  ]
  const captures = [{ route: '/app/vehicles', requests: [{ method: 'GET', url: '/api/v1/assets', status: 200 }] }]
  assert.deepEqual(unusedOperations(ops, captures), ['GET /api/v1/tires'])
})
