// qa/route-audit/lib/workflows.mjs
import { classify } from './checks.mjs'

// Mirrors internal/domain/purchaseorder/workflow.go, read at plan time.
// Verified against the source by test/workflows.test.mjs — if that test fails,
// the source moved and this table is the thing that is wrong.
//
// Rejection is not terminal: it returns to DRAFT via an explicit `revise`, so
// a rejected order keeps its line items instead of forcing a new one.
export const PO_TRANSITIONS = [
  { action: 'submit', legalFrom: ['DRAFT', 'REJECTED'], to: 'PENDING_APPROVAL', requiresApproval: false, requiresReason: false },
  { action: 'approve', legalFrom: ['PENDING_APPROVAL'], to: 'APPROVED', requiresApproval: true, requiresReason: false },
  { action: 'reject', legalFrom: ['PENDING_APPROVAL'], to: 'REJECTED', requiresApproval: true, requiresReason: true },
  { action: 'revise', legalFrom: ['REJECTED'], to: 'DRAFT', requiresApproval: false, requiresReason: false },
  { action: 'purchase', legalFrom: ['APPROVED'], to: 'PURCHASED', requiresApproval: false, requiresReason: false },
  { action: 'receive-partial', legalFrom: ['PURCHASED'], to: 'RECEIVED_PARTIAL', requiresApproval: false, requiresReason: false },
  { action: 'receive-full', legalFrom: ['PURCHASED', 'RECEIVED_PARTIAL'], to: 'RECEIVED_FULL', requiresApproval: false, requiresReason: false },
  { action: 'close', legalFrom: ['RECEIVED_FULL'], to: 'CLOSED', requiresApproval: false, requiresReason: false },
]

// One ordered walk that visits every action exactly once. The reject/revise
// detour is what makes that possible: without it, reaching CLOSED would leave
// three actions unexercised.
export const PO_WALK = [
  { action: 'submit', expect: 'PENDING_APPROVAL' },
  { action: 'reject', expect: 'REJECTED', body: { reason: 'audit: exercising the reject branch' } },
  { action: 'revise', expect: 'DRAFT' },
  { action: 'submit', expect: 'PENDING_APPROVAL' },
  { action: 'approve', expect: 'APPROVED' },
  { action: 'purchase', expect: 'PURCHASED' },
  { action: 'receive-partial', expect: 'RECEIVED_PARTIAL' },
  { action: 'receive-full', expect: 'RECEIVED_FULL' },
  { action: 'close', expect: 'CLOSED' },
]

function wf(check, opKey, ok, expected, actual, evidence) {
  return { check, opKey, ok, expected, actual, severity: classify(check, ok), evidence }
}

