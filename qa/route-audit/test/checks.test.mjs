import { test } from 'node:test'
import assert from 'node:assert/strict'
import {
  fillPath, MISSING_ID, classify, PUBLIC_ROUTES,
  checkUnauthenticated, checkNotFound,
} from '../lib/checks.mjs'

function makeStubClient(response) {
  const calls = []
  return {
    calls,
    async request(method, path, opts = {}) {
      calls.push({ method, path, opts })
      return response
    },
  }
}

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

// ---------------------------------------------------------------------------
// DEFECT 1: public routes are exempt from the unauthenticated-probe check.
// ---------------------------------------------------------------------------

test('PUBLIC_ROUTES names the three auth endpoints and healthz, not switch-company', () => {
  assert.deepEqual(
    [...PUBLIC_ROUTES].sort(),
    ['GET /healthz', 'POST /auth/login', 'POST /auth/logout', 'POST /auth/refresh']
  )
  assert.ok(!PUBLIC_ROUTES.has('POST /auth/switch-company'))
})

test('checkUnauthenticated returns null for a public route and sends no probe', async () => {
  const client = makeStubClient({ status: 400, body: { error: 'invalid request body' } })
  const op = { opKey: 'POST /auth/login', method: 'POST', path: '/auth/login', requestSchema: { type: 'object' } }
  const out = await checkUnauthenticated(client, op)
  assert.equal(out, null)
  assert.deepEqual(client.calls, [])
})

test('checkUnauthenticated still expects 401 for switch-company, a non-public route', async () => {
  const client = makeStubClient({ status: 401, body: { error: 'unauthorized' } })
  const op = { opKey: 'POST /auth/switch-company', method: 'POST', path: '/auth/switch-company', requestSchema: { type: 'object' } }
  const out = await checkUnauthenticated(client, op)
  assert.ok(out)
  assert.equal(out.ok, true)
})

// ---------------------------------------------------------------------------
// DEFECT 2: checkNotFound must probe every named path placeholder, not just
// {id}/{child_id} — m2m link routes use {employee_id}, {issue_id}, {fault_id}.
// ---------------------------------------------------------------------------

test('checkNotFound substitutes a named link param, not just id/child_id', async () => {
  const client = makeStubClient({ status: 404, body: { error: 'not found' } })
  const op = {
    opKey: 'DELETE /api/v1/issues/{id}/assigned-to/{employee_id}',
    method: 'DELETE',
    path: '/api/v1/issues/{id}/assigned-to/{employee_id}',
    requestSchema: null,
  }
  const out = await checkNotFound(client, op)
  assert.equal(client.calls.length, 1)
  assert.equal(client.calls[0].path, `/api/v1/issues/${MISSING_ID}/assigned-to/${MISSING_ID}`)
  assert.equal(out.ok, true)
})
