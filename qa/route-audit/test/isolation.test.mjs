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
