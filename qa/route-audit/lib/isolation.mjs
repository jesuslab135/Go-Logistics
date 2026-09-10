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
    body: { name: `${tag} Isolation Co` },
    opKey: 'POST /api/v1/companies',
  })
  if (res.status < 200 || res.status >= 300 || res.body?.id == null) {
    throw new Error(`could not create the second tenant: ${res.status} ${JSON.stringify(res.body)}`)
  }
  return { companyId: res.body.id }
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
  const { companyId } = await createSecondTenant(client, graph.tag)

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

    // A list in the new tenant must be empty of company-1 rows.
    const vendors = await client.request('GET', '/api/v1/vendors', {
      query: { limit: 200 }, opKey: 'GET /api/v1/vendors (cross-tenant list)',
    })
    const leaked = (vendors.body?.data ?? []).some(r => r.id === graph.ids.vendor)
    results.push({
      check: 'tenant-isolation',
      opKey: 'GET /api/v1/vendors (list)',
      ok: !leaked,
      expected: "the other tenant's vendor is absent from this company's list",
      actual: leaked ? "the other tenant's vendor is listed" : 'absent',
      severity: leaked ? 'high' : 'info',
      evidence: { total: vendors.body?.total },
    })
  }

  // Return to the home tenant so teardown deletes in the right company.
  await switchTo(client, homeCompanyId)

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

  return results
}
