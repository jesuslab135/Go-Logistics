import { conforms } from './conform.mjs'

export const MISSING_ID = 999999999

// Routes that are public by design: no bearer token is expected, so the
// unauthenticated-probe check does not apply to them. POST /auth/login,
// /auth/logout and /auth/refresh correctly return 400 (invalid empty body)
// for an anonymous probe, and GET /healthz correctly returns 200 — that is
// the entire point of a health check. POST /auth/switch-company is
// deliberately NOT in this set: it requires a valid access token.
export const PUBLIC_ROUTES = new Set([
  'POST /auth/login',
  'POST /auth/logout',
  'POST /auth/refresh',
  'GET /healthz',
])

const SEVERITY = {
  'unauthenticated': 'high',
  'tenant-isolation': 'high',
  'module-gate': 'high',
  'happy-path': 'high',
  'retired-verb': 'medium',
  'not-found': 'medium',
  'list-contract': 'medium',
  'validation': 'medium',
  'conformance': 'low',
}

export function classify(check, ok) {
  return ok ? 'info' : (SEVERITY[check] ?? 'medium')
}

export function fillPath(path, ids) {
  return path.replace(/\{(\w+)\}/g, (match, key) =>
    key in ids ? String(ids[key]) : match
  )
}

function result(check, opKey, ok, expected, actual, evidence) {
  return { check, opKey, ok, expected, actual, severity: classify(check, ok), evidence }
}

export async function checkUnauthenticated(client, op) {
  const path = fillPath(op.path, { id: MISSING_ID, child_id: MISSING_ID })
  const res = await client.request(op.method, path, {
    anonymous: true,
    opKey: op.opKey,
    retryOn401: false,
    body: op.requestSchema ? {} : undefined,
  })
  // A public-by-design route (POST /auth/login, /auth/logout, /auth/refresh,
  // GET /healthz) has no 401 expectation at all — the anonymous probe is
  // still sent, so `actual` carries what the server really answered (useful
  // evidence: healthz says 200, login says 400 to an empty body, both
  // correctly ungated), but the check is recorded as a PASSING result
  // rather than skipped. That keeps this op's single coverage row intact
  // (it still comes only from the verb dispatch further down) while making
  // the exemption visible in the report instead of silently absent.
  if (PUBLIC_ROUTES.has(op.opKey)) {
    return {
      check: 'unauthenticated',
      opKey: op.opKey,
      ok: true,
      expected: 'no authentication required — this route is public by design',
      actual: `${res.status} (not gated)`,
      severity: 'info',
      evidence: res.body,
    }
  }
  const ok = res.status === 401
  return result('unauthenticated', op.opKey, ok, '401', String(res.status), res.body)
}

export async function checkNotFound(client, op) {
  if (!op.path.includes('{id}')) return null
  if (op.method === 'POST' && !op.path.endsWith('}')) {
    // Action routes such as /purchase-orders/{id}/approve are covered by the
    // workflow suites, which know the legal states. A bare 404 probe here
    // would report a state error as a missing row.
    return null
  }
  // Nested many-to-many link routes use a named target param instead of
  // {child_id} (e.g. {employee_id}, {issue_id}, {fault_id}). Extract every
  // placeholder in the path so each one gets a missing-id value, not just
  // the two conventional names.
  const placeholders = [...op.path.matchAll(/\{(\w+)\}/g)].map(m => m[1])
  const missingIds = Object.fromEntries(placeholders.map(name => [name, MISSING_ID]))
  const path = fillPath(op.path, missingIds)
  const res = await client.request(op.method, path, {
    opKey: op.opKey,
    body: op.requestSchema ? {} : undefined,
  })
  const ok = res.status === 404
  return result('not-found', op.opKey, ok, '404', String(res.status), res.body)
}

export async function checkHappyPath(client, op, ids, body) {
  const path = fillPath(op.path, ids)
  const res = await client.request(op.method, path, { opKey: op.opKey, body })
  const expected = op.successStatus ?? 200
  const ok = res.status === expected
  return {
    ...result('happy-path', op.opKey, ok, String(expected), String(res.status), res.body),
    response: res,
  }
}

export function checkConformance(spec, op, response) {
  if (!op.successSchema) return null
  if (response.status < 200 || response.status >= 300) return null
  const c = conforms(spec, op.successSchema, response.body)
  const parts = []
  if (c.missing.length) parts.push(`missing: ${c.missing.join(', ')}`)
  if (c.extra.length) parts.push(`undocumented: ${c.extra.join(', ')}`)
  if (c.typeErrors.length) {
    parts.push(`type: ${c.typeErrors.map(e => `${e.path} expected ${e.expected} got ${e.actual}`).join('; ')}`)
  }
  return result(
    'conformance', op.opKey, c.ok,
    'response matches the documented schema',
    parts.length ? parts.join(' | ') : 'matches',
    { missing: c.missing, extra: c.extra, typeErrors: c.typeErrors }
  )
}

export async function checkListContract(client, op) {
  const isList = op.method === 'GET'
    && !op.path.endsWith('}')
    && (op.params ?? []).some(p => p.name === 'limit')
  if (!isList) return null

  const path = fillPath(op.path, { id: MISSING_ID, child_id: MISSING_ID })
  if (path.includes('{')) return null

  const res = await client.request(op.method, path, { opKey: op.opKey, query: { limit: 1, offset: 0 } })
  if (res.status !== 200) {
    return result('list-contract', op.opKey, false, '200 with a page envelope', String(res.status), res.body)
  }
  const b = res.body ?? {}
  const problems = []
  if (!Array.isArray(b.data)) problems.push('data is not an array')
  if (typeof b.total !== 'number') problems.push('total is not a number')
  if (typeof b.has_next !== 'boolean') problems.push('has_next is not a boolean')
  if (Array.isArray(b.data) && b.data.length > 1) problems.push(`limit=1 returned ${b.data.length} rows`)
  const ok = problems.length === 0
  return result('list-contract', op.opKey, ok, 'limit honoured, envelope complete',
    ok ? 'correct' : problems.join('; '), { total: b.total, limit: b.limit, returned: b.data?.length })
}

export async function checkRetiredVerb(client, op) {
  const retired = op.path === '/api/v1/inventory-journal-entries/{id}'
    && (op.method === 'PUT' || op.method === 'DELETE')
  if (!retired) return null
  const path = fillPath(op.path, { id: MISSING_ID })
  const res = await client.request(op.method, path, { opKey: op.opKey, body: op.requestSchema ? {} : undefined })
  const ok = res.status === 405 && typeof res.body?.error?.message === 'string'
  return result('retired-verb', op.opKey, ok, '405 naming the replacement route',
    `${res.status} ${res.body?.error?.message ?? ''}`.trim(), res.body)
}
