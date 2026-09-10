# Full-stack Route Audit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Verify all 58 deployed frontend routes against all 374 backend operations, and produce one documented Markdown file listing every frontend and backend fix the audit finds.

**Architecture:** A dependency-free Node harness (`qa/route-audit/`) drives the live API from the local OpenAPI spec, building a fixture graph in foreign-key order before sweeping operations. A separate claude-in-chrome walkthrough records every request the real frontend emits across its 58 routes. A cross-reference step joins the two and classifies mismatches. All three feed one report generator.

**Tech Stack:** Node 25 (native `fetch`, `node:test`, ESM `.mjs`), zero runtime dependencies. claude-in-chrome for the browser layer. No build step.

**Spec:** `docs/superpowers/specs/2026-09-09-fullstack-route-audit-design.md`

## Global Constraints

- **Node ≥ 22** required (native `fetch` and `node:test`). Verified present: v25.2.1.
- **Zero runtime dependencies.** No `npm install`, no `package.json` dependencies block. If a task seems to need a library, it does not.
- **ESM only.** All files use the `.mjs` extension and `import`/`export`.
- **Credentials come from environment variables only.** `AUDIT_EMAIL`, `AUDIT_PASSWORD`, `AUDIT_BASE_URL`. Never hardcode, never commit, never write them into `report.json` or the Markdown report. The recorder must redact the `Authorization` header and any `password` field before an entry is persisted.
- **Base URL:** `https://go-logistics.jesuslab135.com` (default when `AUDIT_BASE_URL` is unset).
- **Every created row is prefixed `ZZ-TEST-<runId>-`** where `runId` is passed in, never generated from a clock inside a pure function.
- **The harness mutates only rows it created.** Existing company-1 data is read, never written.
- **Report output:** `docs/2026-09-09_fullstack-route-audit.md`.
- **Harness location:** `qa/route-audit/`, committed.
- **List envelope:** `{ data: [], total: int, limit: int, offset: int, has_next: bool }`.
- **Error envelope:** `{ error: { code: string, message: string, details?: any } }`.
- **Access token lifetime is 900 seconds.** Any run longer than that must refresh. Treat an unexpected 401 as "refresh and retry once" before recording it as a finding.

---

## File Structure

| File | Responsibility |
|---|---|
| `qa/route-audit/config.mjs` | Environment config and defaults. Single source of base URL and credentials. |
| `qa/route-audit/lib/spec.mjs` | Load `docs/swagger.json`, enumerate operations, resolve `$ref`s. |
| `qa/route-audit/lib/recorder.mjs` | Append-only request/response log with secret redaction. |
| `qa/route-audit/lib/client.mjs` | `ApiClient`: login, auto-refresh, request wrapper feeding the recorder. |
| `qa/route-audit/lib/conform.mjs` | Pure schema-conformance comparison. No I/O. |
| `qa/route-audit/lib/checks.mjs` | The nine per-operation checks, each returning a `CheckResult`. |
| `qa/route-audit/lib/fixtures.mjs` | Build and tear down the fixture graph in dependency order. |
| `qa/route-audit/lib/sweep.mjs` | Run the checks across every operation, module by module. |
| `qa/route-audit/lib/workflows.mjs` | The six stateful suites. |
| `qa/route-audit/lib/isolation.mjs` | Second-tenant isolation suite. |
| `qa/route-audit/lib/crossref.mjs` | Join captured frontend requests to backend operations. |
| `qa/route-audit/lib/report.mjs` | Render the Markdown report from results. |
| `qa/route-audit/run.mjs` | Orchestrator CLI. |
| `qa/route-audit/ui/routes.json` | The canonical 58 frontend routes. |
| `qa/route-audit/ui/walkthrough.md` | The browser protocol a human or agent follows. |
| `qa/route-audit/ui/captures/` | Per-route capture JSON written during the walkthrough. |
| `qa/route-audit/test/*.test.mjs` | Harness unit tests (`node --test`). |
| `qa/route-audit/README.md` | How to run it, and the data-safety contract. |

Tests target the harness's own pure logic — conformance comparison, spec enumeration, cross-referencing, report rendering. Network-driving code is exercised against the live API by the run itself, not mocked.

---

### Task 1: Config and spec loader

**Files:**
- Create: `qa/route-audit/config.mjs`
- Create: `qa/route-audit/lib/spec.mjs`
- Test: `qa/route-audit/test/spec.test.mjs`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `config` — `{ baseUrl: string, email: string, password: string, specPath: string }`
  - `loadSpec(specPath: string) -> Spec` where `Spec = { paths, definitions }`
  - `resolveRef(spec: Spec, ref: string) -> object`
  - `listOperations(spec: Spec) -> Operation[]` where
    `Operation = { opKey: string, method: string, path: string, tags: string[], params: object[], successStatus: number|null, successSchema: object|null, requestSchema: object|null }`
    and `opKey` is `` `${method.toUpperCase()} ${path}` ``.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/spec.test.mjs
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/spec.test.mjs`
Expected: FAIL — `Cannot find module '../lib/spec.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/config.mjs
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const here = dirname(fileURLToPath(import.meta.url))

export const config = {
  baseUrl: process.env.AUDIT_BASE_URL ?? 'https://go-logistics.jesuslab135.com',
  email: process.env.AUDIT_EMAIL ?? '',
  password: process.env.AUDIT_PASSWORD ?? '',
  specPath: resolve(here, '../../docs/swagger.json'),
  captureDir: resolve(here, 'ui/captures'),
  routesPath: resolve(here, 'ui/routes.json'),
}

export function requireCredentials() {
  if (!config.email || !config.password) {
    throw new Error(
      'AUDIT_EMAIL and AUDIT_PASSWORD must be set. ' +
      'They are read from the environment so they never enter the repository.'
    )
  }
}
```

```javascript
// qa/route-audit/lib/spec.mjs
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/spec.test.mjs`
Expected: PASS, 4 tests.

If the count assertion fails because the spec was regenerated, update the `374` literal to the new count and note the change in the commit message — the number is a tripwire against silent spec drift, not a magic constant.

- [ ] **Step 5: Commit**

```bash
git add qa/route-audit/config.mjs qa/route-audit/lib/spec.mjs qa/route-audit/test/spec.test.mjs
git commit -m "qa: spec loader for the route audit harness"
```

---

### Task 2: Recorder with secret redaction

**Files:**
- Create: `qa/route-audit/lib/recorder.mjs`
- Test: `qa/route-audit/test/recorder.test.mjs`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `createRecorder() -> Recorder`
  - `Recorder = { record(entry: Entry): void, entries(): Entry[], forOp(opKey: string): Entry[] }`
  - `Entry = { opKey, method, url, requestBody, requestHeaders, status, responseBody, durationMs }`
  - `redact(value: any) -> any` — exported for direct testing.

Redaction is its own unit because a leaked bearer token in a committed report is a credential disclosure, and the report is the artifact most likely to be shared.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/recorder.test.mjs
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/recorder.test.mjs`
Expected: FAIL — `Cannot find module '../lib/recorder.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/lib/recorder.mjs

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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/recorder.test.mjs`
Expected: PASS, 6 tests.

- [ ] **Step 5: Commit**

```bash
git add qa/route-audit/lib/recorder.mjs qa/route-audit/test/recorder.test.mjs
git commit -m "qa: request recorder with secret redaction"
```

---

### Task 3: API client with auto-refresh

**Files:**
- Create: `qa/route-audit/lib/client.mjs`
- Test: `qa/route-audit/test/client.test.mjs`

**Interfaces:**
- Consumes: `createRecorder` (Task 2), `config` (Task 1).
- Produces:
  - `class ApiClient`
    - `constructor({ baseUrl, recorder })`
    - `async login(email: string, password: string): Promise<void>` — stores both tokens
    - `async request(method, path, opts): Promise<Response>` where
      `opts = { body?, query?, opKey?, anonymous?, retryOn401? }` and
      `Response = { status: number, body: any, headers: object, durationMs: number }`
    - `async refresh(): Promise<void>`
    - `get token(): string|null`
    - `setToken(token: string|null): void` — used by the isolation suite to swap identities
  - `buildUrl(baseUrl, path, query) -> string` — exported pure helper.

`request` never throws on a non-2xx. Status is data here; an audit that throws on 500 cannot report the 500.

Token expiry is 900s, so `request` transparently refreshes once on an unexpected 401 and retries. `anonymous: true` suppresses the header entirely — that is how the unauthenticated check is expressed — and disables the retry.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/client.test.mjs
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/client.test.mjs`
Expected: FAIL — `Cannot find module '../lib/client.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/lib/client.mjs

export function buildUrl(baseUrl, path, query) {
  const base = baseUrl.replace(/\/+$/, '')
  let url = `${base}${path}`
  if (query) {
    const params = new URLSearchParams()
    for (const [k, v] of Object.entries(query)) {
      if (v === undefined || v === null) continue
      params.append(k, String(v))
    }
    const qs = params.toString()
    if (qs) url += `?${qs}`
  }
  return url
}

export class ApiClient {
  constructor({ baseUrl, recorder }) {
    this.baseUrl = baseUrl
    this.recorder = recorder
    this.accessToken = null
    this.refreshToken = null
  }

  get token() { return this.accessToken }

  setToken(token) { this.accessToken = token }

  async login(email, password) {
    const res = await this.request('POST', '/auth/login', {
      body: { email, password },
      anonymous: true,
      opKey: 'POST /auth/login',
    })
    if (res.status !== 200) {
      throw new Error(`login failed: ${res.status} ${JSON.stringify(res.body)}`)
    }
    this.accessToken = res.body.access_token
    this.refreshToken = res.body.refresh_token
  }

  async refresh() {
    const res = await this.request('POST', '/auth/refresh', {
      body: { refresh_token: this.refreshToken },
      anonymous: true,
      opKey: 'POST /auth/refresh',
    })
    if (res.status !== 200) {
      throw new Error(`refresh failed: ${res.status} ${JSON.stringify(res.body)}`)
    }
    this.accessToken = res.body.access_token
    if (res.body.refresh_token) this.refreshToken = res.body.refresh_token
  }

  async request(method, path, opts = {}) {
    const { body, query, opKey, anonymous = false, retryOn401 = true } = opts
    const res = await this.#send(method, path, { body, query, anonymous })

    // The access token lives 900s. An unexpected 401 on an authenticated call
    // is far more often expiry than a real authorization defect, so refresh
    // once and retry before letting it be recorded as a finding.
    if (res.status === 401 && !anonymous && retryOn401 && this.refreshToken) {
      await this.refresh()
      const retried = await this.#send(method, path, { body, query, anonymous })
      this.#log(opKey, method, path, body, retried)
      return retried
    }

    this.#log(opKey, method, path, body, res)
    return res
  }

  async #send(method, path, { body, query, anonymous }) {
    const url = buildUrl(this.baseUrl, path, query)
    const headers = {}
    if (body !== undefined) headers['Content-Type'] = 'application/json'
    if (!anonymous && this.accessToken) {
      headers.Authorization = `Bearer ${this.accessToken}`
    }

    const started = process.hrtime.bigint()
    const response = await fetch(url, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    })
    const durationMs = Number(process.hrtime.bigint() - started) / 1e6

    const text = await response.text()
    let parsed = null
    if (text) {
      try { parsed = JSON.parse(text) } catch { parsed = text }
    }

    return {
      status: response.status,
      body: parsed,
      headers: Object.fromEntries(response.headers.entries()),
      durationMs,
    }
  }

  #log(opKey, method, path, requestBody, res) {
    this.recorder.record({
      opKey: opKey ?? `${method} ${path}`,
      method,
      url: path,
      requestBody: requestBody ?? null,
      status: res.status,
      responseBody: res.body,
      durationMs: res.durationMs,
    })
  }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/client.test.mjs`
Expected: PASS, 4 tests.

- [ ] **Step 5: Verify the client against the live API**

Run:

```bash
AUDIT_EMAIL='...' AUDIT_PASSWORD='...' node --input-type=module -e "
import { ApiClient } from './qa/route-audit/lib/client.mjs'
import { createRecorder } from './qa/route-audit/lib/recorder.mjs'
import { config, requireCredentials } from './qa/route-audit/config.mjs'
requireCredentials()
const c = new ApiClient({ baseUrl: config.baseUrl, recorder: createRecorder() })
await c.login(config.email, config.password)
const r = await c.request('GET', '/api/v1/vendors', { query: { limit: 1 } })
console.log(r.status, Object.keys(r.body))
"
```

Expected: `200 [ 'data', 'total', 'limit', 'offset', 'has_next' ]`

- [ ] **Step 6: Commit**

```bash
git add qa/route-audit/lib/client.mjs qa/route-audit/test/client.test.mjs
git commit -m "qa: API client with token auto-refresh"
```

---

### Task 4: Schema conformance comparison

**Files:**
- Create: `qa/route-audit/lib/conform.mjs`
- Test: `qa/route-audit/test/conform.test.mjs`

**Interfaces:**
- Consumes: `resolveRef` (Task 1).
- Produces:
  - `conforms(spec, schema, value, path?) -> Conformance`
  - `Conformance = { ok: boolean, missing: string[], extra: string[], typeErrors: {path, expected, actual}[] }`

`missing` lists documented-required keys absent from the response. `extra` lists response keys the schema never declares — undocumented fields are a real contract defect, because a generated client will not carry them.

Type checking is deliberately shallow-but-honest: `integer` accepts a JS number with no fractional part; `number` accepts numeric strings, because the backend serialises `decimal.Decimal` money columns as strings (`"parts_subtotal": "100"`) and flagging every one of those would bury the report in noise.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/conform.test.mjs
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/conform.test.mjs`
Expected: FAIL — `Cannot find module '../lib/conform.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/lib/conform.mjs
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/conform.test.mjs`
Expected: PASS, 8 tests.

