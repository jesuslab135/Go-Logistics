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

node --test qa/route-audit/test/*.test.mjs     # harness unit tests
node qa/route-audit/run.mjs sweep,gates,workflows,isolation
# then follow qa/route-audit/ui/walkthrough.md in the browser
node qa/route-audit/run.mjs report
node qa/route-audit/run.mjs teardown
```

`AUDIT_BASE_URL` defaults to `https://go-logistics.jesuslab135.com`.

Phases are selectable because the browser walkthrough is manual; re-rendering
the report should not mean re-running several hundred API calls.

`AUDIT_RUN_ID` is required on every invocation and is never generated from a
clock or random source. It has to be the *same* value across invocations for
a given run, because it is how a resumed or recovered run finds the rows an
earlier invocation created — see Data safety below.

## Data safety

**This targets production.** There is no staging environment. Run it
knowing that.

Every row this harness creates is named `ZZ-TEST-<AUDIT_RUN_ID>-...`. The
harness writes only to rows it created.

If any phase throws after fixtures exist — an unexpected API response,
anything the harness didn't anticipate — the harness tears down
automatically. This is unconditional: it happens regardless of which
phases were requested for that invocation, because a crashed run's
fixtures are not resumable state worth preserving as-is. The console will
print `run failed; tearing down fixtures created by this run` followed by
the teardown counts, then the process exits non-zero.

The one case automatic teardown can't cover is the process being killed
hard enough that it never gets to run at all — `SIGKILL`, a lost
connection, a power loss. For that, the fixture graph is persisted to
`qa/route-audit/out/fixtures.json` before any phase that could throw, so
re-running `node qa/route-audit/run.mjs teardown` with the **same**
`AUDIT_RUN_ID` reads that file and cleans up the rows the dead run
created.

Teardown walks the created rows in reverse dependency order and reports into
four buckets:

| Bucket | Meaning |
|---|---|
| `deleted` | The row's DELETE endpoint succeeded, or a later GET confirmed it is gone (e.g. removed by a cascade from a row deleted afterward in the walk). |
| `archived` | DELETE was refused because the row is a referenced catalog entry (vendor, part, asset, service task, inspection form); it was archived instead. The row still exists, marked inactive. |
| `reversed` | The row lives on an append-only ledger (currently only `journalEntry`) that cannot be deleted at all. A reversal row was posted to neutralise it. **The original row still exists** — the ledger is append-only by design, so "cleaned up" here means balanced, not removed. |
| `failed` | Every fallback was refused and a follow-up GET did not confirm removal. These need manual attention; the report includes their key, id, status, and response body. |

The isolation suite creates a second company and deletes it before
finishing.

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
