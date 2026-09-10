import { test } from 'node:test'
import assert from 'node:assert/strict'
import { ISOLATION_TARGETS, verdictFor, runIsolationSuite } from '../lib/isolation.mjs'
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

// ---------------------------------------------------------------------------
// The most serious failure mode in this suite is an exception that escapes
// runIsolationSuite: it kills the whole audit script mid-run and skips
// teardown, orphaning every fixture row in a live production database. This
// proves company creation failing (e.g. a validation error) is recorded as
// a finding, not thrown, and that the suite gives up cleanly without
// attempting to switch into or delete a tenant that was never created.
// ---------------------------------------------------------------------------

function makeStubClient(handlers, fallback = { status: 200, body: {}, headers: {}, durationMs: 1 }) {
  const calls = []
  return {
    calls,
    setToken() {},
    refreshToken: null,
    async request(method, path, opts = {}) {
      calls.push({ method, path, opts })
      const handler = handlers[`${method} ${path}`]
      if (typeof handler === 'function') return handler(opts)
      return handler ?? fallback
    },
  }
}

test('runIsolationSuite resolves (never throws) and records one high-severity failure when tenant creation fails, without switching companies', async () => {
  const client = makeStubClient({
    'POST /api/v1/companies': {
      status: 422,
      body: { error: { code: 'validation_failed', message: 'validation failed', details: { tax_id: 'this field is required' } } },
      headers: {},
      durationMs: 1,
    },
  })
  const graph = { tag: 'ZZ-TEST-stub', ids: {} }

  const results = await runIsolationSuite(client, graph, 1)

  assert.deepEqual(results.length, 1)
  assert.equal(results[0].check, 'tenant-isolation')
  assert.equal(results[0].opKey, 'POST /api/v1/companies')
  assert.equal(results[0].ok, false)
  assert.equal(results[0].severity, 'high')

  const switchCalls = client.calls.filter(c => c.path === '/auth/switch-company')
  assert.deepEqual(switchCalls, [])
})

// ---------------------------------------------------------------------------
// IMPORTANT 1: the throwaway company created at the top of the suite was
// deleted with no try/finally around it — a throw anywhere between its
// creation and its delete (the real one found on the live run: `.some(...)`
// on a non-array `data`) skipped the delete and orphaned a whole tenant.
// ---------------------------------------------------------------------------

test('a throw during the cross-tenant probes still switches back to home and deletes the throwaway tenant, then propagates', async () => {
  const client = makeStubClient({
    'POST /api/v1/companies': { status: 201, body: { id: 555 }, headers: {}, durationMs: 1 },
    'POST /auth/switch-company': (opts) => (
      { status: 200, body: { access_token: `tok-${opts.body.company_id}` }, headers: {}, durationMs: 1 }
    ),
    'GET /api/v1/vendors': () => { throw new Error('boom: unexpected shape') },
    'DELETE /api/v1/companies/555': { status: 204, body: null, headers: {}, durationMs: 1 },
  })
  // No ids set, so the per-target cross-tenant loop finds nothing to probe
  // and the only thing left to throw is the vendors list GET below it.
  const graph = { tag: 'ZZ-TEST-stub', ids: {} }

  await assert.rejects(() => runIsolationSuite(client, graph, 1), /boom/)

  const switchedTo = client.calls
    .filter(c => c.path === '/auth/switch-company')
    .map(c => c.opts.body.company_id)
  assert.deepEqual(switchedTo, [555, 1], 'must switch into the throwaway tenant, then back home, even though a probe threw')

  const deleteCall = client.calls.find(c => c.method === 'DELETE')
  assert.ok(deleteCall, 'the throwaway company must still be deleted even though a probe threw')
  assert.equal(deleteCall.path, '/api/v1/companies/555')

  assert.equal(graph.tenantSwitchBackFailed, false, 'the switch-back itself succeeded in this scenario')
})

test('a non-array vendors list is a reported failure, not a thrown TypeError, and cleanup still runs', async () => {
  const client = makeStubClient({
    'POST /api/v1/companies': { status: 201, body: { id: 555 }, headers: {}, durationMs: 1 },
    'POST /auth/switch-company': (opts) => (
      { status: 200, body: { access_token: `tok-${opts.body.company_id}` }, headers: {}, durationMs: 1 }
    ),
    // The exact shape that used to throw: `data` is a non-array object.
    'GET /api/v1/vendors': { status: 200, body: { data: { not: 'an array' }, total: 1 }, headers: {}, durationMs: 1 },
    'DELETE /api/v1/companies/555': { status: 204, body: null, headers: {}, durationMs: 1 },
  })
  const graph = { tag: 'ZZ-TEST-stub', ids: {} }

  const results = await runIsolationSuite(client, graph, 1)

  const listCheck = results.find(r => r.opKey === 'GET /api/v1/vendors (list)')
  assert.ok(listCheck, 'expected a recorded result for the cross-tenant vendors list check')
  assert.equal(listCheck.ok, false)
  assert.equal(listCheck.severity, 'medium')
  assert.match(listCheck.actual, /not an array/)

  // Cleanup must still have completed normally.
  const deleteCall = client.calls.find(c => c.method === 'DELETE')
  assert.ok(deleteCall, 'the throwaway company must be deleted')
  assert.equal(graph.tenantSwitchBackFailed, false)
})

