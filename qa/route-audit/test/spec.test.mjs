import { test } from 'node:test'
import assert from 'node:assert/strict'
import { loadSpec, listOperations, resolveRef } from '../lib/spec.mjs'
import { config } from '../config.mjs'

test('loadSpec reads the committed swagger document', () => {
  const spec = loadSpec(config.specPath)
  assert.ok(spec.paths, 'spec has paths')
  assert.ok(spec.definitions, 'spec has definitions')
})

test('listOperations enumerates every HTTP verb across every path', () => {
  const ops = listOperations(loadSpec(config.specPath))
  assert.equal(ops.length, 374)
  const vendorsList = ops.find(o => o.opKey === 'GET /api/v1/vendors')
  assert.ok(vendorsList, 'GET /api/v1/vendors is enumerated')
  assert.equal(vendorsList.successStatus, 200)
  assert.deepEqual(vendorsList.tags, ['vendors'])
})

test('listOperations resolves the success schema, not just its ref', () => {
  const spec = loadSpec(config.specPath)
  const ops = listOperations(spec)
  const op = ops.find(o => o.opKey === 'GET /api/v1/vendors')
  assert.equal(op.successSchema.type, 'object')
  assert.ok(op.successSchema.properties.data, 'page envelope has data')
  assert.ok(op.successSchema.properties.has_next, 'page envelope has has_next')
})

test('resolveRef follows a definition pointer', () => {
  const spec = loadSpec(config.specPath)
  const schema = resolveRef(spec, '#/definitions/dto.ErrorResponse')
  assert.ok(schema.properties.error, 'ErrorResponse wraps an error body')
})
