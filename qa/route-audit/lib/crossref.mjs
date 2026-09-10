export function normalisePath(url) {
  let path = url
  const originMatch = /^https?:\/\/[^/]+(\/.*)$/.exec(url)
  if (originMatch) path = originMatch[1]
  path = path.split('?')[0]

  const placeholders = ['{id}', '{child_id}']
  let seen = 0
  return path
    .split('/')
    .map(seg => {
      if (!/^\d+$/.test(seg)) return seg
      const token = placeholders[seen] ?? '{id}'
      seen += 1
      return token
    })
    .join('/')
}

export function matchOperation(ops, method, url) {
  const path = normalisePath(url)
  return ops.find(o => o.method === method.toUpperCase() && o.path === path) ?? null
}

let counter = 0
function nextId(prefix) {
  counter += 1
  return `${prefix}-${String(counter).padStart(3, '0')}`
}

export function resetFindingIds() { counter = 0 }

export function crossReference(spec, ops, captures) {
  const findings = []

  for (const capture of captures) {
    for (const req of capture.requests ?? []) {
      const op = matchOperation(ops, req.method, req.url)

      if (!op) {
        findings.push({
          id: nextId('FE'),
          layer: 'frontend',
          severity: 'high',
          route: capture.route,
          summary: `The screen calls ${req.method} ${normalisePath(req.url)}, for which there is no such backend operation`,
          expected: 'a request to an operation the API defines',
          actual: `${req.method} ${normalisePath(req.url)} -> ${req.status}`,
          evidence: req,
        })
        continue
      }

      if (req.status >= 500) {
        findings.push({
          id: nextId('BE'),
          layer: 'backend',
          severity: 'high',
          route: capture.route,
          summary: `${op.opKey} returned ${req.status} to the frontend`,
          expected: `${op.successStatus ?? '2xx'}`,
          actual: String(req.status),
          evidence: req,
        })
        continue
      }

      if (req.status === 400 || req.status === 422) {
        findings.push({
          id: nextId('CT'),
          layer: 'contract',
          severity: 'high',
          route: capture.route,
          summary: `${op.opKey} rejected the payload the frontend sent`,
          expected: `${op.successStatus ?? '2xx'}`,
          actual: `${req.status} ${JSON.stringify(req.responseSummary ?? '')}`,
          evidence: req,
        })
        continue
      }

      if (req.status === 403 || req.status === 404) {
        findings.push({
          id: nextId('CT'),
          layer: 'contract',
          severity: 'medium',
          route: capture.route,
          summary: `${op.opKey} refused the frontend with ${req.status}`,
          expected: `${op.successStatus ?? '2xx'} for an authorised admin`,
          actual: String(req.status),
          evidence: req,
        })
      }
    }

    const allOk = (capture.requests ?? []).every(r => r.status < 400)
    if (capture.rendered === 'blank' && allOk) {
      findings.push({
        id: nextId('FE'),
        layer: 'frontend',
        severity: 'high',
        route: capture.route,
        summary: 'The screen rendered nothing even though every request succeeded',
        expected: 'rendered data or a deliberate empty state',
        actual: `blank; console: ${(capture.consoleErrors ?? []).join(' | ') || 'no errors logged'}`,
        evidence: { consoleErrors: capture.consoleErrors ?? [] },
      })
    }

    if (capture.rendered === 'error' && allOk) {
      findings.push({
        id: nextId('FE'),
        layer: 'frontend',
        severity: 'medium',
        route: capture.route,
        summary: 'The screen showed an error state even though every request succeeded',
        expected: 'rendered data or a deliberate empty state',
        actual: `error; console: ${(capture.consoleErrors ?? []).join(' | ') || 'no errors logged'}`,
        evidence: { consoleErrors: capture.consoleErrors ?? [] },
      })
    }
  }

  return findings
}

export function unusedOperations(ops, captures) {
  const called = new Set()
  for (const capture of captures) {
    for (const req of capture.requests ?? []) {
      const op = matchOperation(ops, req.method, req.url)
      if (op) called.add(op.opKey)
    }
  }
  return ops.filter(o => !called.has(o.opKey)).map(o => o.opKey)
}
