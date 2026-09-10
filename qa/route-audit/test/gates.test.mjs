import { test } from 'node:test'
import assert from 'node:assert/strict'
import { MODULE_PROBES, runGateSuite } from '../lib/gates.mjs'
import { ApiClient } from '../lib/client.mjs'

test('every module named by the permissions endpoint has a probe', () => {
  const modules = [
    'assets', 'company', 'employees', 'fuel', 'inspections', 'inventory',
    'issues', 'mileage_goals', 'parts', 'purchase_orders', 'roles', 'service',
    'tire_approvals', 'tires', 'vendors', 'warranties', 'work_orders',
  ]
  const covered = new Set(MODULE_PROBES.map(p => p.module))
  for (const m of modules) {
    assert.ok(covered.has(m), `no gate probe for module ${m}`)
  }
})

test('every probe targets a read verb, so a 403 cannot be confused with a write failure, except tire_approvals', () => {
  for (const p of MODULE_PROBES) {
    if (p.module === 'tire_approvals') continue
    assert.equal(p.probe.method, 'GET', `${p.module} probe should be a GET`)
  }
})

test('tire_approvals is pinned as the one POST probe, against the approve action, with a nonexistent id', () => {
  const p = MODULE_PROBES.find(m => m.module === 'tire_approvals')
  assert.ok(p, 'expected a tire_approvals probe')
  assert.equal(p.probe.method, 'POST')
  assert.ok(p.probe.path.includes('999999999'), 'expected the nonexistent-id placeholder value')
})

test('every probe path is a collection or a literal id, needing no fixture id placeholder', () => {
  for (const p of MODULE_PROBES) {
    assert.ok(!p.probe.path.includes('{'), `${p.module} probe path has a placeholder`)
  }
})

// ---------------------------------------------------------------------------
// runGateSuite logs in a *second*, internally-constructed ApiClient (the
// "limited identity" client) rather than reusing the client it is passed —
// see lib/gates.mjs. That internal instantiation is the only seam available
// to stub without touching the network, so this test patches
// ApiClient.prototype for its duration instead of passing a plain stub
// object positionally. No real fetch is ever made.
// ---------------------------------------------------------------------------

test('a probe returning 200 instead of 403 is recorded as a high-severity failure', async () => {
  const originalLogin = ApiClient.prototype.login
  const originalRequest = ApiClient.prototype.request
  const leakingProbe = MODULE_PROBES[0].probe

  ApiClient.prototype.login = async function stubLogin() {
    this.accessToken = 'stub-token'
    this.refreshToken = 'stub-refresh'
  }
  ApiClient.prototype.request = async function stubRequest(method, path) {
    // Every probe correctly refuses except the one under test, which leaks.
    if (method === leakingProbe.method && path === leakingProbe.path) {
      return { status: 200, body: { data: [], total: 0, has_next: false }, headers: {}, durationMs: 1 }
    }
    if (method === 'GET' && path === '/api/v1/me/permissions') {
      return { status: 200, body: { modules: [] }, headers: {}, durationMs: 1 }
    }
    return { status: 403, body: { error: { message: 'forbidden' } }, headers: {}, durationMs: 1 }
  }

  try {
    const client = { baseUrl: 'https://example.invalid', recorder: { record() {} } }
    const graph = { tag: 'ZZ-TEST-stub', ids: { employee: 1, role: 2 } }
    const limited = { email: 'zz-test-stub-employee@example.invalid', employeeId: 1, roleId: 2 }

    const results = await runGateSuite(client, graph, limited, 'irrelevant-password')

    const leaked = results.find(r => r.opKey === `${leakingProbe.method} ${leakingProbe.path}`)
    assert.ok(leaked, 'expected a result for the leaking probe')
    assert.equal(leaked.ok, false)
    assert.equal(leaked.severity, 'high')
  } finally {
    ApiClient.prototype.login = originalLogin
    ApiClient.prototype.request = originalRequest
  }
})
