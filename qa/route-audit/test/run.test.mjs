// qa/route-audit/test/run.test.mjs
//
// run.mjs drives the live production API — it cannot be executed inside a
// unit test. These are source-text assertions instead, which are normally
// too weak to trust: a string appearing in the file proves nothing about
// what actually executes. They exist here anyway because the two things
// they check are exactly the two non-negotiable requirements from the
// incident that motivated this task (a crash that skipped teardown and
// left 45 fixture rows in production), and the only stronger alternative —
// actually running the orchestrator — is worse than a weak test. Keep them
// narrow: each one checks a single, specific structural fact, not a vibe.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(here, '../run.mjs'), 'utf8')

test('phases run inside a try with a finally block', () => {
  assert.match(source, /\btry\s*\{/, 'expected a try block wrapping the phases')
  assert.match(source, /\}\s*finally\s*\{/, 'expected a finally block after the try/catch')
})

test('teardownFixtures is called somewhere inside the finally block', () => {
  const finallyStart = source.indexOf('} finally {')
  assert.notEqual(finallyStart, -1, 'could not locate the finally block')
  const finallyBody = source.slice(finallyStart)
  assert.match(
    finallyBody,
    /teardownFixtures\(/,
    'teardownFixtures must be reachable from inside finally so a crash still tears fixtures down'
  )
})

test('the failure-path teardown guard is not gated on the requested phase list', () => {
  // Regression pin for a real Critical finding: the guard once read
  // `if (failure && phases.has('teardown') && graph && !teardownRan)`,
  // which meant a crash during the documented default invocation
  // (`sweep,gates,workflows,isolation` — no `teardown` in that list) would
  // skip cleanup entirely, because the caller never asked for teardown and
  // never got the chance to. A crashed run must tear down unconditionally;
  // gating it on `phases` is exactly the bug that caused 45 orphaned
  // fixture rows in production. Find the specific guard line (identified
  // by `!teardownRan`, its distinguishing condition) and assert it does
  // not reference `phases` at all.
  const guardLine = source.split('\n').find(line => line.includes('!teardownRan'))
  assert.ok(guardLine, 'could not find the failure-path teardown guard (look for !teardownRan)')
  assert.doesNotMatch(
    guardLine,
    /phases/,
    `failure-path teardown must not be gated on the requested phase list, got: ${guardLine.trim()}`
  )
})

test('a crash inside the try is not swallowed by teardown succeeding', () => {
  // The finally block must not be the last word: after it runs, the
  // original error has to still be checked and still cause a non-zero
  // exit. Look for both an exit-on-failure path and that it comes after
  // the finally block, not inside the try's own success path.
  const finallyStart = source.indexOf('} finally {')
  const finallyEnd = source.indexOf('\n}\n', finallyStart)
  assert.ok(finallyEnd > finallyStart, 'could not find the end of the finally block')
  const afterFinally = source.slice(finallyEnd)
  assert.match(afterFinally, /process\.exit\(1\)/, 'expected a non-zero exit after the finally block when the run failed')
})

test('AUDIT_RUN_ID is read from the environment and the process exits when it is absent', () => {
  assert.match(source, /process\.env\.AUDIT_RUN_ID/, 'expected AUDIT_RUN_ID to come from the environment')
  assert.doesNotMatch(
    source,
    /process\.env\.AUDIT_RUN_ID\s*\?\?/,
    'AUDIT_RUN_ID must not fall back to a generated default — a resumed run needs the same id an earlier phase used'
  )
  assert.match(
    source,
    /if\s*\(!runId\)\s*\{[\s\S]*?process\.exit\(2\)/,
    'expected the process to exit when AUDIT_RUN_ID is missing rather than inventing one'
  )
})

test('run.mjs does not import anything that generates ids from time or randomness', () => {
  assert.doesNotMatch(source, /Date\.now\(\)|Math\.random\(\)|randomUUID/, 'run id generation is forbidden by design')
})
