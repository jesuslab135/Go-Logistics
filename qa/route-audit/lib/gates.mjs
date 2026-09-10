import { ApiClient } from './client.mjs'
import { classify } from './checks.mjs'

export const MODULE_PROBES = [
  { module: 'assets', probe: { method: 'GET', path: '/api/v1/assets' } },
  { module: 'company', probe: { method: 'GET', path: '/api/v1/companies' } },
  { module: 'employees', probe: { method: 'GET', path: '/api/v1/employees' } },
  { module: 'fuel', probe: { method: 'GET', path: '/api/v1/fuel-entries' } },
  { module: 'inspections', probe: { method: 'GET', path: '/api/v1/inspection-forms' } },
  { module: 'inventory', probe: { method: 'GET', path: '/api/v1/measurement-units' } },
  { module: 'issues', probe: { method: 'GET', path: '/api/v1/issues' } },
  { module: 'mileage_goals', probe: { method: 'GET', path: '/api/v1/weekly-mileage-goals' } },
  { module: 'parts', probe: { method: 'GET', path: '/api/v1/parts' } },
  { module: 'purchase_orders', probe: { method: 'GET', path: '/api/v1/purchase-orders' } },
  { module: 'roles', probe: { method: 'GET', path: '/api/v1/roles' } },
  { module: 'service', probe: { method: 'GET', path: '/api/v1/service-tasks' } },
  { module: 'tire_approvals', probe: { method: 'GET', path: '/api/v1/tire-assignment-requests' } },
  { module: 'tires', probe: { method: 'GET', path: '/api/v1/tires' } },
  { module: 'vendors', probe: { method: 'GET', path: '/api/v1/vendors' } },
  { module: 'warranties', probe: { method: 'GET', path: '/api/v1/warranties' } },
  { module: 'work_orders', probe: { method: 'GET', path: '/api/v1/work-orders' } },
]

// Routes that must refuse a non-platform-admin outright.
const PLATFORM_ADMIN_PROBES = [
  { method: 'GET', path: '/api/v1/admin/employees' },
  { method: 'GET', path: '/api/v1/admin/companies/1/roles' },
  { method: 'GET', path: '/api/v1/admin/companies/1/owner' },
  { method: 'GET', path: '/api/v1/admin/companies/1/work-order-statuses' },
]

export async function provisionLimitedIdentity(client, graph, password) {
  const employeeId = graph.ids.employee
  if (!employeeId) throw new Error('fixture employee was not created; gate suite cannot run')

  const res = await client.request('POST', `/api/v1/employees/${employeeId}/set-password`, {
    body: { password },
    opKey: 'POST /api/v1/employees/{id}/set-password',
  })
  if (res.status < 200 || res.status >= 300) {
    throw new Error(`set-password failed: ${res.status} ${JSON.stringify(res.body)}`)
  }

  return {
    email: `${graph.tag.toLowerCase()}-employee@example.invalid`,
    employeeId,
    roleId: graph.ids.role,
  }
}

function gateResult(check, opKey, ok, expected, actual, evidence) {
  return { check, opKey, ok, expected, actual, severity: classify(check, ok), evidence }
}

export async function runGateSuite(client, graph, limited, password) {
  const results = []

  const limitedClient = new ApiClient({ baseUrl: client.baseUrl, recorder: client.recorder })
  await limitedClient.login(limited.email, password)

  // The fixture role grants no modules, so every module probe must refuse.
  for (const { module, probe } of MODULE_PROBES) {
    const res = await limitedClient.request(probe.method, probe.path, {
      opKey: `${probe.method} ${probe.path} (module-gate:${module})`,
    })
    const ok = res.status === 403
    results.push(gateResult(
      'module-gate',
      `${probe.method} ${probe.path}`,
      ok,
      `403 for a role without the ${module} module`,
      String(res.status),
      { module, body: res.body }
    ))
  }

  // A tenant admin is not a platform operator. These must refuse too.
  for (const probe of PLATFORM_ADMIN_PROBES) {
    const res = await limitedClient.request(probe.method, probe.path, {
      opKey: `${probe.method} ${probe.path} (platform-admin-gate)`,
    })
    const ok = res.status === 403
    results.push(gateResult(
      'module-gate',
      `${probe.method} ${probe.path}`,
      ok,
      '403 for a non-platform-admin',
      String(res.status),
      res.body
    ))
  }

  // /me/permissions is gated on identity alone: a client must be able to
  // discover that it may do nothing, without first being refused.
  const me = await limitedClient.request('GET', '/api/v1/me/permissions', {
    opKey: 'GET /api/v1/me/permissions (limited identity)',
  })
  results.push(gateResult(
    'module-gate',
    'GET /api/v1/me/permissions',
    me.status === 200,
    '200 even for a role with no modules',
    String(me.status),
    me.body
  ))

  return results
}
