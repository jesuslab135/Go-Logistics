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
// tears down the rows the earlier phase created. Never default this: a
// generated id would silently orphan whatever a prior invocation created.
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
let teardown = loadOut('teardown.json', { deleted: [], archived: [], reversed: [], failed: [] })

// NON-NEGOTIABLE: once buildFixtures has created anything, every subsequent
// phase runs inside this try so that teardown still executes if any of them
// throws. A prior version of this harness ran the phases sequentially at
// top level; a suite that threw on an unexpected API response killed the
// run before teardown ever ran, leaving 45 fixture rows behind in
// production — including six that were foreign-key pinned by nested
// children and had to be hunted down and deleted by hand. Do not repeat
// that: the fixture graph must be persisted before any phase that can
// throw, and teardown must run even on failure.
let failure = null
let teardownRan = false
try {
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
    teardownRan = true
    console.log(`teardown: ${teardown.deleted.length} deleted, ${teardown.archived.length} archived, ${teardown.reversed.length} reversed, ${teardown.failed.length} failed`)
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
} catch (err) {
  // Persist whatever fixture graph exists BEFORE anything else, so a dead
  // run can still be recovered by `node run.mjs teardown` with the same
  // AUDIT_RUN_ID even if this process is about to exit non-zero.
  failure = err
} finally {
  if (graph) save('fixtures.json', graph)

  // If the run died before reaching the teardown phase on its own, and the
  // caller actually asked for teardown in this invocation, run it now so a
  // crash never leaves fixture rows behind. Guard this: if teardown itself
  // throws, that must not replace or hide the original failure.
  if (failure && phases.has('teardown') && graph && !teardownRan) {
    try {
      teardown = await teardownFixtures(client, graph)
      save('teardown.json', teardown)
      console.error(`teardown after failure: ${teardown.deleted.length} deleted, ${teardown.archived.length} archived, ${teardown.reversed.length} reversed, ${teardown.failed.length} failed`)
    } catch (teardownErr) {
      console.error('teardown also failed while handling the original error:')
      console.error(teardownErr)
    }
  }
}

if (failure) {
  console.error('route audit run failed:')
  console.error(failure)
  process.exit(1)
}
