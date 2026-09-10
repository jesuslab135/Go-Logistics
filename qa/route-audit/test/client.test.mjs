import { test } from 'node:test'
import assert from 'node:assert/strict'
import { buildUrl } from '../lib/client.mjs'

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