export async function runPurchaseOrderWorkflow(client, graph) {
  const results = []
  const poId = graph.ids.purchaseOrder
  if (!poId) {
    return [wf('workflow', 'purchase-order lifecycle', false,
      'a fixture purchase order', 'fixture was not created', null)]
  }

  // An illegal transition first: closing a draft must be refused. Running it
  // before the legal walk means a false pass cannot come from the order
  // already sitting in the right state.
  const illegal = await client.request('POST', `/api/v1/purchase-orders/${poId}/close`, {
    opKey: 'POST /api/v1/purchase-orders/{id}/close (illegal from DRAFT)',
  })
  results.push(wf('workflow', 'POST /api/v1/purchase-orders/{id}/close',
    illegal.status === 409 || illegal.status === 400,
    '409 or 400 refusing an illegal transition',
    String(illegal.status), illegal.body))

  // Rejection must say why. Sending none has to be refused, or a buyer is
  // sent back with nothing to act on.
  const submitted = await client.request('POST', `/api/v1/purchase-orders/${poId}/submit`, {
    opKey: 'POST /api/v1/purchase-orders/{id}/submit (for the reason check)',
  })
  if (submitted.status >= 200 && submitted.status < 300) {
    const noReason = await client.request('POST', `/api/v1/purchase-orders/${poId}/reject`, {
      body: {},
      opKey: 'POST /api/v1/purchase-orders/{id}/reject (no reason)',
    })
    results.push(wf('workflow', 'POST /api/v1/purchase-orders/{id}/reject',
      noReason.status >= 400,
      '4xx refusing a rejection that carries no reason',
      String(noReason.status), noReason.body))

    // Put the order back in DRAFT so the full walk starts from the top.
    // dto.PurchaseOrderTransitionRequest declares only `reason` — verified
    // against internal/http/dto/purchase_order.go — so this body is correct
    // as written.
    await client.request('POST', `/api/v1/purchase-orders/${poId}/reject`, {
      body: { reason: `${graph.tag} reset` },
      opKey: 'POST /api/v1/purchase-orders/{id}/reject (reset)',
    })
    await client.request('POST', `/api/v1/purchase-orders/${poId}/revise`, {
      opKey: 'POST /api/v1/purchase-orders/{id}/revise (reset)',
    })
  }

  // The full walk. Every action is visited once, and the resulting state is
  // read back rather than inferred from the status code.
  for (const step of PO_WALK) {
    const res = await client.request('POST', `/api/v1/purchase-orders/${poId}/${step.action}`, {
      body: step.body,
      opKey: `POST /api/v1/purchase-orders/{id}/${step.action}`,
    })
    const moved = res.status >= 200 && res.status < 300

    const after = await client.request('GET', `/api/v1/purchase-orders/${poId}`, {
      opKey: 'GET /api/v1/purchase-orders/{id} (state read-back)',
    })
    const state = after.body?.state ?? null

    results.push(wf('workflow', `POST /api/v1/purchase-orders/{id}/${step.action}`,
      moved && state === step.expect,
      `2xx leaving the order in ${step.expect}`,
      `${res.status}, state ${state ?? 'unknown'}`,
      { response: res.body, state }))

    // A failed step invalidates every later expectation, so stop rather than
    // report a cascade of failures that all trace to one cause.
    if (!moved || state !== step.expect) {
      results.push(wf('workflow', 'purchase-order walk',
        false,
        `the remaining steps after ${step.action}`,
        `abandoned: the order is in ${state ?? 'unknown'}, not ${step.expect}`,
        { stoppedAt: step.action }))
      break
    }
  }

  // The history must have recorded the walk.
  const logs = await client.request('GET', `/api/v1/purchase-orders/${poId}/status-logs`, {
    opKey: 'GET /api/v1/purchase-orders/{id}/status-logs',
  })
  const count = Array.isArray(logs.body?.data) ? logs.body.data.length : 0
  results.push(wf('workflow', 'GET /api/v1/purchase-orders/{id}/status-logs',
    logs.status === 200 && count > 0,
    'a status log row per completed transition',
    `${logs.status}, ${count} rows`, logs.body))

  // Departing from the computed total is an audited act, not a field write.
  //
  // dto.TotalOverrideRequest (internal/http/dto/money.go) declares `amount`
  // (a *decimal.Decimal) and `reason` — NOT `total`. `total` is not a field
  // on the struct at all, so Gin would silently drop it: the request would
  // still 2xx, but the override would never be set (Amount stays nil, and
  // because Amount is nil the handler even clears Reason back to ""), which
  // is exactly the class of bug this harness exists to catch. Sending
  // `amount` here.
  const override = await client.request('POST', `/api/v1/purchase-orders/${poId}/override-total`, {
    body: { amount: '123.45', reason: `${graph.tag} override` },
    opKey: 'POST /api/v1/purchase-orders/{id}/override-total',
  })
  results.push(wf('workflow', 'POST /api/v1/purchase-orders/{id}/override-total',
    override.status >= 200 && override.status < 300,
    '2xx recording an audited override',
    String(override.status), override.body))

  return results
}