- [ ] **Step 5: Commit**

```bash
git add qa/route-audit/lib/conform.mjs qa/route-audit/test/conform.test.mjs
git commit -m "qa: schema conformance comparison"
```

---

### Task 5: The per-operation checks

**Files:**
- Create: `qa/route-audit/lib/checks.mjs`
- Test: `qa/route-audit/test/checks.test.mjs`

**Interfaces:**
- Consumes: `ApiClient` (Task 3), `conforms` (Task 4), `Operation` (Task 1).
- Produces:
  - `CheckResult = { check: string, opKey: string, ok: boolean, expected: string, actual: string, severity: 'high'|'medium'|'low'|'info', evidence: any }`
  - `async checkUnauthenticated(client, op) -> CheckResult`
  - `async checkNotFound(client, op) -> CheckResult|null`
  - `async checkHappyPath(client, spec, op, resolvePath) -> CheckResult`
  - `async checkConformance(spec, op, response) -> CheckResult`
  - `async checkListContract(client, op) -> CheckResult|null`
  - `async checkRetiredVerb(client, op) -> CheckResult|null`
  - `fillPath(path: string, ids: object) -> string` — pure, exported for testing.
  - `MISSING_ID` — the constant `999999999` used for the not-found probe.

`fillPath` substitutes `{id}` and `{child_id}` placeholders. Its own test matters because a wrong substitution silently turns a 404 check into a 200 against a real row.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/checks.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { fillPath, MISSING_ID, classify } from '../lib/checks.mjs'

test('fillPath substitutes a single id', () => {
  assert.equal(fillPath('/api/v1/vendors/{id}', { id: 7 }), '/api/v1/vendors/7')
})

test('fillPath substitutes parent and child ids', () => {
  assert.equal(
    fillPath('/api/v1/work-orders/{id}/line-items/{child_id}', { id: 3, child_id: 9 }),
    '/api/v1/work-orders/3/line-items/9'
  )
})

test('fillPath leaves an unknown placeholder in place so it fails loudly', () => {
  assert.equal(fillPath('/api/v1/x/{other}', { id: 1 }), '/api/v1/x/{other}')
})

test('MISSING_ID is far outside any real sequence', () => {
  assert.ok(MISSING_ID > 100000000)
})

test('classify marks an auth gap as high severity', () => {
  assert.equal(classify('unauthenticated', false), 'high')
  assert.equal(classify('tenant-isolation', false), 'high')
})

test('classify marks an undocumented-field drift as low severity', () => {
  assert.equal(classify('conformance', false), 'low')
})

