import { conforms } from './conform.mjs'

export const MISSING_ID = 999999999

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
  const path = fillPath(op.path, { id: MISSING_ID, child_id: MISSING_ID })
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
