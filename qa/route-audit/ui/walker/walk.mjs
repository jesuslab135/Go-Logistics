// qa/route-audit/ui/walker/walk.mjs
//
// Browser walkthrough for the route audit. Logs in once, then visits every
// route in ../routes.json via page.goto (full navigation per route, for
// isolation and a meaningful httpStatus — this is a TanStack Router SPA, so
// in-app client nav would not give us that), and writes one capture file
// per route to ../captures/<slug>.json.
//
// Required env: AUDIT_EMAIL, AUDIT_PASSWORD (never printed, never written
// to a capture file).
// Optional env:
//   WALK_ONLY=/app/vehicles,/app/trailers   walk only these route paths
//   WALK_HEADED=1                            launch non-headless
//
// Run from qa/route-audit/ui/walker/:
//   node walk.mjs

import { chromium } from 'playwright'
import { fileURLToPath } from 'node:url'
import path from 'node:path'
import fs from 'node:fs/promises'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const ROUTES_PATH = path.join(__dirname, '..', 'routes.json')
const FIXTURES_PATH = path.join(__dirname, '..', '..', 'out', 'fixtures.json')
const CAPTURES_DIR = path.join(__dirname, '..', 'captures')

const APP_ORIGIN = 'https://go-logistics.netlify.app'

const SENSITIVE_FIELDS = new Set(['password', 'token', 'access_token', 'refresh_token'])

const ERROR_RE = /error|failed|something went wrong|try again/i
const EMPTY_RE = /no .* found|nothing here|no results|empty/i

const BLANK_THRESHOLD = 150 // chars; dashboard measured 643

function envFlag(name) {
  const v = (process.env[name] ?? '').trim().toLowerCase()
  return v !== '' && v !== '0' && v !== 'false' && v !== 'no'
}

function truncate(s, n) {
  if (typeof s !== 'string') s = String(s)
  return s.length > n ? s.slice(0, n) : s
}

function redact(obj) {
  if (obj === null || typeof obj !== 'object') return obj
  if (Array.isArray(obj)) return obj.map(redact)
  const out = {}
  for (const [k, v] of Object.entries(obj)) {
    out[k] = SENSITIVE_FIELDS.has(k) ? '[REDACTED]' : redact(v)
  }
  return out
}

function summarizeResponse(status, body) {
  if (status != null && status >= 400) {
    const code = body?.error?.code
    return truncate(code ? `error.code ${code}` : `HTTP ${status}`, 100)
  }
  if (body && Array.isArray(body.data) && typeof body.has_next === 'boolean') {
    return truncate(`${body.data.length} rows, has_next ${body.has_next}`, 100)
  }
  if (body && typeof body === 'object' && body.id !== undefined && body.id !== null) {
    return truncate(`id ${body.id}`, 100)
  }
  if (body == null) return 'no body'
  return truncate('object response', 100)
}

function slugify(routePath) {
  return routePath.replace(/[/$]/g, '-').replace(/^-+/, '')
}

function resolveRoutePath(route, ids) {
  if (!route.path.includes('$id')) return route.path
  if (route.path === '/app/admin/companies/$id') {
    return route.path.replace('$id', '1')
  }
  const key = route.needsFixture
  if (!key || !(key in ids)) {
    throw new Error(`no fixture id available for ${route.path} (needsFixture=${key})`)
  }
  const idVal = ids[key]
  if (idVal === null || idVal === undefined || idVal === true) {
    throw new Error(`fixture id for ${route.path} (needsFixture=${key}) is not a usable id: ${JSON.stringify(idVal)}`)
  }
  return route.path.replace('$id', String(idVal))
}

// ---- form-filling helpers (best-effort; a failure to locate a control is
// a finding, not a crash) ----------------------------------------------

async function tryFill(page, labelPatterns, value) {
  for (const pat of labelPatterns) {
    try {
      let loc = page.getByLabel(pat, { exact: false })
      if ((await loc.count()) === 0) loc = page.getByPlaceholder(pat)
      if ((await loc.count()) === 0) continue
      await loc.first().fill(value, { timeout: 5000 })
      return { ok: true }
    } catch {
      // try next pattern
    }
  }
  return { ok: false }
}