// ---------------------------------------------------------------------------
// IMPORTANT 2: the switch back to the home company at the end of the suite
// was fired and its result discarded. If it silently fails, the client stays
// scoped to the throwaway tenant and every fixture DELETE in the next phase
// (teardown) 404s against the WRONG company — teardownFixtures reads that as
// "genuinely gone" and the report claims a clean database while every row is
// still live. This proves a failed switch-back produces a failing,
// high-severity CheckResult, does NOT attempt the company delete (whose
// outcome would be meaningless against an unknown-identity client), and sets
// `graph.tenantSwitchBackFailed` — the signal run.mjs checks before running
// teardown at all — so a false clean-teardown report is impossible.
// ---------------------------------------------------------------------------

test('a failed switch-back to home is a high-severity failing result, skips the company delete, and flags the graph so teardown cannot run', async () => {
  const client = makeStubClient({
    'POST /api/v1/companies': { status: 201, body: { id: 555 }, headers: {}, durationMs: 1 },
    'POST /auth/switch-company': (opts) => {
      // Succeed switching INTO the throwaway tenant; fail switching back.
      if (opts.body.company_id === 555) {
        return { status: 200, body: { access_token: 'tok-555' }, headers: {}, durationMs: 1 }
      }
      return { status: 500, body: { error: { code: 'internal', message: 'token service unavailable' } }, headers: {}, durationMs: 1 }
    },
    'GET /api/v1/vendors': { status: 200, body: { data: [], total: 0 }, headers: {}, durationMs: 1 },
  })
  const graph = { tag: 'ZZ-TEST-stub', ids: {} }

  const results = await runIsolationSuite(client, graph, 1)

  const switchBack = results.find(r => r.opKey === 'POST /auth/switch-company (return to home)')
  assert.ok(switchBack, 'expected a recorded result for the switch-back to home')
  assert.equal(switchBack.ok, false)
  assert.equal(switchBack.severity, 'high')
  assert.equal(switchBack.actual, '500')

  // The delete must never be attempted against a client of unknown identity.
  const deleteCall = client.calls.find(c => c.method === 'DELETE')
  assert.equal(deleteCall, undefined, 'must not attempt the company delete when the client identity is unknown')

  const deleteResult = results.find(r => r.opKey === 'DELETE /api/v1/companies/{id}')
  assert.ok(deleteResult, 'expected a recorded (failing) result explaining the delete was skipped')
  assert.equal(deleteResult.ok, false)
  assert.equal(deleteResult.severity, 'high')
  assert.match(deleteResult.actual, /skipped/i)

  // The signal a caller (run.mjs) checks before running teardown at all.
  assert.equal(graph.tenantSwitchBackFailed, true)
})

// Fail-closed regression: switchTo() calls client.request, which is
// documented elsewhere as "never throws" but CAN in reality — e.g. the
// switch-back 401s and the subsequent token refresh also fails. That throw
// happens inside the `finally` block itself, before the line that would
// normally clear `tenantSwitchBackFailed` on confirmed success ever runs.
// The flag must already be true by the time control reaches that point (set
// before the `try`, ahead of anything that can throw), or it would default
// to `undefined` — falsy — and both of run.mjs's guards would read that as
// "safe to run teardown" against a client of unknown tenant.
test('graph.tenantSwitchBackFailed stays true when the switch-back request itself throws, not just when it returns non-2xx', async () => {
  const client = makeStubClient({
    'POST /api/v1/companies': { status: 201, body: { id: 555 }, headers: {}, durationMs: 1 },
    'POST /auth/switch-company': (opts) => {
      if (opts.body.company_id === 555) {
        return { status: 200, body: { access_token: 'tok-555' }, headers: {}, durationMs: 1 }
      }
      // Simulates the 401-then-refresh-also-fails path: the request layer
      // throws instead of returning a status at all.
      throw new Error('switch-back 401d and the token refresh also failed')
    },
    'GET /api/v1/vendors': { status: 200, body: { data: [], total: 0 }, headers: {}, durationMs: 1 },
  })
  const graph = { tag: 'ZZ-TEST-stub', ids: {} }

  await assert.rejects(
    () => runIsolationSuite(client, graph, 1),
    /switch-back 401d/
  )

  // This is the exact case the fail-closed fix protects: the assignment
  // that would clear the flag never runs, because switchTo() threw before
  // reaching it. Without the pre-set `true`, this would read `undefined`.
  assert.equal(graph.tenantSwitchBackFailed, true)

  // The delete must never be attempted either, for the same reason as the
  // non-throwing failure case above.
  const deleteCall = client.calls.find(c => c.method === 'DELETE')
  assert.equal(deleteCall, undefined, 'must not attempt the company delete when the switch-back threw')
})
