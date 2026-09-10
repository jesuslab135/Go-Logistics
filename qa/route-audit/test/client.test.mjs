import { test } from 'node:test'
import assert from 'node:assert/strict'
import { buildUrl, ApiClient } from '../lib/client.mjs'
import { createRecorder } from '../lib/recorder.mjs'

test('buildUrl joins base and path without doubling the slash', () => {
  assert.equal(
    buildUrl('https://api.example.com', '/api/v1/vendors'),
    'https://api.example.com/api/v1/vendors'
  )
  assert.equal(
    buildUrl('https://api.example.com/', '/api/v1/vendors'),
    'https://api.example.com/api/v1/vendors'
  )
})

test('buildUrl appends a query string when given one', () => {
  assert.equal(
    buildUrl('https://api.example.com', '/api/v1/vendors', { limit: 1, offset: 0 }),
    'https://api.example.com/api/v1/vendors?limit=1&offset=0'
  )
})

test('buildUrl omits undefined and null query values', () => {
  assert.equal(
    buildUrl('https://api.example.com', '/x', { a: 1, b: undefined, c: null }),
    'https://api.example.com/x?a=1'
  )
})

test('buildUrl encodes values', () => {
  assert.equal(
    buildUrl('https://api.example.com', '/x', { q: 'ZZ TEST&1' }),
    'https://api.example.com/x?q=ZZ+TEST%261'
  )
})

test('request returns status 0 and does not throw on network error', async () => {
  const recorder = createRecorder()
  const client = new ApiClient({ baseUrl: 'http://127.0.0.1:1', recorder })

  const res = await client.request('GET', '/test')

  assert.equal(res.status, 0)
  assert.equal(res.body.error.code, 'network')
  assert.ok(typeof res.body.error.message === 'string')
  assert.ok(res.durationMs >= 0)

  const entries = recorder.entries()
  assert.equal(entries.length, 1)
  assert.equal(entries[0].status, 0)
  assert.equal(entries[0].method, 'GET')
  assert.equal(entries[0].url, '/test')
})