test('classify returns info for a passing check', () => {
  assert.equal(classify('unauthenticated', true), 'info')
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/checks.test.mjs`
Expected: FAIL — `Cannot find module '../lib/checks.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/lib/checks.mjs
import { conforms } from './conform.mjs'

export const MISSING_ID = 999999999

const SEVERITY = {
  'unauthenticated': 'high',
  'tenant-isolation': 'high',
  'module-gate': 'high',
  'happy-path': 'high',
  'retired-verb': 'medium',
  'not-found': 'medium',
  'list-contract': 'medium',
  'validation': 'medium',
  'conformance': 'low',
}

export function classify(check, ok) {
  return ok ? 'info' : (SEVERITY[check] ?? 'medium')
}

export function fillPath(path, ids) {
  return path.replace(/\{(\w+)\}/g, (match, key) =>
    key in ids ? String(ids[key]) : match
  )
}

function result(check, opKey, ok, expected, actual, evidence) {
  return { check, opKey, ok, expected, actual, severity: classify(check, ok), evidence }
}

export async function checkUnauthenticated(client, op) {
  const path = fillPath(op.path, { id: MISSING_ID, child_id: MISSING_ID })
  const res = await client.request(op.method, path, {
    anonymous: true,
    opKey: op.opKey,
    retryOn401: false,
    body: op.requestSchema ? {} : undefined,
  })
  const ok = res.status === 401
  return result('unauthenticated', op.opKey, ok, '401', String(res.status), res.body)
}

export async function checkNotFound(client, op) {
  if (!op.path.includes('{id}')) return null
  if (op.method === 'POST' && !op.path.endsWith('}')) {
    // Action routes such as /purchase-orders/{id}/approve are covered by the
    // workflow suites, which know the legal states. A bare 404 probe here
    // would report a state error as a missing row.
    return null
  }
  const path = fillPath(op.path, { id: MISSING_ID, child_id: MISSING_ID })
  const res = await client.request(op.method, path, {
    opKey: op.opKey,
    body: op.requestSchema ? {} : undefined,
  })
  const ok = res.status === 404
  return result('not-found', op.opKey, ok, '404', String(res.status), res.body)
}

export async function checkHappyPath(client, op, ids, body) {
  const path = fillPath(op.path, ids)
  const res = await client.request(op.method, path, { opKey: op.opKey, body })
  const expected = op.successStatus ?? 200
  const ok = res.status === expected
  return {
    ...result('happy-path', op.opKey, ok, String(expected), String(res.status), res.body),
    response: res,
  }
}

export function checkConformance(spec, op, response) {
  if (!op.successSchema) return null
  if (response.status < 200 || response.status >= 300) return null
  const c = conforms(spec, op.successSchema, response.body)
  const parts = []
  if (c.missing.length) parts.push(`missing: ${c.missing.join(', ')}`)
  if (c.extra.length) parts.push(`undocumented: ${c.extra.join(', ')}`)
  if (c.typeErrors.length) {
    parts.push(`type: ${c.typeErrors.map(e => `${e.path} expected ${e.expected} got ${e.actual}`).join('; ')}`)
  }
  return result(
    'conformance', op.opKey, c.ok,
    'response matches the documented schema',
    parts.length ? parts.join(' | ') : 'matches',
    { missing: c.missing, extra: c.extra, typeErrors: c.typeErrors }
  )
}

export async function checkListContract(client, op) {
  const isList = op.method === 'GET'
    && !op.path.endsWith('}')
    && (op.params ?? []).some(p => p.name === 'limit')
  if (!isList) return null

  const path = fillPath(op.path, { id: MISSING_ID, child_id: MISSING_ID })
  if (path.includes('{')) return null

  const res = await client.request(op.method, path, { opKey: op.opKey, query: { limit: 1, offset: 0 } })
  if (res.status !== 200) {
    return result('list-contract', op.opKey, false, '200 with a page envelope', String(res.status), res.body)
  }
  const b = res.body ?? {}
  const problems = []
  if (!Array.isArray(b.data)) problems.push('data is not an array')
  if (typeof b.total !== 'number') problems.push('total is not a number')
  if (typeof b.has_next !== 'boolean') problems.push('has_next is not a boolean')
  if (Array.isArray(b.data) && b.data.length > 1) problems.push(`limit=1 returned ${b.data.length} rows`)
  const ok = problems.length === 0
  return result('list-contract', op.opKey, ok, 'limit honoured, envelope complete',
    ok ? 'correct' : problems.join('; '), { total: b.total, limit: b.limit, returned: b.data?.length })
}

export async function checkRetiredVerb(client, op) {
  const retired = op.path === '/api/v1/inventory-journal-entries/{id}'
    && (op.method === 'PUT' || op.method === 'DELETE')
  if (!retired) return null
  const path = fillPath(op.path, { id: MISSING_ID })
  const res = await client.request(op.method, path, { opKey: op.opKey, body: op.requestSchema ? {} : undefined })
  const ok = res.status === 405 && typeof res.body?.error?.message === 'string'
  return result('retired-verb', op.opKey, ok, '405 naming the replacement route',
    `${res.status} ${res.body?.error?.message ?? ''}`.trim(), res.body)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/checks.test.mjs`
Expected: PASS, 7 tests.

- [ ] **Step 5: Commit**

```bash
git add qa/route-audit/lib/checks.mjs qa/route-audit/test/checks.test.mjs
git commit -m "qa: per-operation contract checks"
```

---

### Task 6: Fixture graph

**Files:**
- Create: `qa/route-audit/lib/fixtures.mjs`
- Test: `qa/route-audit/test/fixtures.test.mjs`

**Interfaces:**
- Consumes: `ApiClient` (Task 3).
- Produces:
  - `FIXTURE_PLAN: FixtureStep[]` where
    `FixtureStep = { key: string, path: string, body: (ids, tag) => object, dependsOn: string[] }`
  - `async buildFixtures(client, runId) -> FixtureGraph` where
    `FixtureGraph = { runId, tag: string, ids: Record<string, number>, created: {key, path, id}[], failed: {key, status, body}[] }`
  - `async teardownFixtures(client, graph) -> { deleted: string[], archived: string[], failed: {key, status, body}[] }`
  - `topoSort(plan: FixtureStep[]) -> FixtureStep[]` — pure, exported for testing.

`tag` is `` `ZZ-TEST-${runId}` `` and every created row's name carries it.

Creation order is a topological sort of `dependsOn`, not a hand-maintained list — a hand-maintained list drifts the moment a step is inserted. Teardown walks `created` in reverse. A delete that fails with a foreign-key conflict falls back to the resource's `archive` action where one exists, which is the system's own answer to retiring a referenced row.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/fixtures.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { topoSort, FIXTURE_PLAN } from '../lib/fixtures.mjs'

test('topoSort places dependencies before dependents', () => {
  const plan = [
    { key: 'c', dependsOn: ['b'] },
    { key: 'a', dependsOn: [] },
    { key: 'b', dependsOn: ['a'] },
  ]
  assert.deepEqual(topoSort(plan).map(s => s.key), ['a', 'b', 'c'])
})

test('topoSort throws on a cycle rather than looping forever', () => {
  const plan = [
    { key: 'a', dependsOn: ['b'] },
    { key: 'b', dependsOn: ['a'] },
  ]
  assert.throws(() => topoSort(plan), /cycle/i)
})

test('topoSort throws when a dependency is not in the plan', () => {
  const plan = [{ key: 'a', dependsOn: ['ghost'] }]
  assert.throws(() => topoSort(plan), /unknown dependency/i)
})

test('every FIXTURE_PLAN dependency refers to a real step', () => {
  const keys = new Set(FIXTURE_PLAN.map(s => s.key))
  for (const step of FIXTURE_PLAN) {
    for (const dep of step.dependsOn) {
      assert.ok(keys.has(dep), `${step.key} depends on unknown ${dep}`)
    }
  }
})

test('FIXTURE_PLAN sorts without a cycle', () => {
  assert.equal(topoSort(FIXTURE_PLAN).length, FIXTURE_PLAN.length)
})

test('every fixture body carries the run tag so teardown can find it', () => {
  const ids = Object.fromEntries(FIXTURE_PLAN.map(s => [s.key, 1]))
  for (const step of FIXTURE_PLAN) {
    const body = step.body(ids, 'ZZ-TEST-42')
    const serialised = JSON.stringify(body)
    assert.ok(
      serialised.includes('ZZ-TEST-42'),
      `${step.key} body does not carry the run tag: ${serialised}`
    )
  }
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/fixtures.test.mjs`
Expected: FAIL — `Cannot find module '../lib/fixtures.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/lib/fixtures.mjs

// Resources that answer POST .../{id}/archive when a delete is refused.
const ARCHIVABLE = new Set(['vendor', 'part', 'asset', 'serviceTask', 'inspectionForm'])

export const FIXTURE_PLAN = [
  // Level 0 — depend on nothing but the company.
  { key: 'vendor', path: '/api/v1/vendors', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-vendor`, external_id: `${tag}-v` }) },
  { key: 'location', path: '/api/v1/locations', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-location` }) },
  { key: 'assetType', path: '/api/v1/asset-types', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-asset-type` }) },
  { key: 'assetStatus', path: '/api/v1/asset-statuses', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-asset-status` }) },
  { key: 'partCategory', path: '/api/v1/part-categories', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-part-category` }) },
  { key: 'partManufacturer', path: '/api/v1/part-manufacturers', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-part-manufacturer` }) },
  { key: 'measurementUnit', path: '/api/v1/measurement-units', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-unit`, abbreviation: 'ZZ' }) },
  { key: 'partLocation', path: '/api/v1/part-locations', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-part-location` }) },
  { key: 'adjustmentReason', path: '/api/v1/inventory-adjustment-reasons', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-adjustment-reason` }) },
  { key: 'fuelType', path: '/api/v1/fuel-types', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-fuel-type` }) },
  { key: 'trailerClassification', path: '/api/v1/trailer-classifications', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-trailer-class` }) },
  { key: 'issuePriority', path: '/api/v1/issue-priorities', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-priority` }) },
  { key: 'workOrderStatus', path: '/api/v1/work-order-statuses', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-wo-status` }) },
  { key: 'vehicleMake', path: '/api/v1/vehicle-makes', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-make` }) },
  { key: 'axleTemplate', path: '/api/v1/axle-templates', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-axle-template` }) },
  { key: 'serviceTask', path: '/api/v1/service-tasks', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-service-task` }) },
  { key: 'role', path: '/api/v1/roles', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-role`, is_admin: false }) },
  { key: 'group', path: '/api/v1/groups', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-group` }) },
  { key: 'fault', path: '/api/v1/faults', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-fault` }) },
  { key: 'inspectionForm', path: '/api/v1/inspection-forms', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-inspection-form` }) },
  { key: 'tireModel', path: '/api/v1/tire-models', dependsOn: [],
    body: (_, tag) => ({ name: `${tag}-tire-model` }) },

  // Level 1 — one hop.
  { key: 'vehicleModel', path: '/api/v1/vehicle-models', dependsOn: ['vehicleMake'],
    body: (ids, tag) => ({ name: `${tag}-model`, make_id: ids.vehicleMake }) },
  { key: 'part', path: '/api/v1/parts', dependsOn: ['partCategory', 'partManufacturer', 'measurementUnit'],
    body: (ids, tag) => ({
      name: `${tag}-part`, part_number: `${tag}-PN`,
      category_id: ids.partCategory, manufacturer_id: ids.partManufacturer,
      measurement_unit_id: ids.measurementUnit,
    }) },
  { key: 'asset', path: '/api/v1/assets', dependsOn: ['assetType', 'assetStatus'],
    body: (ids, tag) => ({
      name: `${tag}-asset`, vin_sn: `${tag}-VIN`,
      asset_type_id: ids.assetType, status_id: ids.assetStatus,
      vehicle_type: 'VEHICLE', ownership_type: 'OWNED',
    }) },
  { key: 'trailerAsset', path: '/api/v1/assets', dependsOn: ['assetType', 'assetStatus'],
    body: (ids, tag) => ({
      name: `${tag}-trailer-asset`, vin_sn: `${tag}-TVIN`,
      asset_type_id: ids.assetType, status_id: ids.assetStatus,
      vehicle_type: 'TRAILER', ownership_type: 'OWNED',
    }) },
  { key: 'tire', path: '/api/v1/tires', dependsOn: ['tireModel'],
    body: (ids, tag) => ({ serial_number: `${tag}-tire`, tire_model_id: ids.tireModel }) },
  { key: 'axleDefinition', path: '/api/v1/axle-templates/{axleTemplate}/definitions', dependsOn: ['axleTemplate'],
    body: (_, tag) => ({ name: `${tag}-axle-def`, position: 1, wheel_count: 2 }) },
  { key: 'employee', path: '/api/v1/employees', dependsOn: ['role', 'group'],
    body: (ids, tag) => ({
      first_name: 'ZZTEST', last_name: `${tag}-employee`,
      email: `${tag.toLowerCase()}-employee@example.invalid`,
      role_id: ids.role, group_id: ids.group, is_active: true,
    }) },

  // Level 2 — two hops.
  { key: 'workOrder', path: '/api/v1/work-orders', dependsOn: ['asset', 'workOrderStatus'],
    body: (ids, tag) => ({
      number: `${tag}-WO`, description: `${tag} work order`,
      asset_id: ids.asset, status_id: ids.workOrderStatus,
      issued_at: '2026-09-09T00:00:00Z',
    }) },
  { key: 'issue', path: '/api/v1/issues', dependsOn: ['asset', 'issuePriority'],
    body: (ids, tag) => ({
      summary: `${tag}-issue`, asset_id: ids.asset, priority_id: ids.issuePriority,
    }) },
  { key: 'purchaseOrder', path: '/api/v1/purchase-orders', dependsOn: ['vendor'],
    body: (ids, tag) => ({ number: `${tag}-PO`, vendor_id: ids.vendor }) },
  { key: 'serviceEntry', path: '/api/v1/service-entries', dependsOn: ['asset', 'vendor'],
    body: (ids, tag) => ({
      asset_id: ids.asset, vendor_id: ids.vendor,
      reference: `${tag}-SE`, started_at: '2026-09-09T00:00:00Z',
    }) },
  { key: 'fuelEntry', path: '/api/v1/assets/{asset}/fuel-entries', dependsOn: ['asset', 'fuelType', 'vendor'],
    body: (ids, tag) => ({
      fuel_type_id: ids.fuelType, vendor_id: ids.vendor, reference: `${tag}-FE`,
      filled_at: '2026-09-09T00:00:00Z', volume: '10', total_cost: '100',
    }) },
  { key: 'partInventory', path: '/api/v1/parts/{part}/inventory', dependsOn: ['part', 'partLocation'],
    body: (ids, tag) => ({ location_id: ids.partLocation, quantity: '10', notes: `${tag}-inv` }) },
  { key: 'warranty', path: '/api/v1/warranties', dependsOn: ['asset'],
    body: (ids, tag) => ({ name: `${tag}-warranty`, asset_id: ids.asset }) },
  { key: 'weeklyMileageGoal', path: '/api/v1/weekly-mileage-goals', dependsOn: ['asset'],
    body: (ids, tag) => ({ asset_id: ids.asset, target_miles: '100', notes: `${tag}-goal` }) },
  { key: 'trailerAssignment', path: '/api/v1/assets/{asset}/trailer-assignments', dependsOn: ['asset', 'trailerAsset'],
    body: (ids, tag) => ({ trailer_id: ids.trailerAsset, notes: `${tag}-assignment` }) },
  { key: 'inspectionFormItem', path: '/api/v1/inspection-forms/{inspectionForm}/items', dependsOn: ['inspectionForm'],
    body: (_, tag) => ({ label: `${tag}-form-item`, position: 1, item_type: 'PASS_FAIL' }) },

  // Level 3 — line items and logs.
  { key: 'workOrderLineItem', path: '/api/v1/work-orders/{workOrder}/line-items', dependsOn: ['workOrder', 'serviceTask'],
    body: (ids, tag) => ({ service_task_id: ids.serviceTask, description: `${tag}-wo-line` }) },
  { key: 'purchaseOrderLineItem', path: '/api/v1/purchase-orders/{purchaseOrder}/line-items', dependsOn: ['purchaseOrder', 'part'],
    body: (ids, tag) => ({ part_id: ids.part, quantity: '2', unit_cost: '5', description: `${tag}-po-line` }) },
  { key: 'serviceEntryLineItem', path: '/api/v1/service-entries/{serviceEntry}/line-items', dependsOn: ['serviceEntry', 'serviceTask'],
    body: (ids, tag) => ({ service_task_id: ids.serviceTask, description: `${tag}-se-line` }) },
  { key: 'wheelPosition', path: '/api/v1/axle-definitions/{axleDefinition}/wheel-positions', dependsOn: ['axleDefinition'],
    body: (_, tag) => ({ label: `${tag}-wheel`, position: 1 }) },
  { key: 'journalEntry', path: '/api/v1/inventory-journal-entries', dependsOn: ['part', 'partLocation', 'adjustmentReason'],
    body: (ids, tag) => ({
      part_id: ids.part, location_id: ids.partLocation,
      reason_id: ids.adjustmentReason, quantity_delta: '5', notes: `${tag}-journal`,
    }) },

  // Level 4 — grandchildren.
  { key: 'workOrderSubLineItem', path: '/api/v1/work-order-line-items/{workOrderLineItem}/sub-line-items', dependsOn: ['workOrderLineItem', 'part'],
    body: (ids, tag) => ({ part_id: ids.part, quantity: '1', unit_cost: '3', description: `${tag}-sub-line` }) },
  { key: 'laborEntry', path: '/api/v1/work-order-sub-line-items/{workOrderSubLineItem}/labor-entries', dependsOn: ['workOrderSubLineItem', 'employee'],
    body: (ids, tag) => ({ employee_id: ids.employee, hours: '1', notes: `${tag}-labor` }) },
]

export function topoSort(plan) {
  const byKey = new Map(plan.map(s => [s.key, s]))
  for (const step of plan) {
    for (const dep of step.dependsOn) {
      if (!byKey.has(dep)) {
        throw new Error(`${step.key} has unknown dependency ${dep}`)
      }
    }
  }

  const sorted = []
  const state = new Map() // key -> 'visiting' | 'done'

  const visit = (step, trail) => {
    const mark = state.get(step.key)
    if (mark === 'done') return
    if (mark === 'visiting') {
      throw new Error(`cycle in fixture plan: ${[...trail, step.key].join(' -> ')}`)
    }
    state.set(step.key, 'visiting')
    for (const dep of step.dependsOn) visit(byKey.get(dep), [...trail, step.key])
    state.set(step.key, 'done')
    sorted.push(step)
  }

  for (const step of plan) visit(step, [])
  return sorted
}

// Fixture paths reference earlier fixtures by key, e.g.
// /api/v1/assets/{asset}/fuel-entries. Placeholders resolve from ids.
function resolveFixturePath(path, ids) {
  return path.replace(/\{(\w+)\}/g, (match, key) =>
    key in ids ? String(ids[key]) : match
  )
}

export async function buildFixtures(client, runId) {
  const tag = `ZZ-TEST-${runId}`
  const ids = {}
  const created = []
  const failed = []

  for (const step of topoSort(FIXTURE_PLAN)) {
    const missing = step.dependsOn.filter(d => !(d in ids))
    if (missing.length) {
      failed.push({ key: step.key, status: null, body: `skipped: unmet dependencies ${missing.join(', ')}` })
      continue
    }
    const path = resolveFixturePath(step.path, ids)
    const res = await client.request('POST', path, {
      body: step.body(ids, tag),
      opKey: `POST ${step.path} (fixture)`,
    })
    if (res.status >= 200 && res.status < 300 && res.body?.id != null) {
      ids[step.key] = res.body.id
      created.push({ key: step.key, path, id: res.body.id })
    } else {
      failed.push({ key: step.key, status: res.status, body: res.body })
    }
  }

  return { runId, tag, ids, created, failed }
}

export async function teardownFixtures(client, graph) {
  const deleted = []
  const archived = []
  const failed = []

  for (const row of [...graph.created].reverse()) {
    const target = `${row.path}/${row.id}`
    const res = await client.request('DELETE', target, { opKey: `DELETE ${row.key} (teardown)` })
    if (res.status >= 200 && res.status < 300) {
      deleted.push(`${row.key}#${row.id}`)
      continue
    }
    if (ARCHIVABLE.has(row.key)) {
      const arc = await client.request('POST', `${target}/archive`, { opKey: `ARCHIVE ${row.key} (teardown)` })
      if (arc.status >= 200 && arc.status < 300) {
        archived.push(`${row.key}#${row.id}`)
        continue
      }
    }
    failed.push({ key: `${row.key}#${row.id}`, status: res.status, body: res.body })
  }

  return { deleted, archived, failed }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/fixtures.test.mjs`
Expected: PASS, 6 tests.

- [ ] **Step 5: Dry-run the fixture graph against the live API**

Run:

```bash
AUDIT_EMAIL='...' AUDIT_PASSWORD='...' node --input-type=module -e "
import { ApiClient } from './qa/route-audit/lib/client.mjs'
import { createRecorder } from './qa/route-audit/lib/recorder.mjs'
import { buildFixtures, teardownFixtures } from './qa/route-audit/lib/fixtures.mjs'
import { config, requireCredentials } from './qa/route-audit/config.mjs'
requireCredentials()
const c = new ApiClient({ baseUrl: config.baseUrl, recorder: createRecorder() })
await c.login(config.email, config.password)
const g = await buildFixtures(c, 'dryrun')
console.log('created', g.created.length, 'failed', g.failed.length)
console.log(JSON.stringify(g.failed, null, 1))
const t = await teardownFixtures(c, g)
console.log('deleted', t.deleted.length, 'archived', t.archived.length, 'failed', t.failed.length)
"
```

Expected: every step creates. **Any entry in `failed` is itself a finding** — record the exact status and body, because a fixture that cannot be created means either the plan's payload is wrong or the endpoint is broken, and both belong in the report. Fix payload errors here; carry endpoint errors forward as findings rather than papering over them.

- [ ] **Step 6: Commit**

```bash
git add qa/route-audit/lib/fixtures.mjs qa/route-audit/test/fixtures.test.mjs
git commit -m "qa: fixture dependency graph with reverse-order teardown"
```

---

### Task 7: Sweep runner

**Files:**
- Create: `qa/route-audit/lib/sweep.mjs`
- Test: `qa/route-audit/test/sweep.test.mjs`

**Interfaces:**
- Consumes: `listOperations` (Task 1), all checks (Task 5), `FixtureGraph` (Task 6).
- Produces:
  - `async sweep(client, spec, graph, { onProgress }) -> SweepResult`
  - `SweepResult = { results: CheckResult[], byOp: Map<string, CheckResult[]>, coverage: {opKey, tested: boolean, reason: string|null}[] }`
  - `idsForOperation(op, graph) -> object|null` — pure, exported. Maps an operation path to the fixture ids that satisfy its placeholders, or `null` when no fixture fits.

`idsForOperation` is the load-bearing pure function. It maps `/api/v1/work-orders/{id}/line-items/{child_id}` to `{ id: graph.ids.workOrder, child_id: graph.ids.workOrderLineItem }`. When it returns `null`, the operation is recorded in `coverage` as untested with a reason rather than silently skipped — the spec requires every operation to carry a verdict or an explicit reason.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/sweep.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { idsForOperation, PATH_FIXTURE_MAP } from '../lib/sweep.mjs'

const graph = { ids: { workOrder: 10, workOrderLineItem: 20, vendor: 30, asset: 40 } }

test('idsForOperation maps a top-level detail path to its fixture', () => {
  const op = { path: '/api/v1/vendors/{id}' }
  assert.deepEqual(idsForOperation(op, graph), { id: 30 })
})

test('idsForOperation maps a nested path to parent and child fixtures', () => {
  const op = { path: '/api/v1/work-orders/{id}/line-items/{child_id}' }
  assert.deepEqual(idsForOperation(op, graph), { id: 10, child_id: 20 })
})

test('idsForOperation returns an empty object for a collection path', () => {
  const op = { path: '/api/v1/vendors' }
  assert.deepEqual(idsForOperation(op, graph), {})
})

test('idsForOperation returns null when no fixture matches the collection', () => {
  const op = { path: '/api/v1/unknown-thing/{id}' }
  assert.equal(idsForOperation(op, graph), null)
})

test('idsForOperation returns null when the mapped fixture failed to build', () => {
  const op = { path: '/api/v1/assets/{id}/fuel-entries/{child_id}' }
  assert.equal(idsForOperation(op, { ids: { asset: 40 } }), null)
})

test('every PATH_FIXTURE_MAP value names a key the fixture plan can produce', async () => {
  const { FIXTURE_PLAN } = await import('../lib/fixtures.mjs')
  const keys = new Set(FIXTURE_PLAN.map(s => s.key))
  for (const [collection, fixtureKey] of Object.entries(PATH_FIXTURE_MAP)) {
    assert.ok(keys.has(fixtureKey), `${collection} maps to unknown fixture ${fixtureKey}`)
  }
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/sweep.test.mjs`
Expected: FAIL — `Cannot find module '../lib/sweep.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/lib/sweep.mjs
import {
  checkUnauthenticated, checkNotFound, checkHappyPath,
  checkConformance, checkListContract, checkRetiredVerb,
} from './checks.mjs'

// Collection segment -> fixture key. The sweep resolves {id} from the segment
// immediately preceding it, and {child_id} from the segment after it.
export const PATH_FIXTURE_MAP = {
  'vendors': 'vendor',
  'locations': 'location',
  'asset-types': 'assetType',
  'asset-statuses': 'assetStatus',
  'part-categories': 'partCategory',
  'part-manufacturers': 'partManufacturer',
  'measurement-units': 'measurementUnit',
  'part-locations': 'partLocation',
  'inventory-adjustment-reasons': 'adjustmentReason',
  'fuel-types': 'fuelType',
  'trailer-classifications': 'trailerClassification',
  'issue-priorities': 'issuePriority',
  'work-order-statuses': 'workOrderStatus',
  'vehicle-makes': 'vehicleMake',
  'vehicle-models': 'vehicleModel',
  'axle-templates': 'axleTemplate',
  'axle-definitions': 'axleDefinition',
  'service-tasks': 'serviceTask',
  'roles': 'role',
  'groups': 'group',
  'faults': 'fault',
  'inspection-forms': 'inspectionForm',
  'tire-models': 'tireModel',
  'tires': 'tire',
  'parts': 'part',
  'assets': 'asset',
  'employees': 'employee',
  'work-orders': 'workOrder',
  'work-order-line-items': 'workOrderLineItem',
  'work-order-sub-line-items': 'workOrderSubLineItem',
  'issues': 'issue',
  'purchase-orders': 'purchaseOrder',
  'service-entries': 'serviceEntry',
  'service-entry-line-items': 'serviceEntryLineItem',
  'warranties': 'warranty',
  'weekly-mileage-goals': 'weeklyMileageGoal',
  'inventory-journal-entries': 'journalEntry',
  'line-items': 'workOrderLineItem',
  'sub-line-items': 'workOrderSubLineItem',
  'labor-entries': 'laborEntry',
  'wheel-positions': 'wheelPosition',
  'definitions': 'axleDefinition',
  'inventory': 'partInventory',
  'fuel-entries': 'fuelEntry',
  'trailer-assignments': 'trailerAssignment',
  'items': 'inspectionFormItem',
}

export function idsForOperation(op, graph) {
  const segments = op.path.split('/').filter(Boolean)
  const ids = {}

  for (let i = 0; i < segments.length; i++) {
    const seg = segments[i]
    if (seg !== '{id}' && seg !== '{child_id}') continue

    // {id} belongs to the collection before it; {child_id} to the collection
    // that follows {id}.
    const collection = seg === '{id}'
      ? segments[i - 1]
      : segments[i - 1]
    const fixtureKey = PATH_FIXTURE_MAP[collection]
    if (!fixtureKey) return null
    const value = graph.ids[fixtureKey]
    if (value == null) return null
    ids[seg === '{id}' ? 'id' : 'child_id'] = value
  }

  return ids
}

function bodyForOperation(op, graph) {
  if (!op.requestSchema) return undefined
  // Update verbs replay the fixture's own shape where one exists; the sweep
  // deliberately does not invent field values it cannot justify.
  return { }
}

export async function sweep(client, spec, ops, graph, { onProgress } = {}) {
  const results = []
  const coverage = []
  const byOp = new Map()

  const push = r => {
    if (!r) return
    results.push(r)
    if (!byOp.has(r.opKey)) byOp.set(r.opKey, [])
    byOp.get(r.opKey).push(r)
  }

  for (const op of ops) {
    onProgress?.(op)

    push(await checkUnauthenticated(client, op))
    push(await checkRetiredVerb(client, op))
    push(await checkNotFound(client, op))
    push(await checkListContract(client, op))

    const ids = idsForOperation(op, graph)
    if (ids === null) {
      coverage.push({
        opKey: op.opKey,
        tested: false,
        reason: 'no fixture row satisfies this path',
      })
      continue
    }

    // Destructive verbs are exercised by teardown and the workflow suites,
    // never by the generic sweep: a DELETE here would dissolve the fixture
    // graph the remaining operations depend on.
    if (op.method === 'DELETE') {
      coverage.push({ opKey: op.opKey, tested: false, reason: 'covered by teardown' })
      continue
    }

    const happy = await checkHappyPath(client, op, ids, bodyForOperation(op, graph))
    push(happy)
    push(checkConformance(spec, op, happy.response))
    coverage.push({ opKey: op.opKey, tested: true, reason: null })
  }

  return { results, byOp, coverage }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/sweep.test.mjs`
