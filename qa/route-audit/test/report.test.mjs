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
    findings: [], captures: [],
    teardown: { deleted: [], archived: [], reversed: [], failed: [] },
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
    findings: [], captures: [],
    teardown: { deleted: [], archived: [], reversed: [], failed: [] },
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
    captures: [],
    teardown: { deleted: [], archived: [], reversed: [], failed: [] },
    unused: [],
  })
  assert.ok(md.includes('[REDACTED]'))
  assert.ok(!md.includes('Bearer '))
})

test('renderReport shows a reversed teardown row and never counts it as deleted', () => {
  const md = renderReport({
    runId: 'test', baseUrl: 'https://example.com',
    generatedAt: '2026-09-09T00:00:00Z',
    operations: [], sweep: { results: [], coverage: [] },
    gates: [], workflows: [], isolation: [],
    findings: [], captures: [],
    teardown: {
      deleted: ['asset#1'],
      archived: [],
      reversed: ['journalEntry#7'],
      failed: [],
    },
    unused: [],
  })
  assert.ok(md.includes('journalEntry#7'), 'reversed row content is missing from the report')
  // The reversed row must not be folded into the Deleted count or row.
  const deletedRowMatch = /\|\s*Deleted\s*\|\s*(\d+)\s*\|/i.exec(md)
  assert.ok(deletedRowMatch, 'could not find a Deleted row in the teardown table')
  assert.equal(deletedRowMatch[1], '1')
  // The reversed bucket must appear as its own labelled row (neutralised,
  // not removed), not merged into Deleted or Archived.
  assert.match(md, /reversed/i)
  assert.match(md, /neutralised|not removed/i)
})
