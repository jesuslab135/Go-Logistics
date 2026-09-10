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
   `needsFixture`, read from the run's `report.json`. The one exception is
   `/app/admin/companies/$id`, whose `needsFixture` is `null` by design —
   substitute company id `1` there instead of reading a fixture id.
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
