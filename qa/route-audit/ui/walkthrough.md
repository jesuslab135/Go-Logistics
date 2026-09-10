# Browser walkthrough protocol

Run after the Layer 1 sweep, so fixture ids exist for the `$id` routes.

## Setup

1. Confirm the claude-in-chrome extension has permission for
   `go-logistics.netlify.app`.
2. Open `https://go-logistics.netlify.app/login`.
3. Sign in with `AUDIT_EMAIL` / `AUDIT_PASSWORD`.
4. Open DevTools, Network tab, and enable "Preserve log".
5. Find the run's tag. The sweep phase is invoked with `AUDIT_RUN_ID` (for
   example `AUDIT_RUN_ID=0909a`); the run tag is
   `ZZ-TEST-<AUDIT_RUN_ID>` — so for `AUDIT_RUN_ID=0909a` the tag is
   `ZZ-TEST-0909a`. This is the same tag `qa/route-audit/lib/fixtures.mjs`
   uses to build and later tear down its rows. Every free-text value you
   type into a form during this walkthrough must start with this tag, e.g.
   a vehicle name field should read `ZZ-TEST-0909a-vehicle`.

## Fixture ids

`$id` segments in `routes.json` are resolved from
`qa/route-audit/out/fixtures.json`, written by the sweep phase. That file's
top-level shape is `{ runId, tag, ids, created, failed }`; the ids you need
live under its `ids` key.

Each route in `routes.json` names the fixture key that supplies its id in
its own `needsFixture` field — that field name is the key to read under
`ids`. For example:
- `/app/vehicles/$id` has `needsFixture: "asset"` → use `ids.asset`.
- `/app/trailers/$id` has `needsFixture: "trailerAsset"` → use `ids.trailerAsset`.
- `/app/maintenance/$id` has `needsFixture: "workOrder"` → use `ids.workOrder`.

The one exception is `/app/admin/companies/$id`, whose `needsFixture` is
`null` by design — substitute company id `1` there instead of reading a
fixture id.

## Per route

For each entry in `routes.json`, in the order listed:

1. Navigate to the route, substituting `$id` per "Fixture ids" above.
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

5. On a `form` route, submit the form once, using the field table below for
   the route being walked. Prefix every free-text field with the run tag
   (see Setup, step 5). Record the resulting request and status. Do not
   submit a second time.

   If the form cannot be submitted at all — a required dropdown has no
   valid option, a submit button does nothing, a required field will not
   accept the value the table below specifies — that is itself a finding.
   Do not skip the route. Record the capture with `rendered: "error"` and
   describe exactly what blocked submission in `notes`. A route silently
   omitted from the captures directory is indistinguishable, in the report,
   from a route that passed.

### Form field table

`routes.json` currently has four `form`-kind routes with a fixture behind
them: `/app/vehicles/new`, `/app/vehicles/$id/edit`, `/app/trailers/new`,
and `/app/trailers/$id/edit`. Their field values mirror the `asset` and
`trailerAsset` entries in `FIXTURE_PLAN`
(`qa/route-audit/lib/fixtures.mjs`), so the browser layer and the API layer
exercise comparable data. `<tag>` below is the run tag from Setup step 5
(e.g. `ZZ-TEST-0909a`).

| Route | Field | Value |
|---|---|---|
| `/app/vehicles/new` | Name | `<tag>-asset` |
| | VIN / Serial Number | `<tag>-VIN` |
| | Asset Type | select the option named `<tag>-asset-type` (created by the sweep's `assetType` fixture) |
| | Status | select the option named `<tag>-asset-status` (created by the sweep's `assetStatus` fixture) |
| | Vehicle Type | `Vehicle` (`VEHICLE`), if the form asks — otherwise accept the default |
| | Ownership Type | `Owned` (`OWNED`), if the form asks — otherwise accept the default |
| | any other field | accept the default |
| `/app/vehicles/$id/edit` | Name | `<tag>-asset-edited` (prove the edit round-trips; keep the `<tag>-VIN` value already on the fixture row) |
| | Asset Type, Status, Vehicle Type, Ownership Type | leave as loaded — do not change; accept the default for anything else |
| | any other field | accept the default |
| `/app/trailers/new` | Name | `<tag>-trailer-asset` |
| | VIN / Serial Number | `<tag>-TVIN` |
| | Asset Type | select the option named `<tag>-asset-type` |
| | Status | select the option named `<tag>-asset-status` |
| | Vehicle Type | `Trailer` (`TRAILER`), if the form asks — otherwise accept the default |
| | Ownership Type | `Owned` (`OWNED`), if the form asks — otherwise accept the default |
| | any other field | accept the default |
| `/app/trailers/$id/edit` | Name | `<tag>-trailer-asset-edited` (keep the `<tag>-TVIN` value already on the fixture row) |
| | Asset Type, Status, Vehicle Type, Ownership Type | leave as loaded — do not change; accept the default for anything else |
| | any other field | accept the default |

"Accept the default" means: do not invent a value — leave whatever the
form pre-fills or offers as its first/default option, so different
operators produce the same capture.

## Rules

- Never record a bearer token or password into a capture file.
- If a route 404s in the frontend router, record `httpStatus: 404` and
  `rendered: "blank"`. A route present in the bundle but not reachable is a
  finding.
- If a screen calls an endpoint that does not exist in `swagger.json`, note it
  in `notes`. The cross-reference step will classify it.
- If a form cannot be submitted at all, that is a finding, not a skip: record
  the capture with `rendered: "error"` and explain what blocked submission
  in `notes` (see step 5 above).
