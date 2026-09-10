import { test } from 'node:test'
import assert from 'node:assert/strict'
import { createRecorder, redact } from '../lib/recorder.mjs'

test('redact masks Authorization headers', () => {
  const out = redact({ headers: { Authorization: 'Bearer eyJhbGciOi.secret.sig' } })
  assert.equal(out.headers.Authorization, '[REDACTED]')
})

test('redact masks Authorization regardless of header casing', () => {
  const out = redact({ headers: { authorization: 'Bearer abc' } })
  assert.equal(out.headers.authorization, '[REDACTED]')
})

test('redact masks password and token fields at any depth', () => {
  const out = redact({ a: { password: 'hunter2', access_token: 'abc', refresh_token: 'def' } })
  assert.equal(out.a.password, '[REDACTED]')
  assert.equal(out.a.access_token, '[REDACTED]')
  assert.equal(out.a.refresh_token, '[REDACTED]')
})

test('redact leaves ordinary values untouched', () => {
  const out = redact({ name: 'ZZ-TEST-1-vendor', id: 7, nested: [{ ok: true }] })
  assert.deepEqual(out, { name: 'ZZ-TEST-1-vendor', id: 7, nested: [{ ok: true }] })
})

test('recorder stores entries and filters them by operation', () => {
  const rec = createRecorder()
  rec.record({ opKey: 'GET /api/v1/vendors', status: 200 })
  rec.record({ opKey: 'POST /api/v1/vendors', status: 201 })
  assert.equal(rec.entries().length, 2)
  assert.equal(rec.forOp('GET /api/v1/vendors').length, 1)
})

test('recorder redacts on the way in, not on the way out', () => {
  const rec = createRecorder()
  rec.record({ opKey: 'POST /auth/login', requestBody: { password: 'hunter2' } })
  assert.equal(rec.entries()[0].requestBody.password, '[REDACTED]')
})
