export const ISOLATION_TARGETS = [
  { key: 'vendor', path: '/api/v1/vendors' },
  { key: 'asset', path: '/api/v1/assets' },
  { key: 'part', path: '/api/v1/parts' },
  { key: 'workOrder', path: '/api/v1/work-orders' },
  { key: 'purchaseOrder', path: '/api/v1/purchase-orders' },
  { key: 'issue', path: '/api/v1/issues' },
  { key: 'serviceEntry', path: '/api/v1/service-entries' },
  { key: 'employee', path: '/api/v1/employees' },
  { key: 'journalEntry', path: '/api/v1/inventory-journal-entries' },
  { key: 'tire', path: '/api/v1/tires' },
]

export function verdictFor(status) {
  if (status === 403 || status === 404) {
    return { ok: true, severity: 'info' }
  }
  if (status >= 200 && status < 300) {
    return { ok: false, severity: 'high' }
  }
  return { ok: false, severity: 'medium' }
}

export async function createSecondTenant(client, tag) {
  const res = await client.request('POST', '/api/v1/companies', {
    body: { name: `${tag} Isolation Co`, tax_id: `${tag}-TAXID` },
    opKey: 'POST /api/v1/companies',
  })
  // Never throw: the client was built never to throw, and an exception here
  // would escape runIsolationSuite, kill the whole run, and skip teardown —
  // leaving every fixture row orphaned in a live production database. A
  // failure to create the throwaway tenant is itself a finding, not a crash.
  if (res.status < 200 || res.status >= 300 || res.body?.id == null) {
    return { companyId: null, response: res }
  }
  return { companyId: res.body.id, response: res }
}

async function switchTo(client, companyId) {
  const res = await client.request('POST', '/auth/switch-company', {
    body: { company_id: companyId },
    opKey: 'POST /auth/switch-company',
  })
  if (res.status >= 200 && res.status < 300 && res.body?.access_token) {
    client.setToken(res.body.access_token)
    if (res.body.refresh_token) client.refreshToken = res.body.refresh_token
  }
  return res
}

