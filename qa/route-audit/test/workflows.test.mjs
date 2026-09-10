// qa/route-audit/test/workflows.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { PO_TRANSITIONS, PO_WALK, runPurchaseOrderWorkflow } from '../lib/workflows.mjs'

const src = readFileSync('internal/domain/purchaseorder/workflow.go', 'utf8')

test('every PO action in the harness exists in the Go workflow source', () => {
  for (const t of PO_TRANSITIONS) {
    assert.ok(
      src.includes(`ActionSubmit`) && src.includes(`"${t.action}"`),
      `action ${t.action} is not named in workflow.go`
    )
  }
})

test('the harness table covers every action the Go source declares', () => {
  // Action constants are declared as ActionX = "value" in a const block.
  const declared = [...src.matchAll(/Action\w+\s*=\s*"([a-z-]+)"/g)].map(m => m[1])
  assert.equal(declared.length, 8, 'expected 8 declared actions')
  const covered = new Set(PO_TRANSITIONS.map(t => t.action))
  for (const action of declared) {
    assert.ok(covered.has(action), `harness table is missing ${action}`)
  }
})

test('the transition table declares a legal source state for each action', () => {
  for (const t of PO_TRANSITIONS) {
    assert.ok(Array.isArray(t.legalFrom), `${t.action} has no legalFrom`)
    assert.ok(t.legalFrom.length > 0, `${t.action} has an empty legalFrom`)
    assert.ok(typeof t.to === 'string', `${t.action} has no destination state`)
  }
})

test('approve and reject are the transitions gated on the approve action', () => {
  const gated = PO_TRANSITIONS.filter(t => t.requiresApproval).map(t => t.action).sort()
  assert.deepEqual(gated, ['approve', 'reject'])
})

test('reject is the transition that requires a reason', () => {
  const needReason = PO_TRANSITIONS.filter(t => t.requiresReason).map(t => t.action)
  assert.deepEqual(needReason, ['reject'])
})

test('the walk visits every declared action at least once', () => {
  const visited = new Set(PO_WALK.map(s => s.action))
  for (const t of PO_TRANSITIONS) {
    assert.ok(visited.has(t.action), `the walk never exercises ${t.action}`)
  }
})

test('the walk chains: each step starts from the previous step\'s destination', () => {
  let state = 'DRAFT'
  for (const step of PO_WALK) {
    const t = PO_TRANSITIONS.find(x => x.action === step.action)
    assert.ok(
      t.legalFrom.includes(state),
      `${step.action} is not legal from ${state}`
    )
    assert.equal(t.to, step.expect, `${step.action} lands in ${t.to}, not ${step.expect}`)
    state = t.to
  }
  assert.equal(state, 'CLOSED', 'the walk should finish in CLOSED')
})

test('the PO walk stops cleanly at the first failed transition rather than cascading', async () => {
  // A stub purchase order state machine that behaves exactly like the real
  // transitions except `approve`, which is forced to fail with an
  // unexpected server error. Without the break-on-failure guard in
  // runPurchaseOrderWorkflow, this would produce eight confusing downstream
  // failures (purchase, receive-partial, receive-full, close, plus their
  // read-backs) that all trace back to the one cause. This pins that a
  // single abandoned marker is recorded instead, and that the walk never
  // even calls the routes downstream of the failure.
  let state = 'DRAFT'
  const calls = []

  const stubClient = {
    async request(method, path, opts) {
      calls.push({ method, path, opKey: opts?.opKey, body: opts?.body })

      if (method === 'GET' && path === '/api/v1/purchase-orders/1') {
        return { status: 200, body: { state } }
      }
      if (method === 'GET' && path === '/api/v1/purchase-orders/1/status-logs') {
        return { status: 200, body: { data: [{ id: 1 }] } }
      }
      if (method === 'POST' && path === '/api/v1/purchase-orders/1/close') {
        if (state !== 'RECEIVED_FULL') return { status: 409, body: { error: { message: 'invalid_transition' } } }
        state = 'CLOSED'
        return { status: 200, body: { state } }
      }
      if (method === 'POST' && path === '/api/v1/purchase-orders/1/submit') {
        if (!['DRAFT', 'REJECTED'].includes(state)) return { status: 409, body: {} }
        state = 'PENDING_APPROVAL'
        return { status: 200, body: { state } }
      }
      if (method === 'POST' && path === '/api/v1/purchase-orders/1/reject') {
        const reason = opts?.body?.reason
        if (!reason) return { status: 422, body: { error: { message: 'a reason is required' } } }
        state = 'REJECTED'
        return { status: 200, body: { state } }
      }
      if (method === 'POST' && path === '/api/v1/purchase-orders/1/revise') {
        state = 'DRAFT'
        return { status: 200, body: { state } }
      }
      if (method === 'POST' && path === '/api/v1/purchase-orders/1/approve') {
        // The forced failure. State does not move.
        return { status: 500, body: { error: { message: 'boom' } } }
      }
      if (method === 'POST' && path === '/api/v1/purchase-orders/1/purchase') {
        throw new Error('the walk must not reach purchase after approve fails')
      }
      if (method === 'POST' && path === '/api/v1/purchase-orders/1/receive-partial') {
        throw new Error('the walk must not reach receive-partial after approve fails')
      }
      if (method === 'POST' && path === '/api/v1/purchase-orders/1/receive-full') {
        throw new Error('the walk must not reach receive-full after approve fails')
      }
      if (method === 'POST' && path === '/api/v1/purchase-orders/1/override-total') {
        return { status: 200, body: {} }
      }
      throw new Error(`stub client received unexpected call: ${method} ${path}`)
    },
  }

  const graph = { runId: 'stub', tag: 'ZZ-TEST-stub', ids: { purchaseOrder: 1 }, created: [], failed: [] }
  const results = await runPurchaseOrderWorkflow(stubClient, graph)

  // The routes downstream of the failed transition must never be called.
  for (const action of ['purchase', 'receive-partial', 'receive-full']) {
    assert.ok(
      !calls.some(c => c.path === `/api/v1/purchase-orders/1/${action}`),
      `${action} must not be called once approve has failed`
    )
  }
  // close was legitimately probed once, up front, as the illegal-transition
  // check — but never again as part of the walk after approve fails.
  const closeCalls = calls.filter(c => c.path === '/api/v1/purchase-orders/1/close')
  assert.equal(closeCalls.length, 1, 'close should only be called for the illegal-transition probe')

  const approveResult = results.find(r => r.opKey === 'POST /api/v1/purchase-orders/{id}/approve')
  assert.ok(approveResult, 'the failed approve step must still be recorded')
  assert.equal(approveResult.ok, false)

  const abandoned = results.filter(r => r.opKey === 'purchase-order walk')
  assert.equal(abandoned.length, 1, 'exactly one abandoned marker, not a cascade of per-action failures')
  assert.equal(abandoned[0].ok, false)
  assert.equal(abandoned[0].evidence?.stoppedAt, 'approve')
  assert.match(String(abandoned[0].actual), /abandoned/)

  // No result for any action past the failure point.
  for (const action of ['purchase', 'receive-partial', 'receive-full']) {
    const r = results.find(r => r.opKey === `POST /api/v1/purchase-orders/{id}/${action}`)
    assert.equal(r, undefined, `no result should exist for ${action} once the walk has stopped`)
  }
})