export async function runWorkOrderWorkflow(client, graph) {
  const results = []
  const woId = graph.ids.workOrder
  const statusId = graph.ids.workOrderStatus
  if (!woId || !statusId) {
    return [wf('workflow', 'work-order lifecycle', false,
      'fixture work order and status', 'fixtures were not created', null)]
  }

  const before = await client.request('GET', `/api/v1/work-orders/${woId}/status-logs`, {
    opKey: 'GET /api/v1/work-orders/{id}/status-logs (before)',
  })
  const beforeCount = Array.isArray(before.body?.data) ? before.body.data.length : 0

  // A second status to move to. Creating one is cheaper than assuming the
  // company already has two. dto.CreateWorkOrderStatusRequest declares
  // `name` — verified against internal/http/dto/work_order_status.go.
  const second = await client.request('POST', '/api/v1/work-order-statuses', {
    body: { name: `${graph.tag}-wo-status-2` },
    opKey: 'POST /api/v1/work-order-statuses (workflow)',
  })
  const secondId = second.body?.id

  if (secondId) {
    // The PUT round-trips the GET response as the whole body. That is safe:
    // dto.UpdateWorkOrderRequest's fields are a strict subset of
    // dto.WorkOrderResponse's field names (verified against
    // internal/http/dto/work_order.go) — the response-only fields the GET
    // carries (id, company_id, computed subtotals/totals, created_at,
    // updated_at, ...) are simply names Gin does not bind and silently
    // ignores, not names that collide with anything Update declares.
    const current = await client.request('GET', `/api/v1/work-orders/${woId}`, {
      opKey: 'GET /api/v1/work-orders/{id} (workflow)',
    })
    const updated = await client.request('PUT', `/api/v1/work-orders/${woId}`, {
      body: { ...current.body, status_id: secondId },
      opKey: 'PUT /api/v1/work-orders/{id} (status change)',
    })
    results.push(wf('workflow', 'PUT /api/v1/work-orders/{id}',
      updated.status >= 200 && updated.status < 300,
      '2xx changing the status',
      String(updated.status), updated.body))

    const after = await client.request('GET', `/api/v1/work-orders/${woId}/status-logs`, {
      opKey: 'GET /api/v1/work-orders/{id}/status-logs (after)',
    })
    const afterCount = Array.isArray(after.body?.data) ? after.body.data.length : 0
    results.push(wf('workflow', 'work-order status log is appended automatically',
      afterCount > beforeCount,
      'one more status log row than before the change',
      `${beforeCount} -> ${afterCount}`, after.body))
  }

  // The status history is append-only: writing to it directly must be refused.
  const direct = await client.request('POST', `/api/v1/work-orders/${woId}/status-logs`, {
    body: { status_id: statusId },
    opKey: 'POST /api/v1/work-orders/{id}/status-logs (should not exist)',
  })
  results.push(wf('workflow', 'POST /api/v1/work-orders/{id}/status-logs',
    direct.status === 404 || direct.status === 405,
    '404 or 405, because the history is append-only',
    String(direct.status), direct.body))

  return results
}

export async function runJournalWorkflow(client, graph) {
  const results = []
  const entryId = graph.ids.journalEntry
  if (!entryId) {
    return [wf('workflow', 'inventory journal', false,
      'a fixture journal entry', 'fixture was not created', null)]
  }

  for (const method of ['PUT', 'DELETE']) {
    const res = await client.request(method, `/api/v1/inventory-journal-entries/${entryId}`, {
      body: method === 'PUT' ? {} : undefined,
      opKey: `${method} /api/v1/inventory-journal-entries/{id}`,
    })
    const names = typeof res.body?.error?.message === 'string'
      && res.body.error.message.includes('reverse')
    results.push(wf('workflow', `${method} /api/v1/inventory-journal-entries/{id}`,
      res.status === 405 && names,
      '405 naming the reverse route as the replacement',
      `${res.status} ${res.body?.error?.message ?? ''}`.trim(), res.body))
  }

  // internal/http/handler/inventory_journal_entry.go's Reverse handler takes
  // no request body at all — it calls h.store.Reverse(ctx, id) with nothing
  // but the path id, and binds no JSON. `notes` is not a field Gin drops on
  // an existing struct here; there is no struct to bind against, so sending
  // it would be pure noise. No body is sent.
  const reversal = await client.request('POST', `/api/v1/inventory-journal-entries/${entryId}/reverse`, {
    opKey: 'POST /api/v1/inventory-journal-entries/{id}/reverse',
  })
  results.push(wf('workflow', 'POST /api/v1/inventory-journal-entries/{id}/reverse',
    reversal.status >= 200 && reversal.status < 300,
    '2xx creating a balancing entry',
    String(reversal.status), reversal.body))

  // Reversing a reversal must be refused, or the ledger can be spun forever.
  if (reversal.body?.id) {
    const again = await client.request('POST', `/api/v1/inventory-journal-entries/${reversal.body.id}/reverse`, {
      opKey: 'POST /api/v1/inventory-journal-entries/{id}/reverse (double)',
    })
    results.push(wf('workflow', 'reversing a reversal is refused',
      again.status >= 400,
      '4xx refusing to reverse a reversal',
      String(again.status), again.body))
  }

  return results
}