Expected: PASS, 6 tests.

- [ ] **Step 5: Commit**

```bash
git add qa/route-audit/lib/sweep.mjs qa/route-audit/test/sweep.test.mjs
git commit -m "qa: operation sweep runner with explicit coverage accounting"
```

---

### Task 8: Module gate suite

**Files:**
- Create: `qa/route-audit/lib/gates.mjs`
- Test: `qa/route-audit/test/gates.test.mjs`

**Interfaces:**
- Consumes: `ApiClient` (Task 3), `FixtureGraph` (Task 6).
- Produces:
  - `MODULE_PROBES: {module: string, probe: {method, path}}[]` — one representative operation per module.
  - `async provisionLimitedIdentity(client, graph, password) -> {email, employeeId, roleId}`
  - `async runGateSuite(client, graph, limited) -> CheckResult[]`

The fixture plan already creates a role with `is_admin: false` and an employee bound to it. This task gives that employee a password via `POST /api/v1/employees/{id}/set-password`, logs in as them on a second client, and confirms every module answers 403.

This is the highest-value suite in the audit. `RequireModule`, `RequireAction` and `RequirePlatformAdmin` are the difference between a bug and a breach, and nothing else in the plan tests them.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/gates.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { MODULE_PROBES } from '../lib/gates.mjs'

test('every module named by the permissions endpoint has a probe', () => {
  const modules = [
    'assets', 'company', 'employees', 'fuel', 'inspections', 'inventory',
    'issues', 'mileage_goals', 'parts', 'purchase_orders', 'roles', 'service',
    'tire_approvals', 'tires', 'vendors', 'warranties', 'work_orders',
  ]
  const covered = new Set(MODULE_PROBES.map(p => p.module))
  for (const m of modules) {
    assert.ok(covered.has(m), `no gate probe for module ${m}`)
  }
})

test('every probe targets a read verb, so a 403 cannot be confused with a write failure', () => {
  for (const p of MODULE_PROBES) {
    assert.equal(p.probe.method, 'GET', `${p.module} probe should be a GET`)
  }
})

