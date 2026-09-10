import { test } from 'node:test'
import assert from 'node:assert/strict'
import { conforms } from '../lib/conform.mjs'

const spec = {
  definitions: {
    'dto.Thing': {
      type: 'object',
      required: ['id', 'name'],
      properties: {
        id: { type: 'integer' },
        name: { type: 'string' },
        active: { type: 'boolean' },
      },
    },
    'dto.ThingPage': {
      type: 'object',
      properties: {
        data: { type: 'array', items: { $ref: '#/definitions/dto.Thing' } },
        total: { type: 'integer' },
      },
    },
  },
}
const thing = spec.definitions['dto.Thing']

test('a conforming object passes', () => {
  const r = conforms(spec, thing, { id: 1, name: 'a', active: true })
  assert.equal(r.ok, true)
  assert.deepEqual(r.missing, [])
  assert.deepEqual(r.extra, [])
})

test('a missing required key is reported', () => {
  const r = conforms(spec, thing, { id: 1 })
  assert.equal(r.ok, false)
  assert.deepEqual(r.missing, ['name'])
})

test('an undocumented key is reported as extra', () => {
  const r = conforms(spec, thing, { id: 1, name: 'a', surprise: 9 })
  assert.equal(r.ok, false)
  assert.deepEqual(r.extra, ['surprise'])
})

test('a wrong type is reported with its path', () => {
  const r = conforms(spec, thing, { id: 'seven', name: 'a' })
  assert.equal(r.ok, false)
  assert.deepEqual(r.typeErrors, [{ path: 'id', expected: 'integer', actual: 'string' }])
})

test('numeric strings satisfy number, because money serialises as a string', () => {
  const money = { type: 'object', properties: { subtotal: { type: 'number' } } }
  const r = conforms(spec, money, { subtotal: '100.50' })
  assert.equal(r.ok, true)
})

test('a non-numeric string does not satisfy number', () => {
  const money = { type: 'object', properties: { subtotal: { type: 'number' } } }
  const r = conforms(spec, money, { subtotal: 'abc' })
  assert.equal(r.ok, false)
})

test('null satisfies any declared type, because nullable is not expressed in this spec', () => {
  const r = conforms(spec, thing, { id: 1, name: 'a', active: null })
  assert.equal(r.ok, true)
})

test('array items are checked through their ref', () => {
  const page = spec.definitions['dto.ThingPage']
  const r = conforms(spec, page, { data: [{ id: 1 }], total: 1 })
  assert.equal(r.ok, false)
  assert.deepEqual(r.missing, ['data[0].name'])
})