async function trySelectCombobox(page, labelPatterns, optionText) {
  for (const pat of labelPatterns) {
    try {
      const labeled = page.getByLabel(pat, { exact: false })
      if ((await labeled.count()) > 0) {
        const tagName = await labeled
          .first()
          .evaluate((n) => n.tagName.toLowerCase())
          .catch(() => null)
        if (tagName === 'select') {
          await labeled.first().selectOption({ label: optionText })
          return { ok: true }
        }
      }
      let trigger = page.getByRole('combobox', { name: pat })
      if ((await trigger.count()) === 0) trigger = labeled
      if ((await trigger.count()) === 0) continue
      await trigger.first().click({ timeout: 5000 })
      const option = page.getByRole('option', { name: optionText, exact: false })
      await option.first().waitFor({ state: 'visible', timeout: 5000 })
      await option.first().click({ timeout: 5000 })
      return { ok: true }
    } catch {
      // try next pattern
    }
  }
  return { ok: false }
}

function formSpec(routePath) {
  switch (routePath) {
    case '/app/vehicles/new':
      return { mode: 'new', kind: 'vehicle' }
    case '/app/vehicles/$id/edit':
      return { mode: 'edit', kind: 'vehicle' }
    case '/app/trailers/new':
      return { mode: 'new', kind: 'trailer' }
    case '/app/trailers/$id/edit':
      return { mode: 'edit', kind: 'trailer' }
    default:
      return null
  }
}

async function attemptFormSubmit(page, spec, tag) {
  const issues = []
  const nameValue =
    spec.mode === 'new'
      ? spec.kind === 'vehicle'
        ? `${tag}-asset`
        : `${tag}-trailer-asset`
      : spec.kind === 'vehicle'
        ? `${tag}-asset-edited`
        : `${tag}-trailer-asset-edited`

  const nameResult = await tryFill(page, [/name/i], nameValue)
  if (!nameResult.ok) issues.push('could not locate the Name field by label/placeholder')

  if (spec.mode === 'new') {
    const vinValue = spec.kind === 'vehicle' ? `${tag}-VIN` : `${tag}-TVIN`
    const vinResult = await tryFill(page, [/vin.*serial/i, /serial number/i, /\bvin\b/i], vinValue)
    if (!vinResult.ok) issues.push('could not locate the VIN / Serial Number field by label/placeholder')

    const atResult = await trySelectCombobox(page, [/asset type/i], `${tag}-asset-type`)
    if (!atResult.ok) issues.push('could not select the Asset Type option (combobox/select not found, or the option was not present)')

    const stResult = await trySelectCombobox(page, [/status/i], `${tag}-asset-status`)
    if (!stResult.ok) issues.push('could not select the Status option (combobox/select not found, or the option was not present)')

    // Vehicle Type / Ownership Type: "if the form asks — otherwise accept
    // the default". Best-effort only; not a blocking issue either way.
    await trySelectCombobox(page, [/vehicle type/i], spec.kind === 'vehicle' ? 'Vehicle' : 'Trailer').catch(() => {})
    await trySelectCombobox(page, [/ownership type/i], 'Owned').catch(() => {})
  }
  // On edit: Asset Type, Status, Vehicle Type, Ownership Type are left as
  // loaded per the field table — no interaction.

  const submitBtn = page.getByRole('button', { name: /save|create|submit/i })
  const submitCount = await submitBtn.count().catch(() => 0)
  if (submitCount === 0) issues.push('could not locate a Save/Create/Submit button')

  if (issues.length > 0) {
    return { blocked: true, note: `could not drive this form: ${issues.join('; ')}` }
  }

  await submitBtn.first().click({ timeout: 5000 })
  await page.waitForLoadState('networkidle', { timeout: 15000 }).catch(() => {})
  await page.waitForTimeout(2500)
  return { blocked: false, note: `submitted the ${spec.mode} form (${spec.kind})` }
}

// ---- rendered classification -------------------------------------------

