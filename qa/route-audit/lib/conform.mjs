import { resolveRef } from './spec.mjs'

function deref(spec, schema) {
  if (!schema) return null
  return schema.$ref ? resolveRef(spec, schema.$ref) : schema
}

function typeOk(expected, value) {
  // This spec never marks a field nullable, yet the API returns null for every
  // optional column. Treating null as a type error would flag most of the
  // surface and hide the real defects.
  if (value === null) return true
  switch (expected) {
    case 'string': return typeof value === 'string'
    case 'boolean': return typeof value === 'boolean'
    case 'integer':
      return typeof value === 'number' && Number.isInteger(value)
    case 'number':
      // decimal.Decimal money columns serialise as strings.
      if (typeof value === 'number') return true
      return typeof value === 'string' && value.trim() !== '' && !Number.isNaN(Number(value))
    case 'array': return Array.isArray(value)
    case 'object': return typeof value === 'object' && !Array.isArray(value)
    default: return true
  }
}

function actualType(value) {
  if (value === null) return 'null'
  if (Array.isArray(value)) return 'array'
  return typeof value
}

export function conforms(spec, schema, value, path = '') {
  const result = { ok: true, missing: [], extra: [], typeErrors: [] }
  const resolved = deref(spec, schema)
  if (!resolved) return result

  const at = key => (path ? `${path}.${key}` : key)

  if (resolved.type === 'array') {
    if (!Array.isArray(value)) {
      result.typeErrors.push({ path: path || '(root)', expected: 'array', actual: actualType(value) })
      result.ok = false
      return result
    }
    value.forEach((item, i) => {
      const inner = conforms(spec, resolved.items, item, `${path}[${i}]`)
      result.missing.push(...inner.missing)
      result.extra.push(...inner.extra)
      result.typeErrors.push(...inner.typeErrors)
    })
    result.ok = result.missing.length === 0 && result.extra.length === 0 && result.typeErrors.length === 0
    return result
  }

  if (!resolved.properties) return result
  if (value === null || typeof value !== 'object') {
    result.typeErrors.push({ path: path || '(root)', expected: 'object', actual: actualType(value) })
    result.ok = false
    return result
  }

  for (const key of resolved.required ?? []) {
    if (!(key in value)) result.missing.push(at(key))
  }

  for (const key of Object.keys(value)) {
    if (!(key in resolved.properties)) result.extra.push(at(key))
  }

  for (const [key, propSchema] of Object.entries(resolved.properties)) {
    if (!(key in value)) continue
    const prop = deref(spec, propSchema)
    if (!prop) continue
    if (prop.properties || prop.type === 'array') {
      const inner = conforms(spec, prop, value[key], at(key))
      result.missing.push(...inner.missing)
      result.extra.push(...inner.extra)
      result.typeErrors.push(...inner.typeErrors)
      continue
    }
    if (prop.type && !typeOk(prop.type, value[key])) {
      result.typeErrors.push({ path: at(key), expected: prop.type, actual: actualType(value[key]) })
    }
  }

  result.ok = result.missing.length === 0 && result.extra.length === 0 && result.typeErrors.length === 0
  return result
}