test('every probe path is a collection, needing no fixture id', () => {
  for (const p of MODULE_PROBES) {
    assert.ok(!p.probe.path.includes('{'), `${p.module} probe path has a placeholder`)
  }
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/gates.test.mjs`
Expected: FAIL — `Cannot find module '../lib/gates.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/lib/gates.mjs
import { ApiClient } from './client.mjs'
import { classify } from './checks.mjs'

export const MODULE_PROBES = [
  { module: 'assets', probe: { method: 'GET', path: '/api/v1/assets' } },
  { module: 'company', probe: { method: 'GET', path: '/api/v1/companies' } },
  { module: 'employees', probe: { method: 'GET', path: '/api/v1/employees' } },
  { module: 'fuel', probe: { method: 'GET', path: '/api/v1/fuel-entries' } },
  { module: 'inspections', probe: { method: 'GET', path: '/api/v1/inspection-forms' } },
  { module: 'inventory', probe: { method: 'GET', path: '/api/v1/measurement-units' } },
  { module: 'issues', probe: { method: 'GET', path: '/api/v1/issues' } },
  { module: 'mileage_goals', probe: { method: 'GET', path: '/api/v1/weekly-mileage-goals' } },
  { module: 'parts', probe: { method: 'GET', path: '/api/v1/parts' } },
  { module: 'purchase_orders', probe: { method: 'GET', path: '/api/v1/purchase-orders' } },
  { module: 'roles', probe: { method: 'GET', path: '/api/v1/roles' } },
  { module: 'service', probe: { method: 'GET', path: '/api/v1/service-tasks' } },
  { module: 'tire_approvals', probe: { method: 'GET', path: '/api/v1/tire-assignment-requests' } },
  { module: 'tires', probe: { method: 'GET', path: '/api/v1/tires' } },
  { module: 'vendors', probe: { method: 'GET', path: '/api/v1/vendors' } },
  { module: 'warranties', probe: { method: 'GET', path: '/api/v1/warranties' } },
  { module: 'work_orders', probe: { method: 'GET', path: '/api/v1/work-orders' } },
]

// Routes that must refuse a non-platform-admin outright.
const PLATFORM_ADMIN_PROBES = [
  { method: 'GET', path: '/api/v1/admin/employees' },
  { method: 'GET', path: '/api/v1/admin/companies/1/roles' },
  { method: 'GET', path: '/api/v1/admin/companies/1/owner' },
  { method: 'GET', path: '/api/v1/admin/companies/1/work-order-statuses' },
]

export async function provisionLimitedIdentity(client, graph, password) {
  const employeeId = graph.ids.employee
  if (!employeeId) throw new Error('fixture employee was not created; gate suite cannot run')

  const res = await client.request('POST', `/api/v1/employees/${employeeId}/set-password`, {
    body: { password },
    opKey: 'POST /api/v1/employees/{id}/set-password',
  })
  if (res.status < 200 || res.status >= 300) {
    throw new Error(`set-password failed: ${res.status} ${JSON.stringify(res.body)}`)
  }

  return {
    email: `${graph.tag.toLowerCase()}-employee@example.invalid`,
    employeeId,
    roleId: graph.ids.role,
  }
}

function gateResult(check, opKey, ok, expected, actual, evidence) {
  return { check, opKey, ok, expected, actual, severity: classify(check, ok), evidence }
}

export async function runGateSuite(client, graph, limited, password) {
  const results = []

  const limitedClient = new ApiClient({ baseUrl: client.baseUrl, recorder: client.recorder })
  await limitedClient.login(limited.email, password)

  // The fixture role grants no modules, so every module probe must refuse.
  for (const { module, probe } of MODULE_PROBES) {
    const res = await limitedClient.request(probe.method, probe.path, {
      opKey: `${probe.method} ${probe.path} (module-gate:${module})`,
    })
    const ok = res.status === 403
    results.push(gateResult(
      'module-gate',
      `${probe.method} ${probe.path}`,
      ok,
      `403 for a role without the ${module} module`,
      String(res.status),
      { module, body: res.body }
    ))
  }

  // A tenant admin is not a platform operator. These must refuse too.
  for (const probe of PLATFORM_ADMIN_PROBES) {
    const res = await limitedClient.request(probe.method, probe.path, {
      opKey: `${probe.method} ${probe.path} (platform-admin-gate)`,
    })
    const ok = res.status === 403
    results.push(gateResult(
      'module-gate',
      `${probe.method} ${probe.path}`,
      ok,
      '403 for a non-platform-admin',
      String(res.status),
      res.body
    ))
  }

  // /me/permissions is gated on identity alone: a client must be able to
  // discover that it may do nothing, without first being refused.
  const me = await limitedClient.request('GET', '/api/v1/me/permissions', {
    opKey: 'GET /api/v1/me/permissions (limited identity)',
  })
  results.push(gateResult(
    'module-gate',
    'GET /api/v1/me/permissions',
    me.status === 200,
    '200 even for a role with no modules',
    String(me.status),
    me.body
  ))

  return results
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/gates.test.mjs`
Expected: PASS, 3 tests.

- [ ] **Step 5: Commit**

```bash
git add qa/route-audit/lib/gates.mjs qa/route-audit/test/gates.test.mjs
git commit -m "qa: module and platform-admin gate suite"
```

---

### Task 9: Workflow suites

**Files:**
- Create: `qa/route-audit/lib/workflows.mjs`
- Test: `qa/route-audit/test/workflows.test.mjs`

**Interfaces:**
- Consumes: `ApiClient` (Task 3), `FixtureGraph` (Task 6).
- Produces:
  - `PO_TRANSITIONS: {action: string, legalFrom: string[]}[]`
  - `async runPurchaseOrderWorkflow(client, graph) -> CheckResult[]`
  - `async runWorkOrderWorkflow(client, graph) -> CheckResult[]`
  - `async runJournalWorkflow(client, graph) -> CheckResult[]`
  - `async runArchiveWorkflow(client, graph) -> CheckResult[]`
  - `async runUploadWorkflow(client, graph) -> CheckResult[]`
  - `async runWorkflows(client, graph) -> CheckResult[]` — runs all five and concatenates.

The transition table below was taken from `internal/domain/purchaseorder/workflow.go` at plan time, not guessed. It has **eight** actions, and two details that a five-action guess would miss: rejection is not terminal (it returns to `DRAFT` through an explicit `revise`, so the order keeps its line items), and `reject` carries `RequiresReason`, so a rejection with no reason must be refused.

- [ ] **Step 1: Confirm the transition table still matches the source**

Run: `sed -n '25,66p' internal/domain/purchaseorder/workflow.go`

Expected — eight action constants and eight transitions:

| Action | From | To | Approval | Reason |
|---|---|---|---|---|
| `submit` | `DRAFT`, `REJECTED` | `PENDING_APPROVAL` | | |
| `approve` | `PENDING_APPROVAL` | `APPROVED` | yes | |
| `reject` | `PENDING_APPROVAL` | `REJECTED` | yes | yes |
| `revise` | `REJECTED` | `DRAFT` | | |
| `purchase` | `APPROVED` | `PURCHASED` | | |
| `receive-partial` | `PURCHASED` | `RECEIVED_PARTIAL` | | |
| `receive-full` | `PURCHASED`, `RECEIVED_PARTIAL` | `RECEIVED_FULL` | | |
| `close` | `RECEIVED_FULL` | `CLOSED` | | |

If the source has moved on, correct the literal in Step 4 and say so in the commit message. The test in Step 2 enforces the agreement.

- [ ] **Step 2: Write the failing test**

```javascript
// qa/route-audit/test/workflows.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { PO_TRANSITIONS, PO_WALK } from '../lib/workflows.mjs'

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
```

- [ ] **Step 3: Run test to verify it fails**

Run: `node --test qa/route-audit/test/workflows.test.mjs`
Expected: FAIL — `Cannot find module '../lib/workflows.mjs'`

- [ ] **Step 4: Write the implementation**

Populate `PO_TRANSITIONS` from what Step 1 read. The structure below is the shape; the action names and states come from the source.

```javascript
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
  const override = await client.request('POST', `/api/v1/purchase-orders/${poId}/override-total`, {
    body: { total: '123.45', reason: `${graph.tag} override` },
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
  // company already has two.
  const second = await client.request('POST', '/api/v1/work-order-statuses', {
    body: { name: `${graph.tag}-wo-status-2` },
    opKey: 'POST /api/v1/work-order-statuses (workflow)',
  })
  const secondId = second.body?.id

  if (secondId) {
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

  const reversal = await client.request('POST', `/api/v1/inventory-journal-entries/${entryId}/reverse`, {
    body: { notes: `${graph.tag} reversal` },
    opKey: 'POST /api/v1/inventory-journal-entries/{id}/reverse',
  })
  results.push(wf('workflow', 'POST /api/v1/inventory-journal-entries/{id}/reverse',
    reversal.status >= 200 && reversal.status < 300,
    '2xx creating a balancing entry',
    String(reversal.status), reversal.body))

  // Reversing a reversal must be refused, or the ledger can be spun forever.
  if (reversal.body?.id) {
    const again = await client.request('POST', `/api/v1/inventory-journal-entries/${reversal.body.id}/reverse`, {
      body: { notes: `${graph.tag} double reversal` },
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
  form.append('purpose', 'asset_photo')

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
  const stored = body?.url ?? body?.location ?? null
  if (typeof stored === 'string' && stored.includes('/fleet-private/')) {
    const bare = stored.split('?')[0]
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
```

- [ ] **Step 5: Run test to verify it passes**

Run: `node --test qa/route-audit/test/workflows.test.mjs`
Expected: PASS, 7 tests. If a test fails, `PO_TRANSITIONS` disagrees with `workflow.go` — fix the table, not the test.

- [ ] **Step 6: Commit**

```bash
git add qa/route-audit/lib/workflows.mjs qa/route-audit/test/workflows.test.mjs
git commit -m "qa: stateful workflow suites"
```

---

### Task 10: Tenant isolation suite

**Files:**
- Create: `qa/route-audit/lib/isolation.mjs`
- Test: `qa/route-audit/test/isolation.test.mjs`

**Interfaces:**
- Consumes: `ApiClient` (Task 3), `FixtureGraph` (Task 6).
- Produces:
  - `ISOLATION_TARGETS: {key: string, path: string}[]` — fixture rows in company 1 that company 2 must not reach.
  - `async createSecondTenant(client, tag) -> {companyId: number}`
  - `async runIsolationSuite(client, graph) -> CheckResult[]`

The flow: `POST /api/v1/companies` (account-owner right), `POST /auth/switch-company` to move the token into it, then attempt to read each company-1 fixture row by id. Anything other than 403 or 404 is a cross-tenant leak and the most serious class of finding available.

The suite switches back to company 1 before returning, so teardown runs in the right tenant.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/isolation.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { ISOLATION_TARGETS, verdictFor } from '../lib/isolation.mjs'
import { FIXTURE_PLAN } from '../lib/fixtures.mjs'

test('every isolation target names a real fixture key', () => {
  const keys = new Set(FIXTURE_PLAN.map(s => s.key))
  for (const t of ISOLATION_TARGETS) {
    assert.ok(keys.has(t.key), `${t.key} is not a fixture`)
  }
})

test('verdictFor accepts 403 and 404 as correct refusals', () => {
  assert.equal(verdictFor(403).ok, true)
  assert.equal(verdictFor(404).ok, true)
})

test('verdictFor treats a 200 as a leak at high severity', () => {
  const v = verdictFor(200)
  assert.equal(v.ok, false)
  assert.equal(v.severity, 'high')
})

test('verdictFor treats a 500 as a failure but not a leak', () => {
  const v = verdictFor(500)
  assert.equal(v.ok, false)
  assert.equal(v.severity, 'medium')
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/isolation.test.mjs`
Expected: FAIL — `Cannot find module '../lib/isolation.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/lib/isolation.mjs

export const ISOLATION_TARGETS = [
  { key: 'vendor', path: '/api/v1/vendors' },
  { key: 'asset', path: '/api/v1/assets' },
  { key: 'part', path: '/api/v1/parts' },
  { key: 'workOrder', path: '/api/v1/work-orders' },
  { key: 'purchaseOrder', path: '/api/v1/purchase-orders' },
  { key: 'issue', path: '/api/v1/issues' },
  { key: 'serviceEntry', path: '/api/v1/service-entries' },
  { key: 'employee', path: '/api/v1/employees' },
  { key: 'journalEntry', path: '/api/v1/inventory-journal-entries' },
  { key: 'tire', path: '/api/v1/tires' },
]

export function verdictFor(status) {
  if (status === 403 || status === 404) {
    return { ok: true, severity: 'info' }
  }
  if (status >= 200 && status < 300) {
    return { ok: false, severity: 'high' }
  }
  return { ok: false, severity: 'medium' }
}

export async function createSecondTenant(client, tag) {
  const res = await client.request('POST', '/api/v1/companies', {
    body: { name: `${tag} Isolation Co` },
    opKey: 'POST /api/v1/companies',
  })
  if (res.status < 200 || res.status >= 300 || res.body?.id == null) {
    throw new Error(`could not create the second tenant: ${res.status} ${JSON.stringify(res.body)}`)
  }
  return { companyId: res.body.id }
}

async function switchTo(client, companyId) {
  const res = await client.request('POST', '/auth/switch-company', {
    body: { company_id: companyId },
    opKey: 'POST /auth/switch-company',
  })
  if (res.status >= 200 && res.status < 300 && res.body?.access_token) {
    client.setToken(res.body.access_token)
    if (res.body.refresh_token) client.refreshToken = res.body.refresh_token
  }
  return res
}

export async function runIsolationSuite(client, graph, homeCompanyId) {
  const results = []
  const { companyId } = await createSecondTenant(client, graph.tag)

  const switched = await switchTo(client, companyId)
  results.push({
    check: 'tenant-isolation',
    opKey: 'POST /auth/switch-company',
    ok: switched.status >= 200 && switched.status < 300,
    expected: '2xx returning a token scoped to the new company',
    actual: String(switched.status),
    severity: switched.status >= 200 && switched.status < 300 ? 'info' : 'high',
    evidence: { companyId },
  })

  if (switched.status >= 200 && switched.status < 300) {
    for (const target of ISOLATION_TARGETS) {
      const id = graph.ids[target.key]
      if (!id) continue
      const res = await client.request('GET', `${target.path}/${id}`, {
        opKey: `GET ${target.path}/{id} (cross-tenant)`,
      })
      const v = verdictFor(res.status)
      results.push({
        check: 'tenant-isolation',
        opKey: `GET ${target.path}/{id}`,
        ok: v.ok,
        expected: "403 or 404 reading another company's row",
        actual: String(res.status),
        severity: v.severity,
        evidence: { targetKey: target.key, id, body: res.body },
      })
    }

    // A list in the new tenant must be empty of company-1 rows.
    const vendors = await client.request('GET', '/api/v1/vendors', {
      query: { limit: 200 }, opKey: 'GET /api/v1/vendors (cross-tenant list)',
    })
    const leaked = (vendors.body?.data ?? []).some(r => r.id === graph.ids.vendor)
    results.push({
      check: 'tenant-isolation',
      opKey: 'GET /api/v1/vendors (list)',
      ok: !leaked,
      expected: "the other tenant's vendor is absent from this company's list",
      actual: leaked ? "the other tenant's vendor is listed" : 'absent',
      severity: leaked ? 'high' : 'info',
      evidence: { total: vendors.body?.total },
    })
  }

  // Return to the home tenant so teardown deletes in the right company.
  await switchTo(client, homeCompanyId)

  const cleanup = await client.request('DELETE', `/api/v1/companies/${companyId}`, {
    opKey: 'DELETE /api/v1/companies/{id} (isolation teardown)',
  })
  results.push({
    check: 'tenant-isolation',
    opKey: 'DELETE /api/v1/companies/{id}',
    ok: cleanup.status >= 200 && cleanup.status < 300,
    expected: '2xx removing the throwaway tenant',
    actual: String(cleanup.status),
    severity: cleanup.status >= 200 && cleanup.status < 300 ? 'info' : 'medium',
    evidence: { companyId, body: cleanup.body },
  })

  return results
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/isolation.test.mjs`
Expected: PASS, 4 tests.

- [ ] **Step 5: Commit**

```bash
git add qa/route-audit/lib/isolation.mjs qa/route-audit/test/isolation.test.mjs
git commit -m "qa: cross-tenant isolation suite"
```

---

### Task 11: Frontend route inventory and walkthrough protocol

**Files:**
- Create: `qa/route-audit/ui/routes.json`
- Create: `qa/route-audit/ui/walkthrough.md`
- Create: `qa/route-audit/ui/captures/.gitkeep`
- Test: `qa/route-audit/test/routes.test.mjs`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `routes.json` — `{ routes: Route[] }` where
    `Route = { path: string, kind: 'public'|'list'|'detail'|'form'|'settings'|'admin', needsFixture: string|null }`
  - `capture` file format, written by the walkthrough into `ui/captures/<slug>.json`:
    `{ route, loadedAt, httpStatus, requests: [{method, url, status, requestBody, responseSummary}], consoleErrors: [string], rendered: 'data'|'empty'|'error'|'blank', notes: string }`

`needsFixture` names the fixture key supplying a real id for a `$id` route, so the walkthrough is not left guessing which row to open.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/routes.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { config } from '../config.mjs'

const { routes } = JSON.parse(readFileSync(config.routesPath, 'utf8'))

test('the inventory holds all 58 routes the design pinned', () => {
  assert.equal(routes.length, 58)
})

test('route paths are unique', () => {
  const seen = new Set(routes.map(r => r.path))
  assert.equal(seen.size, routes.length)
})

test('every dynamic route names the fixture that supplies its id', () => {
  for (const r of routes) {
    if (r.path.includes('$id')) {
      assert.ok(r.needsFixture, `${r.path} does not name a fixture`)
    }
  }
})

test('the three public routes are marked public', () => {
  const publicPaths = routes.filter(r => r.kind === 'public').map(r => r.path).sort()
  assert.deepEqual(publicPaths, ['/login', '/onboarding/company', '/signup'])
})

test('every route path starts with a slash', () => {
  for (const r of routes) assert.match(r.path, /^\//)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/routes.test.mjs`
Expected: FAIL — the routes file does not exist.

- [ ] **Step 3: Write the route inventory**

```json
{
  "routes": [
    { "path": "/login", "kind": "public", "needsFixture": null },
    { "path": "/signup", "kind": "public", "needsFixture": null },
    { "path": "/onboarding/company", "kind": "public", "needsFixture": null },
    { "path": "/app/dashboard", "kind": "list", "needsFixture": null },
    { "path": "/app/design-system", "kind": "list", "needsFixture": null },
    { "path": "/app/vehicles", "kind": "list", "needsFixture": null },
    { "path": "/app/vehicles/new", "kind": "form", "needsFixture": null },
    { "path": "/app/vehicles/$id", "kind": "detail", "needsFixture": "asset" },
    { "path": "/app/vehicles/$id/edit", "kind": "form", "needsFixture": "asset" },
    { "path": "/app/trailers", "kind": "list", "needsFixture": null },
    { "path": "/app/trailers/new", "kind": "form", "needsFixture": null },
    { "path": "/app/trailers/$id", "kind": "detail", "needsFixture": "trailerAsset" },
    { "path": "/app/trailers/$id/edit", "kind": "form", "needsFixture": "trailerAsset" },
    { "path": "/app/tires", "kind": "list", "needsFixture": null },
    { "path": "/app/tires/$id", "kind": "detail", "needsFixture": "tire" },
    { "path": "/app/tires/models", "kind": "list", "needsFixture": null },
    { "path": "/app/tires/axle-templates", "kind": "list", "needsFixture": null },
    { "path": "/app/tires/assignment-requests", "kind": "list", "needsFixture": null },
    { "path": "/app/tires/movements", "kind": "list", "needsFixture": null },
    { "path": "/app/maintenance", "kind": "list", "needsFixture": null },
    { "path": "/app/maintenance/$id", "kind": "detail", "needsFixture": "workOrder" },
    { "path": "/app/service", "kind": "list", "needsFixture": null },
    { "path": "/app/service/entries", "kind": "list", "needsFixture": null },
    { "path": "/app/service/entries/$id", "kind": "detail", "needsFixture": "serviceEntry" },
    { "path": "/app/service/reminders", "kind": "list", "needsFixture": null },
    { "path": "/app/purchase-orders", "kind": "list", "needsFixture": null },
    { "path": "/app/purchase-orders/$id", "kind": "detail", "needsFixture": "purchaseOrder" },
    { "path": "/app/inventory", "kind": "list", "needsFixture": null },
    { "path": "/app/inventory/part-categories", "kind": "list", "needsFixture": null },
    { "path": "/app/inventory/part-manufacturers", "kind": "list", "needsFixture": null },
    { "path": "/app/inventory/part-locations", "kind": "list", "needsFixture": null },
    { "path": "/app/inventory/inventory-adjustment-reasons", "kind": "list", "needsFixture": null },
    { "path": "/app/inventory/inventory-journal-entries", "kind": "list", "needsFixture": null },
    { "path": "/app/issues", "kind": "list", "needsFixture": null },
    { "path": "/app/issues/board", "kind": "list", "needsFixture": null },
    { "path": "/app/issues/faults", "kind": "list", "needsFixture": null },
    { "path": "/app/issues/priorities", "kind": "list", "needsFixture": null },
    { "path": "/app/fuel", "kind": "list", "needsFixture": null },
    { "path": "/app/inspections", "kind": "list", "needsFixture": null },
    { "path": "/app/inspections/forms", "kind": "list", "needsFixture": null },
    { "path": "/app/vendors", "kind": "list", "needsFixture": null },
    { "path": "/app/locations", "kind": "list", "needsFixture": null },
    { "path": "/app/vehicle-makes", "kind": "list", "needsFixture": null },
    { "path": "/app/vehicle-models", "kind": "list", "needsFixture": null },
    { "path": "/app/organization", "kind": "list", "needsFixture": null },
    { "path": "/app/organization/employees", "kind": "list", "needsFixture": null },
    { "path": "/app/organization/groups", "kind": "list", "needsFixture": null },
    { "path": "/app/organization/roles", "kind": "list", "needsFixture": null },
    { "path": "/app/settings", "kind": "settings", "needsFixture": null },
    { "path": "/app/settings/custom-fields", "kind": "settings", "needsFixture": null },
    { "path": "/app/settings/fuel-types", "kind": "settings", "needsFixture": null },
    { "path": "/app/settings/measurement-units", "kind": "settings", "needsFixture": null },
    { "path": "/app/settings/service-tasks", "kind": "settings", "needsFixture": null },
    { "path": "/app/settings/weekly-mileage-goals", "kind": "settings", "needsFixture": null },
    { "path": "/app/admin", "kind": "admin", "needsFixture": null },
    { "path": "/app/admin/companies", "kind": "admin", "needsFixture": null },
    { "path": "/app/admin/companies/$id", "kind": "admin", "needsFixture": null },
    { "path": "/app/admin/users", "kind": "admin", "needsFixture": null }
  ]
}
```

Note: `/app/admin/companies/$id` uses company 1 rather than a fixture, since the audit's throwaway tenant is created and destroyed inside the isolation suite. Its `needsFixture` is `null` for that reason.

- [ ] **Step 4: Write the walkthrough protocol**

```markdown
# Browser walkthrough protocol

Run after the Layer 1 sweep, so fixture ids exist for the `$id` routes.

## Setup

1. Confirm the claude-in-chrome extension has permission for
   `go-logistics.netlify.app`.
2. Open `https://go-logistics.netlify.app/login`.
3. Sign in with `AUDIT_EMAIL` / `AUDIT_PASSWORD`.
4. Open DevTools, Network tab, and enable "Preserve log".

## Per route

For each entry in `routes.json`, in the order listed:

1. Navigate to the route. Substitute `$id` with the fixture id named by
   `needsFixture`, read from the run's `report.json`.
2. Wait for the network to go idle.
3. Record a capture file at `qa/route-audit/ui/captures/<slug>.json`, where
   `<slug>` is the path with slashes and dollars replaced by hyphens:

```json
{
  "route": "/app/vehicles",
  "loadedAt": "2026-09-09T00:00:00Z",
  "httpStatus": 200,
  "requests": [
    {
      "method": "GET",
      "url": "/api/v1/assets?limit=25&offset=0",
      "status": 200,
      "requestBody": null,
      "responseSummary": "12 rows, has_next false"
    }
  ],
  "consoleErrors": [],
  "rendered": "data",
  "notes": ""
}
```

4. `rendered` is one of:
   - `data` — the screen shows real rows or a populated form
   - `empty` — the screen shows a legitimate empty state
   - `error` — the screen shows an error state or a toast
   - `blank` — nothing rendered; usually a crash, and always a finding

5. On a `form` route, submit the form once with valid values, prefixing any
   free-text name with the run tag. Record the resulting request and status.
   Do not submit a second time.

## Rules

- Never record a bearer token or password into a capture file.
- If a route 404s in the frontend router, record `httpStatus: 404` and
  `rendered: "blank"`. A route present in the bundle but not reachable is a
  finding.
- If a screen calls an endpoint that does not exist in `swagger.json`, note it
  in `notes`. The cross-reference step will classify it.
```

- [ ] **Step 5: Run test to verify it passes**

Run: `node --test qa/route-audit/test/routes.test.mjs`
Expected: PASS, 5 tests.

- [ ] **Step 6: Commit**

```bash
git add qa/route-audit/ui/
git commit -m "qa: frontend route inventory and browser walkthrough protocol"
```

---

### Task 12: Cross-reference engine

**Files:**
- Create: `qa/route-audit/lib/crossref.mjs`
- Test: `qa/route-audit/test/crossref.test.mjs`

**Interfaces:**
- Consumes: `listOperations` (Task 1), capture files (Task 11).
- Produces:
  - `normalisePath(url: string) -> string` — strips the origin and query, replaces numeric segments with `{id}`.
  - `matchOperation(ops, method, url) -> Operation|null`
  - `crossReference(spec, ops, captures) -> Finding[]` where
    `Finding = { id: string, layer: 'frontend'|'backend'|'contract', severity, route, summary, expected, actual, evidence }`
  - `unusedOperations(ops, captures) -> string[]`

`normalisePath` is the piece most likely to be quietly wrong, so it gets the most tests. A frontend URL is `/api/v1/work-orders/12/line-items/34`; the operation is `/api/v1/work-orders/{id}/line-items/{child_id}`. Getting the second placeholder wrong silently drops every nested route from the join.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/crossref.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { normalisePath, matchOperation, crossReference, unusedOperations } from '../lib/crossref.mjs'

test('normalisePath strips the origin', () => {
  assert.equal(
    normalisePath('https://go-logistics.jesuslab135.com/api/v1/assets'),
    '/api/v1/assets'
  )
})

test('normalisePath strips the query string', () => {
  assert.equal(normalisePath('/api/v1/assets?limit=25&offset=0'), '/api/v1/assets')
})

test('normalisePath replaces the first numeric segment with {id}', () => {
  assert.equal(normalisePath('/api/v1/assets/12'), '/api/v1/assets/{id}')
})

test('normalisePath replaces the second numeric segment with {child_id}', () => {
  assert.equal(
    normalisePath('/api/v1/work-orders/12/line-items/34'),
    '/api/v1/work-orders/{id}/line-items/{child_id}'
  )
})

test('normalisePath leaves non-numeric segments alone', () => {
  assert.equal(
    normalisePath('/api/v1/purchase-orders/12/approve'),
    '/api/v1/purchase-orders/{id}/approve'
  )
})

test('matchOperation finds the operation for a real frontend URL', () => {
  const ops = [
    { opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' },
    { opKey: 'GET /api/v1/assets/{id}', method: 'GET', path: '/api/v1/assets/{id}' },
  ]
  assert.equal(matchOperation(ops, 'GET', '/api/v1/assets/9').opKey, 'GET /api/v1/assets/{id}')
})

test('matchOperation returns null for an endpoint the backend does not define', () => {
  const ops = [{ opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' }]
  assert.equal(matchOperation(ops, 'GET', '/api/v1/ghosts'), null)
})

test('crossReference flags a frontend call to an undefined endpoint', () => {
  const ops = [{ opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' }]
  const captures = [{
    route: '/app/vehicles',
    rendered: 'error',
    consoleErrors: [],
    requests: [{ method: 'GET', url: '/api/v1/ghosts', status: 404 }],
  }]
  const findings = crossReference({ definitions: {} }, ops, captures)
  assert.equal(findings.length, 1)
  assert.equal(findings[0].layer, 'frontend')
  assert.match(findings[0].summary, /no such backend operation/i)
})

test('crossReference flags a failing request on a rendered screen', () => {
  const ops = [{ opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' }]
  const captures = [{
    route: '/app/vehicles',
    rendered: 'error',
    consoleErrors: [],
    requests: [{ method: 'GET', url: '/api/v1/assets', status: 500 }],
  }]
  const findings = crossReference({ definitions: {} }, ops, captures)
  assert.equal(findings[0].layer, 'backend')
  assert.equal(findings[0].severity, 'high')
})

test('crossReference flags a blank render even when every request succeeded', () => {
  const ops = [{ opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' }]
  const captures = [{
    route: '/app/vehicles',
    rendered: 'blank',
    consoleErrors: ['TypeError: x is undefined'],
    requests: [{ method: 'GET', url: '/api/v1/assets', status: 200 }],
  }]
  const findings = crossReference({ definitions: {} }, ops, captures)
  assert.equal(findings.length, 1)
  assert.equal(findings[0].layer, 'frontend')
})

test('unusedOperations lists backend operations no screen calls', () => {
  const ops = [
    { opKey: 'GET /api/v1/assets', method: 'GET', path: '/api/v1/assets' },
    { opKey: 'GET /api/v1/tires', method: 'GET', path: '/api/v1/tires' },
  ]
  const captures = [{ route: '/app/vehicles', requests: [{ method: 'GET', url: '/api/v1/assets', status: 200 }] }]
  assert.deepEqual(unusedOperations(ops, captures), ['GET /api/v1/tires'])
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/crossref.test.mjs`
Expected: FAIL — `Cannot find module '../lib/crossref.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/lib/crossref.mjs

export function normalisePath(url) {
  let path = url
  const originMatch = /^https?:\/\/[^/]+(\/.*)$/.exec(url)
  if (originMatch) path = originMatch[1]
  path = path.split('?')[0]

  const placeholders = ['{id}', '{child_id}']
  let seen = 0
  return path
    .split('/')
    .map(seg => {
      if (!/^\d+$/.test(seg)) return seg
      const token = placeholders[seen] ?? '{id}'
      seen += 1
      return token
    })
    .join('/')
}

export function matchOperation(ops, method, url) {
  const path = normalisePath(url)
  return ops.find(o => o.method === method.toUpperCase() && o.path === path) ?? null
}

let counter = 0
function nextId(prefix) {
  counter += 1
  return `${prefix}-${String(counter).padStart(3, '0')}`
}

export function resetFindingIds() { counter = 0 }

export function crossReference(spec, ops, captures) {
  const findings = []

  for (const capture of captures) {
    for (const req of capture.requests ?? []) {
      const op = matchOperation(ops, req.method, req.url)

      if (!op) {
        findings.push({
          id: nextId('FE'),
          layer: 'frontend',
          severity: 'high',
          route: capture.route,
          summary: `The screen calls ${req.method} ${normalisePath(req.url)}, for which there is no such backend operation`,
          expected: 'a request to an operation the API defines',
          actual: `${req.method} ${normalisePath(req.url)} -> ${req.status}`,
          evidence: req,
        })
        continue
      }

      if (req.status >= 500) {
        findings.push({
          id: nextId('BE'),
          layer: 'backend',
          severity: 'high',
          route: capture.route,
          summary: `${op.opKey} returned ${req.status} to the frontend`,
          expected: `${op.successStatus ?? '2xx'}`,
          actual: String(req.status),
          evidence: req,
        })
        continue
      }

      if (req.status === 400 || req.status === 422) {
        findings.push({
          id: nextId('CT'),
          layer: 'contract',
          severity: 'high',
          route: capture.route,
          summary: `${op.opKey} rejected the payload the frontend sent`,
          expected: `${op.successStatus ?? '2xx'}`,
          actual: `${req.status} ${JSON.stringify(req.responseSummary ?? '')}`,
          evidence: req,
        })
        continue
      }

      if (req.status === 403 || req.status === 404) {
        findings.push({
          id: nextId('CT'),
          layer: 'contract',
          severity: 'medium',
          route: capture.route,
          summary: `${op.opKey} refused the frontend with ${req.status}`,
          expected: `${op.successStatus ?? '2xx'} for an authorised admin`,
          actual: String(req.status),
          evidence: req,
        })
      }
    }

    const allOk = (capture.requests ?? []).every(r => r.status < 400)
    if (capture.rendered === 'blank' && allOk) {
      findings.push({
        id: nextId('FE'),
        layer: 'frontend',
        severity: 'high',
        route: capture.route,
        summary: 'The screen rendered nothing even though every request succeeded',
        expected: 'rendered data or a deliberate empty state',
        actual: `blank; console: ${(capture.consoleErrors ?? []).join(' | ') || 'no errors logged'}`,
        evidence: { consoleErrors: capture.consoleErrors ?? [] },
      })
    }

    if (capture.rendered === 'error' && allOk) {
      findings.push({
        id: nextId('FE'),
        layer: 'frontend',
        severity: 'medium',
        route: capture.route,
        summary: 'The screen showed an error state even though every request succeeded',
        expected: 'rendered data or a deliberate empty state',
        actual: `error; console: ${(capture.consoleErrors ?? []).join(' | ') || 'no errors logged'}`,
        evidence: { consoleErrors: capture.consoleErrors ?? [] },
      })
    }
  }

  return findings
}

export function unusedOperations(ops, captures) {
  const called = new Set()
  for (const capture of captures) {
    for (const req of capture.requests ?? []) {
      const op = matchOperation(ops, req.method, req.url)
      if (op) called.add(op.opKey)
    }
  }
  return ops.filter(o => !called.has(o.opKey)).map(o => o.opKey)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/crossref.test.mjs`
Expected: PASS, 11 tests.

- [ ] **Step 5: Commit**

```bash
git add qa/route-audit/lib/crossref.mjs qa/route-audit/test/crossref.test.mjs
git commit -m "qa: cross-reference frontend captures against backend operations"
```

---

### Task 13: Report generator

**Files:**
- Create: `qa/route-audit/lib/report.mjs`
- Test: `qa/route-audit/test/report.test.mjs`

**Interfaces:**
- Consumes: `SweepResult` (Task 7), gate results (Task 8), workflow results (Task 9), isolation results (Task 10), findings (Task 12).
- Produces:
  - `renderReport(input) -> string` where
    `input = { runId, baseUrl, generatedAt, operations, sweep, gates, workflows, isolation, findings, captures, coverage, teardown, unused }`
  - `severityHistogram(findings) -> Record<string, number>`
  - `verdictForRoute(capture) -> 'works'|'partial'|'broken'|'not-implemented'`

`generatedAt` is passed in, never read from a clock inside the function, so the renderer stays pure and its output is testable.

- [ ] **Step 1: Write the failing test**

```javascript
// qa/route-audit/test/report.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { renderReport, severityHistogram, verdictForRoute } from '../lib/report.mjs'

test('severityHistogram counts findings by severity', () => {
  const h = severityHistogram([
    { severity: 'high' }, { severity: 'high' }, { severity: 'low' },
  ])
  assert.deepEqual(h, { high: 2, medium: 0, low: 1 })
})

test('verdictForRoute calls a clean data render a pass', () => {
  assert.equal(verdictForRoute({
    rendered: 'data', requests: [{ status: 200 }], consoleErrors: [],
  }), 'works')
})

test('verdictForRoute calls a blank render broken', () => {
  assert.equal(verdictForRoute({
    rendered: 'blank', requests: [{ status: 200 }], consoleErrors: [],
  }), 'broken')
})

test('verdictForRoute calls a data render with a failed request partial', () => {
  assert.equal(verdictForRoute({
    rendered: 'data', requests: [{ status: 200 }, { status: 500 }], consoleErrors: [],
  }), 'partial')
})

test('verdictForRoute calls a 404 route not-implemented', () => {
  assert.equal(verdictForRoute({
    rendered: 'blank', httpStatus: 404, requests: [], consoleErrors: [],
  }), 'not-implemented')
})

test('renderReport produces every required section', () => {
  const md = renderReport({
    runId: 'test', baseUrl: 'https://example.com',
    generatedAt: '2026-09-09T00:00:00Z',
    operations: [{ opKey: 'GET /api/v1/assets', tags: ['assets'] }],
    sweep: { results: [], coverage: [{ opKey: 'GET /api/v1/assets', tested: true, reason: null }] },
    gates: [], workflows: [], isolation: [],
    findings: [], captures: [], teardown: { deleted: [], archived: [], failed: [] },
    unused: [],
  })
  for (const heading of [
    '# Full-stack route audit',
    '## Executive summary',
    '## Frontend route verdicts',
    '## Backend operation verdicts',
    '## Findings',
    '## Untested and blocked',
    '## Teardown',
  ]) {
    assert.ok(md.includes(heading), `missing section: ${heading}`)
  }
})

test('renderReport states plainly when nothing was found', () => {
  const md = renderReport({
    runId: 'test', baseUrl: 'https://example.com',
    generatedAt: '2026-09-09T00:00:00Z',
    operations: [], sweep: { results: [], coverage: [] },
    gates: [], workflows: [], isolation: [],
    findings: [], captures: [], teardown: { deleted: [], archived: [], failed: [] },
    unused: [],
  })
  assert.match(md, /No findings/i)
})

test('renderReport never emits a redacted secret placeholder as a real value', () => {
  const md = renderReport({
    runId: 'test', baseUrl: 'https://example.com',
    generatedAt: '2026-09-09T00:00:00Z',
    operations: [], sweep: { results: [], coverage: [] },
    gates: [], workflows: [], isolation: [],
    findings: [{
      id: 'BE-001', layer: 'backend', severity: 'high', route: '/app/x',
      summary: 'boom', expected: '200', actual: '500',
      evidence: { headers: { Authorization: '[REDACTED]' } },
    }],
    captures: [], teardown: { deleted: [], archived: [], failed: [] },
    unused: [],
  })
  assert.ok(md.includes('[REDACTED]'))
  assert.ok(!md.includes('Bearer '))
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test qa/route-audit/test/report.test.mjs`
Expected: FAIL — `Cannot find module '../lib/report.mjs'`

- [ ] **Step 3: Write the implementation**

```javascript
// qa/route-audit/lib/report.mjs

export function severityHistogram(findings) {
  const h = { high: 0, medium: 0, low: 0 }
  for (const f of findings) {
    if (f.severity in h) h[f.severity] += 1
  }
  return h
}

export function verdictForRoute(capture) {
  if (capture.httpStatus === 404) return 'not-implemented'
  const failed = (capture.requests ?? []).some(r => r.status >= 400)
  if (capture.rendered === 'blank') return 'broken'
  if (capture.rendered === 'error') return failed ? 'broken' : 'partial'
  if (failed) return 'partial'
  if ((capture.consoleErrors ?? []).length > 0) return 'partial'
  return 'works'
}

function table(headers, rows) {
  const head = `| ${headers.join(' | ')} |`
  const rule = `|${headers.map(() => '---').join('|')}|`
  const body = rows.map(r => `| ${r.join(' | ')} |`).join('\n')
  return rows.length ? `${head}\n${rule}\n${body}` : `${head}\n${rule}\n| _none_ |${' |'.repeat(headers.length - 1)}`
}

function fence(value) {
  return '```json\n' + JSON.stringify(value, null, 2) + '\n```'
}

export function renderReport(input) {
  const {
    runId, baseUrl, generatedAt, operations, sweep, gates, workflows,
    isolation, findings, captures, teardown, unused,
  } = input

  const allResults = [
    ...(sweep.results ?? []), ...(gates ?? []),
    ...(workflows ?? []), ...(isolation ?? []),
  ]
  const failed = allResults.filter(r => !r.ok)
  const hist = severityHistogram([...findings, ...failed])
  const tested = (sweep.coverage ?? []).filter(c => c.tested).length
  const untested = (sweep.coverage ?? []).filter(c => !c.tested)

  const out = []

  out.push('# Full-stack route audit')
  out.push('')
  out.push(`**Run:** \`${runId}\`  `)
  out.push(`**Target:** ${baseUrl}  `)
  out.push(`**Generated:** ${generatedAt}`)
  out.push('')

  out.push('## Executive summary')
  out.push('')
  out.push(table(
    ['Metric', 'Value'],
    [
      ['Backend operations in the spec', String(operations.length)],
      ['Operations exercised', String(tested)],
      ['Operations not exercised', String(untested.length)],
      ['Frontend routes walked', String(captures.length)],
      ['Checks run', String(allResults.length)],
      ['Checks failed', String(failed.length)],
      ['Findings — high', String(hist.high)],
      ['Findings — medium', String(hist.medium)],
      ['Findings — low', String(hist.low)],
      ['Backend operations no screen calls', String((unused ?? []).length)],
    ]
  ))
  out.push('')

  out.push('## Frontend route verdicts')
  out.push('')
  out.push(table(
    ['Route', 'Verdict', 'Requests', 'Failed', 'Console errors'],
    captures.map(c => [
      `\`${c.route}\``,
      verdictForRoute(c),
      String((c.requests ?? []).length),
      String((c.requests ?? []).filter(r => r.status >= 400).length),
      String((c.consoleErrors ?? []).length),
    ])
  ))
  out.push('')

  out.push('## Backend operation verdicts')
  out.push('')
  const byOpFailures = new Map()
  for (const r of failed) {
    if (!byOpFailures.has(r.opKey)) byOpFailures.set(r.opKey, [])
    byOpFailures.get(r.opKey).push(r.check)
  }
  out.push(table(
    ['Operation', 'Tested', 'Failed checks'],
    (sweep.coverage ?? []).map(c => [
      `\`${c.opKey}\``,
      c.tested ? 'yes' : `no — ${c.reason}`,
      (byOpFailures.get(c.opKey) ?? []).join(', ') || '—',
    ])
  ))
  out.push('')

  out.push('## Findings')
  out.push('')
  if (findings.length === 0 && failed.length === 0) {
    out.push('No findings. Every check passed and every walked route rendered cleanly.')
    out.push('')
  } else {
    const ordered = [...findings].sort((a, b) => {
      const rank = { high: 0, medium: 1, low: 2 }
      return (rank[a.severity] ?? 3) - (rank[b.severity] ?? 3)
    })
    for (const f of ordered) {
      out.push(`### ${f.id} — ${f.summary}`)
      out.push('')
      out.push(`**Severity:** ${f.severity}  `)
      out.push(`**Layer:** ${f.layer}  `)
      out.push(`**Route:** \`${f.route}\``)
      out.push('')
      out.push(`**Expected:** ${f.expected}`)
      out.push('')
      out.push(`**Observed:** ${f.actual}`)
      out.push('')
      out.push('**Evidence:**')
      out.push('')
      out.push(fence(f.evidence))
      out.push('')
    }

    if (failed.length) {
      out.push('### Failed contract checks')
      out.push('')
      out.push(table(
        ['Operation', 'Check', 'Expected', 'Observed', 'Severity'],
        failed.map(r => [
          `\`${r.opKey}\``, r.check, r.expected,
          String(r.actual).replace(/\|/g, '\\|').slice(0, 160), r.severity,
        ])
      ))
      out.push('')
    }
  }

  out.push('## Untested and blocked')
  out.push('')
  out.push('Operations that carry no verdict, and why. This section exists so the')
  out.push('report admits its gaps rather than implying coverage it does not have.')
  out.push('')
  out.push(table(
    ['Operation', 'Reason'],
    untested.map(c => [`\`${c.opKey}\``, c.reason])
  ))
  out.push('')

  if ((unused ?? []).length) {
    out.push('### Backend operations the frontend never calls')
    out.push('')
    out.push('These are proven by Layer 1 but unreachable through the UI. That is')
    out.push('not automatically a defect — it may be unbuilt frontend, or a route')
    out.push('with no screen behind it.')
    out.push('')
    for (const opKey of unused) out.push(`- \`${opKey}\``)
    out.push('')
  }

  out.push('## Teardown')
  out.push('')
  out.push(table(
    ['Outcome', 'Count', 'Rows'],
    [
      ['Deleted', String(teardown.deleted.length), teardown.deleted.join(', ') || '—'],
      ['Archived instead', String(teardown.archived.length), teardown.archived.join(', ') || '—'],
      ['Failed to remove', String(teardown.failed.length),
        teardown.failed.map(f => f.key).join(', ') || '—'],
    ]
  ))
  out.push('')
  if (teardown.failed.length) {
    out.push('Rows that resisted removal are listed above and remain in the database.')
    out.push('')
    out.push(fence(teardown.failed))
    out.push('')
  }

  return out.join('\n')
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `node --test qa/route-audit/test/report.test.mjs`
Expected: PASS, 8 tests.

- [ ] **Step 5: Commit**

```bash
git add qa/route-audit/lib/report.mjs qa/route-audit/test/report.test.mjs
git commit -m "qa: markdown report generator"
```

---

### Task 14: Orchestrator and README

**Files:**
- Create: `qa/route-audit/run.mjs`
- Create: `qa/route-audit/README.md`
- Modify: `.gitignore` — add `qa/route-audit/out/`

**Interfaces:**
- Consumes: every module from Tasks 1–13.
- Produces: `qa/route-audit/out/report.json` and the rendered Markdown at `docs/2026-09-09_fullstack-route-audit.md`.

Phases are individually selectable so the browser walkthrough, which is manual, does not force a re-run of the API sweep.

- [ ] **Step 1: Write the orchestrator**

```javascript
// qa/route-audit/run.mjs
import { mkdirSync, writeFileSync, readFileSync, readdirSync, existsSync } from 'node:fs'
import { resolve, dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { config, requireCredentials } from './config.mjs'
import { loadSpec, listOperations } from './lib/spec.mjs'
import { createRecorder } from './lib/recorder.mjs'
import { ApiClient } from './lib/client.mjs'
import { buildFixtures, teardownFixtures } from './lib/fixtures.mjs'
import { sweep } from './lib/sweep.mjs'
import { provisionLimitedIdentity, runGateSuite } from './lib/gates.mjs'
import { runWorkflows } from './lib/workflows.mjs'
import { runIsolationSuite } from './lib/isolation.mjs'
import { crossReference, unusedOperations, resetFindingIds } from './lib/crossref.mjs'
import { renderReport } from './lib/report.mjs'

const here = dirname(fileURLToPath(import.meta.url))
const outDir = resolve(here, 'out')
const reportPath = resolve(here, '../../docs/2026-09-09_fullstack-route-audit.md')

// The run id must be supplied, not generated from a clock, so a resumed run
// tears down the rows the earlier phase created.
const runId = process.env.AUDIT_RUN_ID
if (!runId) {
  console.error('AUDIT_RUN_ID is required, e.g. AUDIT_RUN_ID=0909a')
  process.exit(2)
}

const phases = new Set(
  (process.argv[2] ?? 'sweep,gates,workflows,isolation,report').split(',')
)

function loadCaptures() {
  if (!existsSync(config.captureDir)) return []
  return readdirSync(config.captureDir)
    .filter(f => f.endsWith('.json'))
    .map(f => JSON.parse(readFileSync(join(config.captureDir, f), 'utf8')))
}

function save(name, value) {
  mkdirSync(outDir, { recursive: true })
  writeFileSync(join(outDir, name), JSON.stringify(value, null, 2))
}

function loadOut(name, fallback) {
  const p = join(outDir, name)
  return existsSync(p) ? JSON.parse(readFileSync(p, 'utf8')) : fallback
}

requireCredentials()

const spec = loadSpec(config.specPath)
const ops = listOperations(spec)
const recorder = createRecorder()
const client = new ApiClient({ baseUrl: config.baseUrl, recorder })

await client.login(config.email, config.password)
const me = await client.request('GET', '/api/v1/me/permissions', { opKey: 'GET /api/v1/me/permissions' })
const homeCompanyId = me.body?.company_id
console.log(`signed in; company ${homeCompanyId}, ${ops.length} operations in the spec`)

let graph = loadOut('fixtures.json', null)
let sweepResult = loadOut('sweep.json', { results: [], coverage: [] })
let gates = loadOut('gates.json', [])
let workflows = loadOut('workflows.json', [])
let isolationResults = loadOut('isolation.json', [])
let teardown = loadOut('teardown.json', { deleted: [], archived: [], failed: [] })

if (phases.has('sweep')) {
  graph = await buildFixtures(client, runId)
  save('fixtures.json', graph)
  console.log(`fixtures: ${graph.created.length} created, ${graph.failed.length} failed`)

  let done = 0
  sweepResult = await sweep(client, spec, ops, graph, {
    onProgress: () => {
      done += 1
      if (done % 25 === 0) console.log(`  swept ${done}/${ops.length}`)
    },
  })
  save('sweep.json', { results: sweepResult.results, coverage: sweepResult.coverage })
  console.log(`sweep: ${sweepResult.results.filter(r => !r.ok).length} failed checks`)
}

if (phases.has('gates')) {
  if (!graph) throw new Error('gates need fixtures; run the sweep phase first')
  const password = `Zz-${runId}-Test-Pass-1`
  const limited = await provisionLimitedIdentity(client, graph, password)
  gates = await runGateSuite(client, graph, limited, password)
  save('gates.json', gates)
  console.log(`gates: ${gates.filter(r => !r.ok).length} failed`)
}

if (phases.has('workflows')) {
  if (!graph) throw new Error('workflows need fixtures; run the sweep phase first')
  workflows = await runWorkflows(client, graph)
  save('workflows.json', workflows)
  console.log(`workflows: ${workflows.filter(r => !r.ok).length} failed`)
}

if (phases.has('isolation')) {
  if (!graph) throw new Error('isolation needs fixtures; run the sweep phase first')
  isolationResults = await runIsolationSuite(client, graph, homeCompanyId)
  save('isolation.json', isolationResults)
  console.log(`isolation: ${isolationResults.filter(r => !r.ok).length} failed`)
}

if (phases.has('teardown')) {
  if (!graph) throw new Error('teardown needs fixtures')
  teardown = await teardownFixtures(client, graph)
  save('teardown.json', teardown)
  console.log(`teardown: ${teardown.deleted.length} deleted, ${teardown.archived.length} archived, ${teardown.failed.length} failed`)
}

if (phases.has('report')) {
  resetFindingIds()
  const captures = loadCaptures()
  const findings = crossReference(spec, ops, captures)
  const unused = unusedOperations(ops, captures)

  save('requests.json', recorder.entries())

  const md = renderReport({
    runId,
    baseUrl: config.baseUrl,
    generatedAt: new Date().toISOString(),
    operations: ops,
    sweep: sweepResult,
    gates,
    workflows,
    isolation: isolationResults,
    findings,
    captures,
    teardown,
    unused,
  })
  writeFileSync(reportPath, md)
  console.log(`report written to ${reportPath}`)
  console.log(`  ${findings.length} cross-reference findings, ${captures.length} routes walked`)
}
```

- [ ] **Step 2: Write the README**

````markdown
# Route audit harness

Drives the deployed API from the committed OpenAPI spec, walks the deployed
frontend, and renders one report.

## Running

Credentials come from the environment. They are never committed and never
written into the report.

```bash
export AUDIT_EMAIL='...'
export AUDIT_PASSWORD='...'
export AUDIT_RUN_ID='0909a'

node --test qa/route-audit/test/              # harness unit tests
node qa/route-audit/run.mjs sweep,gates,workflows,isolation
# then follow qa/route-audit/ui/walkthrough.md in the browser
node qa/route-audit/run.mjs report
node qa/route-audit/run.mjs teardown
```

`AUDIT_BASE_URL` defaults to `https://go-logistics.jesuslab135.com`.

Phases are selectable because the browser walkthrough is manual; re-rendering
the report should not mean re-running several hundred API calls.

## Data safety

Every row this harness creates is named `ZZ-TEST-<AUDIT_RUN_ID>-...`. Teardown
deletes in reverse dependency order and falls back to archiving rows that are
pinned by a foreign key. The harness writes only to rows it created.

It targets **production**. There is no staging environment. Run it knowing that.

The isolation suite creates a second company and deletes it before finishing.
If a run dies partway, re-running the `teardown` phase with the same
`AUDIT_RUN_ID` cleans up, because fixture ids are persisted to
`qa/route-audit/out/fixtures.json`.

## Layout

| Path | Purpose |
|---|---|
| `config.mjs` | Environment config |
| `lib/` | Harness modules |
| `test/` | Unit tests, run with `node --test` |
| `ui/routes.json` | The 58 frontend routes |
| `ui/walkthrough.md` | Browser protocol |
| `ui/captures/` | Per-route capture files |
| `out/` | Run artifacts, git-ignored |
````

- [ ] **Step 3: Ignore run artifacts**

```bash
printf '\n# Route audit run artifacts\nqa/route-audit/out/\n' >> .gitignore
```

- [ ] **Step 4: Verify the whole harness test suite passes**

Run: `node --test qa/route-audit/test/`
Expected: PASS, all tests across the 9 test files.

- [ ] **Step 5: Commit**

```bash
git add qa/route-audit/run.mjs qa/route-audit/README.md .gitignore
git commit -m "qa: route audit orchestrator"
```

---

### Task 15: Execute the audit and write the report

**Files:**
- Create: `qa/route-audit/ui/captures/*.json` (one per walked route)
- Create: `docs/2026-09-09_fullstack-route-audit.md`

**Interfaces:**
- Consumes: the whole harness.
- Produces: the deliverable.

This task runs the audit rather than building it. It is one task because its phases share a fixture graph and a run id — splitting it would strand fixtures between reviews.

- [ ] **Step 1: Run the API phases**

```bash
export AUDIT_EMAIL='...' AUDIT_PASSWORD='...' AUDIT_RUN_ID='0909a'
node qa/route-audit/run.mjs sweep,gates,workflows,isolation
```

Expected: `out/fixtures.json`, `out/sweep.json`, `out/gates.json`, `out/workflows.json`, `out/isolation.json` all written. Note the failed-check counts printed per phase.

- [ ] **Step 2: Walk the frontend**

Follow `qa/route-audit/ui/walkthrough.md` for all 58 routes, reading `$id` values from `out/fixtures.json`. Write one capture file per route.

Every route must produce a capture file, including routes that fail — an absent capture is indistinguishable from an untested route, and the report cannot tell the difference.

- [ ] **Step 3: Render the report**

```bash
node qa/route-audit/run.mjs report
```

Expected: `docs/2026-09-09_fullstack-route-audit.md` written, with all seven sections populated.

- [ ] **Step 4: Re-verify every finding with a fresh token**

For each finding in the report, re-run its reproduction against a newly-issued token. Remove any finding that does not reproduce, and note the removal in the commit message.

An expired-token 401 is indistinguishable from an authorization defect in the report, and the spec requires this pass precisely because the access token lives 900 seconds while the sweep runs longer.

- [ ] **Step 5: Add file:line citations to backend findings**

For each finding whose layer is `backend` or `contract`, locate the handler in `internal/http/` and append the `file:line` reference to the finding's recommended fix. Frontend findings get a precise description of the required request shape instead, since that source is unavailable.

- [ ] **Step 6: Tear down**

```bash
node qa/route-audit/run.mjs teardown
node qa/route-audit/run.mjs report
```

The second `report` run refreshes the teardown section with the real outcome.

Expected: zero rows in "Failed to remove", or an explicit list of what survived and why.

- [ ] **Step 7: Confirm no secrets reached the report**

Run: `grep -nE 'Bearer [A-Za-z0-9._-]+|NT5rwyo|password"\s*:\s*"[^"]' docs/2026-09-09_fullstack-route-audit.md qa/route-audit/ui/captures/*.json`
Expected: no matches. Any match must be redacted before the commit.

- [ ] **Step 8: Commit**

```bash
git add docs/2026-09-09_fullstack-route-audit.md qa/route-audit/ui/captures/
git commit -m "docs: full-stack route audit results"
```

---

## Self-Review

**Spec coverage:**

| Spec section | Task |
|---|---|
| Layer 1 contract harness | 1–7 |
| `spec.mjs` / `client.mjs` / `fixtures.mjs` / `sweep.mjs` | 1, 3, 6, 7 |
| `workflows.mjs` | 9 |
| `isolation.mjs` | 10 |
| `teardown.mjs` | 6 (implementation), 15 (execution) |
| Layer 2 UI harness | 11, 15 |
| Layer 3 cross-reference | 12 |
| Nine per-operation checks | 5, 8 |
| Six workflow suites | 9 |
| Data safety and `ZZ-TEST-` prefixing | 6, 14 |
| Deliverable, all six report sections | 13, 15 |
| Fresh-token re-verification | 15 step 4 |
| Harness at `qa/route-audit/` | all |

Three gaps found during review, two closed and one recorded as blocked:

0. **The purchase-order transition table was wrong on first draft.** It was written from the route names alone as a five-action machine (`submit`, `approve`, `reject`, `receive`, `close`). Reading `internal/domain/purchaseorder/workflow.go` showed eight actions, with `revise`, `purchase`, `receive-partial` and `receive-full` missing and `receive` not existing at all. Two behaviours a five-action table would never have tested: rejection returns to `DRAFT` through an explicit `revise` rather than being terminal, and `reject` carries `RequiresReason`, so a rejection with no reason must be refused. Task 9 now walks all eight in one ordered pass and reads the resulting `state` back rather than trusting the status code.


1. The spec lists nine checks but names validation among them. Because only 13 of 59 create schemas declare required fields and the Go binding tags are overwhelmingly `omitempty`, a spec-driven validation check would mostly assert the absence of validation the backend never claimed. The `validation` severity entry remains in `checks.mjs` for workflow suites that assert specific rejections, but no generic validation probe is run. **The report must state this in "Untested and blocked"** — Task 15 step 3 covers it via the coverage table, and any operation lacking a validation verdict carries its reason there.

2. The spec's tire-assignment workflow has no task. It is genuinely blocked: `POST /tire-assignment-requests/{id}/approve` requires the `tire_approvals.approve` action, and the audit account holds it, but a request must first exist in a pending state — and the fixture plan creates no tire assignment request, because its required payload could not be determined from the spec (`dto.CreateTireAssignmentRequestRequest` declares no required fields, and the store's validation lives in Go). Rather than invent a payload, Task 15 records the tire-assignment workflow in "Untested and blocked" with that reason. Closing it properly means reading `internal/http/handler/` for the store's real constraints, which is follow-up work, not a silent omission.

**Placeholder scan:** No TBD, TODO, or "similar to Task N" references. Every code step carries runnable code. Task 9 step 1 requires reading `workflow.go` before writing the transition table, and its test asserts agreement with the source rather than trusting the literal.

**Type consistency:** `CheckResult` carries `{check, opKey, ok, expected, actual, severity, evidence}` in Tasks 5, 8, 9 and 10. `FixtureGraph` carries `{runId, tag, ids, created, failed}` in Tasks 6, 7, 9 and 10. `Finding` carries `{id, layer, severity, route, summary, expected, actual, evidence}` in Tasks 12 and 13. `checkHappyPath` additionally returns `response`, consumed by `checkConformance` in Task 7 — declared in Task 5's interface block.