async function classifyRendered(page, requests) {
  const bodyText = await page.evaluate(() => document.body?.innerText || '').catch(() => '')
  const trimmedLen = bodyText.trim().length

  if (trimmedLen < BLANK_THRESHOLD) {
    return { rendered: 'blank', reason: `body innerText length ${trimmedLen} < ${BLANK_THRESHOLD} threshold` }
  }

  const hasXhrError = requests.some((r) => r.status != null && r.status >= 400)
  const hasErrorText = ERROR_RE.test(bodyText)
  if (hasXhrError || hasErrorText) {
    return {
      rendered: 'error',
      reason: hasXhrError
        ? `an XHR responded with status >= 400 (${requests.filter((r) => r.status != null && r.status >= 400).map((r) => `${r.method} ${r.url} -> ${r.status}`).join(', ')})`
        : 'visible text matched /error|failed|something went wrong|try again/i',
    }
  }

  const hasEmptyText = EMPTY_RE.test(bodyText)
  const hasZeroRowsTable = await page
    .evaluate(() => {
      const tables = document.querySelectorAll('table')
      for (const t of tables) {
        const tbody = t.querySelector('tbody') || t
        const rows = tbody.querySelectorAll('tr')
        if (rows.length === 0) return true
      }
      return false
    })
    .catch(() => false)
  if (hasEmptyText || hasZeroRowsTable) {
    return {
      rendered: 'empty',
      reason: hasEmptyText
        ? 'visible text matched /no .* found|nothing here|no results|empty/i'
        : 'a <table> rendered with zero rows',
    }
  }

  return { rendered: 'data', reason: 'default: page rendered content with no error or empty-state signal' }
}

// ---- main ----------------------------------------------------------------

