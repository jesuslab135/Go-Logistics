import { readFileSync } from 'node:fs'

const VERBS = ['get', 'post', 'put', 'delete', 'patch']

export function loadSpec(specPath) {
  return JSON.parse(readFileSync(specPath, 'utf8'))
}

export function resolveRef(spec, ref) {
  if (typeof ref !== 'string' || !ref.startsWith('#/')) return null
  return ref
    .slice(2)
    .split('/')
    .reduce((node, key) => (node == null ? null : node[key]), spec)
}

// Swagger 2.0 puts the body schema on a parameter with in: 'body'.
function bodySchema(spec, params) {
  const body = (params ?? []).find(p => p.in === 'body')
  if (!body?.schema) return null
  return body.schema.$ref ? resolveRef(spec, body.schema.$ref) : body.schema
}

// The documented success is the lowest 2xx the operation declares.
function successOf(spec, responses) {
  const codes = Object.keys(responses ?? {})
    .filter(c => /^2\d\d$/.test(c))
    .map(Number)
    .sort((a, b) => a - b)
  if (codes.length === 0) return { successStatus: null, successSchema: null }
  const status = codes[0]
  const schema = responses[String(status)].schema
  return {
    successStatus: status,
    successSchema: schema?.$ref ? resolveRef(spec, schema.$ref) : (schema ?? null),
  }
}

export function listOperations(spec) {
  const ops = []
  for (const [path, item] of Object.entries(spec.paths)) {
    for (const method of VERBS) {
      const op = item[method]
      if (!op) continue
      const params = [...(item.parameters ?? []), ...(op.parameters ?? [])]
      ops.push({
        opKey: `${method.toUpperCase()} ${path}`,
        method: method.toUpperCase(),
        path,
        tags: op.tags ?? [],
        params,
        requestSchema: bodySchema(spec, params),
        ...successOf(spec, op.responses),
      })
    }
  }
  return ops
}
