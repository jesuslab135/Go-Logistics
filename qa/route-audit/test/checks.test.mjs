import { test } from 'node:test'
import assert from 'node:assert/strict'
import {
  fillPath, MISSING_ID, classify, PUBLIC_ROUTES,
  checkUnauthenticated, checkNotFound, checkListContract,
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

test('checkUnauthenticated returns a passing info-severity result for a public route, still sending the probe', async () => {
  const client = makeStubClient({ status: 400, body: { error: 'invalid request body' } })
  const op = { opKey: 'POST /auth/login', method: 'POST', path: '/auth/login', requestSchema: { type: 'object' } }
  const out = await checkUnauthenticated(client, op)
  // The probe is still sent — its response is useful evidence for the
  // report — but the operation is recorded as a passing check, not skipped,
  // so it does not produce a second coverage row alongside the verb
  // dispatch's own entry.
  assert.equal(client.calls.length, 1)
  assert.deepEqual(out, {
    check: 'unauthenticated',
    opKey: 'POST /auth/login',
    ok: true,
    expected: 'no authentication required — this route is public by design',
    actual: '400 (not gated)',
    severity: 'info',
    evidence: { error: 'invalid request body' },
  })
})

test('checkUnauthenticated marks every PUBLIC_ROUTES opKey ok:true at info severity', async () => {
  for (const opKey of PUBLIC_ROUTES) {
    const [method, path] = opKey.split(' ')
    const client = makeStubClient({ status: method === 'GET' ? 200 : 400, body: {} })
    const out = await checkUnauthenticated(client, { opKey, method, path, requestSchema: method === 'GET' ? null : { type: 'object' } })
    assert.equal(out.ok, true, `${opKey} should pass`)
    assert.equal(out.severity, 'info', `${opKey} should be info severity`)
  }
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

// ---------------------------------------------------------------------------
// MINOR 5: checkListContract substituted MISSING_ID into a nested list's
// parent {id} unconditionally, so /admin/companies/{id}/roles and
// /admin/companies/{id}/work-order-statuses — whose backend correctly 404s
// for a nonexistent company — were reported as failing list-contract checks.
// The fix: skip when the caller (sweep.mjs, via idsForOperation) could not
// resolve the op's placeholders to any real fixture row at all (`ids ===
// null`). A route whose parent CAN be resolved (e.g. /assets/{id}/...) must
// still be checked as before.
// ---------------------------------------------------------------------------

function makeListClient(response) {
  const calls = []
  return {
    calls,
    async request(method, path, opts = {}) {
      calls.push({ method, path, opts })
      return response
    },
  }
}

test('checkListContract skips a nested list whose parent id could not be resolved to a real fixture row', async () => {
  const client = makeListClient({ status: 404, body: { error: { code: 'not_found' } } })
  const op = {
    opKey: 'GET /api/v1/admin/companies/{id}/roles',
    method: 'GET',
    path: '/api/v1/admin/companies/{id}/roles',
    params: [{ name: 'limit' }],
  }
  const out = await checkListContract(client, op, null)
  assert.equal(out, null, 'a 404 for an unresolvable parent must not be reported as a list-contract failure')
  assert.equal(client.calls.length, 0, 'must not even send the probe once the parent id is known to be unresolvable')
})

test('checkListContract still runs a nested list whose parent id resolved to a real fixture row', async () => {
  const client = makeListClient({ status: 200, body: { data: [], total: 0, limit: 1, offset: 0, has_next: false } })
  const op = {
    opKey: 'GET /api/v1/assets/{id}/fuel-entries',
    method: 'GET',
    path: '/api/v1/assets/{id}/fuel-entries',
    params: [{ name: 'limit' }],
  }
  const out = await checkListContract(client, op, { id: 42 })
  assert.ok(out, 'a resolvable parent must still be probed')
  assert.equal(out.ok, true)
  assert.equal(client.calls.length, 1)
})

test('checkListContract still runs a top-level list with no parent placeholder at all', async () => {
  const client = makeListClient({ status: 200, body: { data: [], total: 0, limit: 1, offset: 0, has_next: false } })
  const op = {
    opKey: 'GET /api/v1/vendors',
    method: 'GET',
    path: '/api/v1/vendors',
    params: [{ name: 'limit' }],
  }
  const out = await checkListContract(client, op, {})
  assert.ok(out)
  assert.equal(out.ok, true)
})

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