async function main() {
  if (!process.env.AUDIT_EMAIL || !process.env.AUDIT_PASSWORD) {
    console.error('walk.mjs: AUDIT_EMAIL and AUDIT_PASSWORD must both be set in the environment. Aborting.')
    process.exit(1)
  }

  const routesFile = JSON.parse(await fs.readFile(ROUTES_PATH, 'utf8'))
  const allRoutes = routesFile.routes
  const fixtures = JSON.parse(await fs.readFile(FIXTURES_PATH, 'utf8'))
  const { ids, tag } = fixtures

  let routesToWalk = allRoutes
  if (process.env.WALK_ONLY) {
    const only = new Set(
      process.env.WALK_ONLY.split(',')
        .map((s) => s.trim())
        .filter(Boolean)
    )
    routesToWalk = allRoutes.filter((r) => only.has(r.path))
    console.log(`WALK_ONLY set: walking ${routesToWalk.length}/${allRoutes.length} routes`)
  }

  await fs.mkdir(CAPTURES_DIR, { recursive: true })

  const browser = await chromium.launch({ headless: !envFlag('WALK_HEADED') })
  const page = await (await browser.newContext()).newPage()

  // Fresh-per-route collectors: these arrays are reassigned to [] right
  // before each navigation; the listeners below close over the `let`
  // bindings, so resetting the binding is enough — no need to
  // attach/detach per route.
  let requests = []
  let consoleErrors = []

  page.on('console', (msg) => {
    if (msg.type() === 'error') consoleErrors.push(truncate(msg.text(), 200))
  })
  page.on('pageerror', (err) => {
    consoleErrors.push(truncate(err?.message ?? String(err), 200))
  })
  page.on('requestfailed', (req) => {
    try {
      const url = new URL(req.url())
      if (!(url.pathname.startsWith('/api/v1') || url.pathname.startsWith('/auth'))) return
      requests.push({
        method: req.method(),
        url: url.pathname + url.search,
        status: null,
        requestBody: null,
        responseSummary: truncate(`request failed: ${req.failure()?.errorText ?? 'unknown'}`, 100),
      })
    } catch {
      // ignore
    }
  })
  page.on('response', async (response) => {
    try {
      const req = response.request()
      const url = new URL(req.url())
      if (!(url.pathname.startsWith('/api/v1') || url.pathname.startsWith('/auth'))) return
      const method = req.method()
      const status = response.status()
      let requestBody = null
      if (['POST', 'PUT', 'PATCH'].includes(method)) {
        try {
          requestBody = redact(req.postDataJSON())
        } catch {
          requestBody = null
        }
      }
      let bodyJson = null
      try {
        bodyJson = await response.json()
      } catch {
        bodyJson = null
      }
      requests.push({
        method,
        url: url.pathname + url.search,
        status,
        requestBody,
        responseSummary: summarizeResponse(status, bodyJson),
      })
    } catch {
      // never let a listener crash the walk
    }
  })

  // --- login (verified working; do not modify) ---
  await page.goto(`${APP_ORIGIN}/login`, { waitUntil: 'networkidle', timeout: 60000 })
  await page.fill('input[name=email]', process.env.AUDIT_EMAIL)
  await page.fill('input[name=password]', process.env.AUDIT_PASSWORD)
  await page.click('button:has-text("Log In")')
  await page.waitForTimeout(7000)

  const tally = { data: 0, empty: 0, error: 0, blank: 0 }
  let routesWithConsoleErrors = 0
  let routesWithFailingXhr = 0
  const nonDataRoutes = []

  for (const route of routesToWalk) {
    requests = []
    consoleErrors = []

    let httpStatus = null
    let rendered = 'blank'
    let notesParts = []
    let currentUrl = ''

    try {
      const resolvedPath = resolveRoutePath(route, ids)
      const fullUrl = `${APP_ORIGIN}${resolvedPath}`

      const response = await page.goto(fullUrl, { waitUntil: 'networkidle', timeout: 60000 })
      httpStatus = response ? response.status() : null
      await page.waitForTimeout(2500)
      currentUrl = page.url()

      const finalPath = (() => {
        try {
          return new URL(currentUrl).pathname
        } catch {
          return currentUrl
        }
      })()
      if (finalPath !== resolvedPath) {
        notesParts.push(`redirected from ${resolvedPath} to ${finalPath}`)
      }

      if (route.kind === 'form') {
        const spec = formSpec(route.path)
        if (!spec) {
          notesParts.push(`route is marked kind: "form" but no field-table entry exists for ${route.path}`)
        } else {
          const result = await attemptFormSubmit(page, spec, tag)
          notesParts.push(result.note)
          if (result.blocked) {
            rendered = 'error'
          }
          currentUrl = page.url()
        }
      }

      if (rendered !== 'error') {
        const classification = await classifyRendered(page, requests)
        rendered = classification.rendered
        notesParts.push(classification.reason)
      }
    } catch (err) {
      currentUrl = (() => {
        try {
          return page.url()
        } catch {
          return currentUrl
        }
      })()
      rendered = 'blank'
      notesParts.push(`threw during walk: ${truncate(err?.message ?? String(err), 300)} (url at failure: ${currentUrl})`)
      console.log(`${route.path} -> THREW: ${err?.message ?? err} (url: ${currentUrl})`)
    }

    const capture = {
      route: route.path,
      loadedAt: new Date().toISOString(),
      httpStatus,
      requests,
      consoleErrors,
      rendered,
      notes: notesParts.filter(Boolean).join(' | '),
    }

    const slug = slugify(route.path)
    await fs.writeFile(path.join(CAPTURES_DIR, `${slug}.json`), JSON.stringify(capture, null, 2), 'utf8')

    tally[rendered] = (tally[rendered] ?? 0) + 1
    if (rendered !== 'data') nonDataRoutes.push(`${route.path} (${rendered})`)
    if (consoleErrors.length > 0) routesWithConsoleErrors++
    if (requests.some((r) => r.status != null && r.status >= 400)) routesWithFailingXhr++

    console.log(`${route.path} -> ${rendered}, ${requests.length} requests, ${consoleErrors.length} console errors`)
  }

  await browser.close()

  console.log('\n--- walk complete ---')
  console.log(`routes walked: ${routesToWalk.length}`)
  console.log(`rendered tally: data=${tally.data ?? 0} empty=${tally.empty ?? 0} error=${tally.error ?? 0} blank=${tally.blank ?? 0}`)
  console.log(`routes with console errors: ${routesWithConsoleErrors}`)
  console.log(`routes with a failing (>=400) XHR: ${routesWithFailingXhr}`)
  if (nonDataRoutes.length > 0) {
    console.log('non-data routes:')
    for (const r of nonDataRoutes) console.log(`  ${r}`)
  }
}

main().catch((err) => {
  console.error('walk.mjs: fatal error outside the per-route loop:', err)
  process.exit(1)
})
