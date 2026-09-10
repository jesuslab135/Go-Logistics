
const SECRET_KEYS = new Set([
  'authorization', 'password', 'access_token', 'refresh_token', 'token',
])

export function redact(value) {
  if (Array.isArray(value)) return value.map(redact)
  if (value === null || typeof value !== 'object') return value
  const out = {}
  for (const [k, v] of Object.entries(value)) {
    out[k] = SECRET_KEYS.has(k.toLowerCase()) ? '[REDACTED]' : redact(v)
  }
  return out
}

export function createRecorder() {
  const log = []
  return {
    record(entry) { log.push(redact(entry)) },
    entries() { return log },
    forOp(opKey) { return log.filter(e => e.opKey === opKey) },
  }
}