export async function runIsolationSuite(client, graph, homeCompanyId) {
  const results = []
  const { companyId, response } = await createSecondTenant(client, graph.tag)

  if (companyId == null) {
    // No tenant was created, so there is nothing to probe from and nothing
    // to switch back out of or delete. Record the failure and stop here.
    results.push({
      check: 'tenant-isolation',
      opKey: 'POST /api/v1/companies',
      ok: false,
      expected: '2xx creating the throwaway tenant',
      actual: String(response.status),
      severity: 'high',
      evidence: response.body,
    })
    return results
  }

  // Everything from here through the throwaway company's own DELETE runs
  // inside try/finally. Without this, a throw anywhere in the probes below
  // (the original bug: `(vendors.body?.data ?? []).some(...)` blows up if
  // `data` comes back as a non-array object) skips the DELETE entirely and
  // orphans a whole tenant in a live production database — the company is
  // never added to `graph.created` either, so run.mjs's own crash-path
  // teardown cannot reach it. This is the same shape as the two teardown
  // leaks fixed earlier in fixtures.mjs/workflows.mjs: whatever creates a
  // row owns tearing it down, in a finally, regardless of what throws
  // in between.
  try {
    const switched = await switchTo(client, companyId)
    results.push({
      check: 'tenant-isolation',
      opKey: 'POST /auth/switch-company',
      ok: switched.status >= 200 && switched.status < 300,
      expected: '2xx returning a token scoped to the new company',
      actual: String(switched.status),
      severity: switched.status >= 200 && switched.status < 300 ? 'info' : 'high',
      evidence: { companyId },
    })

    if (switched.status >= 200 && switched.status < 300) {
      for (const target of ISOLATION_TARGETS) {
        const id = graph.ids[target.key]
        if (!id) continue
        const res = await client.request('GET', `${target.path}/${id}`, {
          opKey: `GET ${target.path}/{id} (cross-tenant)`,
        })
        const v = verdictFor(res.status)
        results.push({
          check: 'tenant-isolation',
          opKey: `GET ${target.path}/{id}`,
          ok: v.ok,
          expected: "403 or 404 reading another company's row",
          actual: String(res.status),
          severity: v.severity,
          evidence: { targetKey: target.key, id, body: res.body },
        })
      }

      // A list in the new tenant must be empty of company-1 rows. Guarded
      // with Array.isArray, same as every other place in this codebase that
      // reads a list envelope's `data` field: an unexpected response shape
      // (an object instead of an array, say) must become a reported
      // finding, not a thrown TypeError that skips the cleanup below.
      const vendors = await client.request('GET', '/api/v1/vendors', {
        query: { limit: 200 }, opKey: 'GET /api/v1/vendors (cross-tenant list)',
      })
      const vendorData = vendors.body?.data
      const leaked = Array.isArray(vendorData) && vendorData.some(r => r.id === graph.ids.vendor)
      results.push({
        check: 'tenant-isolation',
        opKey: 'GET /api/v1/vendors (list)',
        ok: Array.isArray(vendorData) ? !leaked : false,
        expected: "the other tenant's vendor is absent from this company's list",
        actual: !Array.isArray(vendorData)
          ? 'data is not an array'
          : (leaked ? "the other tenant's vendor is listed" : 'absent'),
        severity: !Array.isArray(vendorData) ? 'medium' : (leaked ? 'high' : 'info'),
        evidence: { total: vendors.body?.total },
      })
    }
  } finally {
    // Return to the home tenant so teardown deletes in the right company.
    // This result MUST be checked, not discarded: `isolation` runs
    // immediately before `teardown` on this same client, and
    // teardownFixtures treats a 404 on its verification GET as "the row is
    // genuinely gone" — a check that is not tenant-aware. If this
    // switch-back silently failed, the client stays scoped to the
    // throwaway tenant, every fixture DELETE in teardown 404s against the
    // WRONG company, and the report would confidently claim zero residue
    // while every fixture row is still live in production. That is the
    // worst failure mode this tool has: a clean bill of health that is
    // false. `graph.tenantSwitchBackFailed` is the signal run.mjs checks
    // before running teardown at all, in both the normal phase flow and
    // the crash-path finally block — it is set here, before any exception
    // from the probes above finishes propagating, so it survives even the
    // throw path.
    const switchedHome = await switchTo(client, homeCompanyId)
    const switchedHomeOk = switchedHome.status >= 200 && switchedHome.status < 300
    graph.tenantSwitchBackFailed = !switchedHomeOk
    results.push({
      check: 'tenant-isolation',
      opKey: 'POST /auth/switch-company (return to home)',
      ok: switchedHomeOk,
      expected: '2xx returning a token scoped back to the home company before teardown runs',
      actual: String(switchedHome.status),
      severity: switchedHomeOk ? 'info' : 'high',
      evidence: { homeCompanyId },
    })

    if (switchedHomeOk) {
      const cleanup = await client.request('DELETE', `/api/v1/companies/${companyId}`, {
        opKey: 'DELETE /api/v1/companies/{id} (isolation teardown)',
      })
      results.push({
        check: 'tenant-isolation',
        opKey: 'DELETE /api/v1/companies/{id}',
        ok: cleanup.status >= 200 && cleanup.status < 300,
        expected: '2xx removing the throwaway tenant',
        actual: String(cleanup.status),
        severity: cleanup.status >= 200 && cleanup.status < 300 ? 'info' : 'medium',
        evidence: { companyId, body: cleanup.body },
      })
    } else {
      // Do not attempt the delete: we cannot be sure which company it
      // would run against. Record the throwaway tenant as still live
      // rather than silently dropping it.
      results.push({
        check: 'tenant-isolation',
        opKey: 'DELETE /api/v1/companies/{id}',
        ok: false,
        expected: 'the throwaway tenant to be removed once the client is confirmed back in the home company',
        actual: 'skipped: switch-back to the home company did not succeed, so the client identity is unknown',
        severity: 'high',
        evidence: { companyId },
      })
    }
  }

  return results
}
