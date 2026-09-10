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
  // teardownFixtures returns four buckets: deleted, archived, reversed,
  // failed. `reversed` is NOT deleted — DELETE on an append-only ledger
  // entry (e.g. an inventory journal entry) returns 405 by design, and the
  // API's own replacement is POST .../reverse, which neutralises the row
  // with a balancing entry rather than removing it. Folding reversed rows
  // into the Deleted count would misstate what is actually left behind in
  // a live production database, so it gets its own row.
  out.push(table(
    ['Outcome', 'Count', 'Rows'],
    [
      ['Deleted', String(teardown.deleted.length), teardown.deleted.join(', ') || '—'],
      ['Archived instead', String(teardown.archived.length), teardown.archived.join(', ') || '—'],
      ['Reversed (neutralised, not removed)', String(teardown.reversed.length),
        teardown.reversed.join(', ') || '—'],
      ['Failed to remove', String(teardown.failed.length),
        teardown.failed.map(f => f.key).join(', ') || '—'],
    ]
  ))
  out.push('')
  if (teardown.reversed.length) {
    out.push('Reversed rows still exist in the database, each beside a balancing')
    out.push('entry the API created to neutralise it. They are not deletions.')
    out.push('')
  }
  if (teardown.failed.length) {
    out.push('Rows that resisted removal are listed above and remain in the database.')
    out.push('')
    out.push(fence(teardown.failed))
    out.push('')
  }

  return out.join('\n')
}