export async function runArchiveWorkflow(client, graph) {
  const results = []
  const targets = [
    { key: 'vendor', path: '/api/v1/vendors', list: '/api/v1/vendors' },
    { key: 'part', path: '/api/v1/parts', list: '/api/v1/parts' },
    { key: 'serviceTask', path: '/api/v1/service-tasks', list: '/api/v1/service-tasks' },
    { key: 'inspectionForm', path: '/api/v1/inspection-forms', list: '/api/v1/inspection-forms' },
  ]

  for (const t of targets) {
    const id = graph.ids[t.key]
    if (!id) continue

    const archived = await client.request('POST', `${t.path}/${id}/archive`, {
      opKey: `POST ${t.path}/{id}/archive`,
    })
    results.push(wf('workflow', `POST ${t.path}/{id}/archive`,
      archived.status >= 200 && archived.status < 300,
      '2xx archiving the row',
      String(archived.status), archived.body))

    const hidden = await client.request('GET', t.list, {
      query: { limit: 200 }, opKey: `GET ${t.list} (archived hidden)`,
    })
    const visible = (hidden.body?.data ?? []).some(r => r.id === id)
    results.push(wf('workflow', `${t.list} hides archived rows by default`,
      !visible, 'the archived row is absent',
      visible ? 'the archived row is still listed' : 'absent', null))

    const included = await client.request('GET', t.list, {
      query: { limit: 200, include_archived: true },
      opKey: `GET ${t.list}?include_archived=true`,
    })
    const reappears = (included.body?.data ?? []).some(r => r.id === id)
    results.push(wf('workflow', `${t.list}?include_archived=true reveals archived rows`,
      reappears, 'the archived row is present',
      reappears ? 'present' : 'still absent', null))

    const restored = await client.request('POST', `${t.path}/${id}/restore`, {
      opKey: `POST ${t.path}/{id}/restore`,
    })
    results.push(wf('workflow', `POST ${t.path}/{id}/restore`,
      restored.status >= 200 && restored.status < 300,
      '2xx restoring the row',
      String(restored.status), restored.body))
  }

  return results
}

export async function runUploadWorkflow(client, graph) {
  const results = []

  // A one-pixel PNG, inline so the harness needs no fixture file on disk.
  const pngBase64 =
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=='
  const bytes = Buffer.from(pngBase64, 'base64')

  const form = new FormData()
  form.append('file', new Blob([bytes], { type: 'image/png' }), `${graph.tag}.png`)
  // `receipt` (not `asset_photo`) on purpose: internal/http/handler/upload.go's
  // uploadPurposes table maps asset_photo and company_logo to
  // storage.VisibilityPublic — everything else, including receipt, to
  // VisibilityPrivate. The point of this suite is to probe the private-bucket
  // access boundary below, which asset_photo can never exercise because a
  // public object's URL is never signed and an unsigned GET is expected to
  // succeed. `receipt` accepts any media type, including this PNG.
  form.append('purpose', 'receipt')

  const url = `${client.baseUrl}/api/v1/uploads`
  const res = await fetch(url, {
    method: 'POST',
    headers: { Authorization: `Bearer ${client.token}` },
    body: form,
  })
  const body = await res.json().catch(() => null)
  client.recorder.record({
    opKey: 'POST /api/v1/uploads', method: 'POST', url: '/api/v1/uploads',
    requestBody: '(multipart)', status: res.status, responseBody: body, durationMs: 0,
  })

  results.push(wf('workflow', 'POST /api/v1/uploads',
    res.status >= 200 && res.status < 300,
    '2xx returning a stored URL',
    String(res.status), body))

  // The private bucket is the access boundary. An unsigned GET must fail.
  //
  // dto.UploadResponse (internal/http/dto/upload.go) names the field
  // `visibility`, set to the literal string "private" or "public" — that is
  // what decides the boundary, not a URL substring. The brief's original
  // check matched the literal text "/fleet-private/", which appears nowhere
  // in this codebase: the real key prefixes are "uploads/private/" and
  // "uploads/public/" (internal/platform/storage/storage.go). Matching the
  // wrong string would make this check silently inert — it would never run,
  // pass or fail, for any real response — so it is replaced with a read of
  // the response's own `visibility` field.
  if (body?.visibility === 'private' && typeof body?.url === 'string') {
    const bare = body.url.split('?')[0]
    const unsigned = await fetch(bare)
    results.push({
      check: 'private-bucket',
      opKey: 'GET fleet-private (unsigned)',
      ok: unsigned.status === 403,
      expected: '403 for an unsigned request to the private bucket',
      actual: String(unsigned.status),
      severity: unsigned.status === 403 ? 'info' : 'high',
      evidence: { url: bare },
    })
  }

  return results
}

export async function runWorkflows(client, graph) {
  return [
    ...await runPurchaseOrderWorkflow(client, graph),
    ...await runWorkOrderWorkflow(client, graph),
    ...await runJournalWorkflow(client, graph),
    ...await runArchiveWorkflow(client, graph),
    ...await runUploadWorkflow(client, graph),
  ]
}
