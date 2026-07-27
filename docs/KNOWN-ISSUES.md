# Known issues

Backend issues found while reconnecting the frontend (`fleet-admin-app`) to this
API. Each entry is written so a backend engineer can act on it without reading
frontend code.

Found 2026-07-26/27 during the Django→Go frontend migration. The frontend-side
view of the same work lives in `fleet-admin-app/docs/backend-migration/`.

**This file is defects only** — routes that exist but misbehave. For capabilities
the client needs that don't exist yet (missing routes, missing fields, missing
filters), see [`FRONTEND-REQUESTS.md`](./FRONTEND-REQUESTS.md).

| # | Issue | Severity | Status |
|---|---|---|---|
| 1 | `STORAGE_MINIO_PUBLIC_URL` dropped the bucket → every uploaded image 403s | High | ✅ fixed |
| 2 | Upload-URL columns are `varchar(100)` — too small for the URLs we now store | **High** | ⬜ open |
| 3 | No default `asset_status` rows, and no way to create them from the product | **High** | ⬜ open (worked around by seeding) |
| 4 | Replaced/cleared uploads leak objects in storage forever | Medium | ⬜ open |
| 5 | `POST /api/v1/uploads` does no server-side type/size validation | Medium | ⬜ open |
| 6 | `/api/v1/companies` is neither tenant-scoped nor admin-gated | **High** | ⬜ open |
| 7 | Public bucket: any upload is world-readable across tenants | Low (by design today) | ⬜ needs a decision |

---

## 1. `STORAGE_MINIO_PUBLIC_URL` dropped the bucket ✅ fixed

**Symptom:** uploads returned `201` and a URL, but every image was broken in the
client. Nothing in the API logs indicated a problem.

**Cause:** `publicBase()` (`internal/platform/storage/minio.go:110-119`) is
asymmetric — when `PublicURL` is set it is used verbatim, but the endpoint
fallback appends the bucket:

```go
if cfg.PublicURL != "" {
    return strings.TrimRight(cfg.PublicURL, "/")           // no bucket
}
return fmt.Sprintf("%s://%s/%s", scheme, cfg.Endpoint, cfg.Bucket)  // bucket
```

`docker-compose.yml` set `STORAGE_MINIO_PUBLIC_URL: http://localhost:9010`, so
objects written to `fleet/uploads/…` were advertised at
`http://localhost:9010/uploads/…`. MinIO's public-read policy only grants
`s3:GetObject` on `arn:aws:s3:::fleet/*`, so those URLs answer `403`.

```
http://localhost:9010/uploads/1/<token>-truck.png        -> 403
http://localhost:9010/fleet/uploads/1/<token>-truck.png  -> 200
```

**Fix applied:** `STORAGE_MINIO_PUBLIC_URL: http://localhost:9010/fleet` in
`docker-compose.yml`, plus a warning in the README's object-storage section.
The `.env.example` QA/PROD blocks were already correct
(`https://qa-cdn.your-domain.com/fleet-qa`) — only compose diverged.

**Worth hardening** (not done): make the asymmetry impossible rather than
documented. Either always append the bucket, or validate at startup that
`PublicURL` ends with the bucket name and fail fast if not. A misconfiguration
here is silent on the write path and only surfaces as broken images in a client.

---

## 2. Upload-URL columns are `varchar(100)` ⬜ open — highest priority

Every column that stores an upload URL is `varchar(100)`, in
`internal/db/migrations/000001_init.up.sql`:

| Line | Column | Nullable |
|---|---|---|
| 32 | `company.logo` | yes |
| 156 | `asset.photo` | yes |
| 1025 | `fuel_photo.file` | **NOT NULL** |
| 1107 | `inspection_submission_item.photo` | yes |
| 1122 | `media.file` | **NOT NULL** |

**Why this is a real bug now and wasn't in Django:** Django's `FileField` defaults
to `max_length=100` and stored a **relative path** (`uploads/1/x.png`, ~20-40
chars). This API stores a **fully-qualified absolute URL**, because that's what
`POST /api/v1/uploads` returns and what clients render directly.

Measured, not theoretical — a real upload of a file named `smoke-truck.png`:

```
http://localhost:9010/fleet/uploads/1/915ba612fc34639c-smoke-truck.png   = 70 chars
```

70% of the budget consumed by a short, friendly filename on a **localhost** base
URL. The fixed prefix is 55 characters
(`http://localhost:9010/fleet/uploads/{company_id}/{16-hex}-`), leaving ~45 for
the filename. Two very ordinary cases overflow it:

- A phone-generated name: `WhatsApp Image 2026-07-26 at 10.32.11 PM.jpeg` (45).
- Any production CDN base longer than `http://localhost:9010/fleet` (27 chars) —
  e.g. `https://cdn.fleetza.com/fleet-prod` is 34, and pushes the ceiling down.

**Impact:** Postgres `varchar(n)` raises `22001 string_data_right_truncation` on
overflow — so this is a hard `500` on save, not a silent truncation. For
`fuel_photo.file` and `media.file` (both NOT NULL) the record cannot be written
at all.

**Suggested migration:**

```sql
ALTER TABLE company                    ALTER COLUMN logo  TYPE text;
ALTER TABLE asset                      ALTER COLUMN photo TYPE text;
ALTER TABLE fuel_photo                 ALTER COLUMN file  TYPE text;
ALTER TABLE inspection_submission_item ALTER COLUMN photo TYPE text;
ALTER TABLE media                      ALTER COLUMN file  TYPE text;
```

`text` and `varchar(n)` are the same underlying type in Postgres with no
performance difference, and widening takes no table rewrite. If a bound is
preferred for validation reasons, `varchar(500)` is ample.

**Alternative worth considering:** store the **key** (`uploads/1/<token>-x.png`)
rather than the full URL, and have the API compose the public URL at read time.
That decouples stored rows from the storage host — as written today, changing
`STORAGE_MINIO_PUBLIC_URL` or moving to a CDN silently invalidates every URL
already in the database. This is the more robust fix; #2's widening is the
minimum.

---

## 3. No default `asset_status` rows ⬜ open

**Symptom:** on a fresh company, the status dropdown in the client's vehicle
create/edit form is empty, which also blocks asset creation (`asset.status_id`
is what the form sets).

**Cause:** `asset_status` is company-scoped (`UNIQUE (company_id, name)`, FK to
`company ON DELETE CASCADE`) and nothing seeds it. `POST /api/v1/companies`
creates the company row only. `GET /api/v1/asset-statuses` correctly returns
`{data: []}` — the API is behaving as written; the gap is that a company is
created in an unusable state.

**Worked around** by seeding directly in dev:

```sql
INSERT INTO asset_status (company_id, name, color_code)
SELECT c.id, v.name, v.color
FROM company c
CROSS JOIN (VALUES
  ('Active','#28a745'), ('In Shop','#f0ad4e'), ('Out of Service','#dc3545'),
  ('In Transit','#0d6efd'), ('Sold','#6c757d')
) AS v(name, color)
ON CONFLICT (company_id, name) DO NOTHING;
```

Those five names/colors are **the frontend's invention**, not a ported
convention — Django had no canonical default set either. They need a product
decision before they mean anything.

**Recommended fix:** seed a default status set inside the `POST /api/v1/companies`
transaction, so a company is never created unusable. `POST/PUT/DELETE
/api/v1/asset-statuses` already exist and work, so a company can still customize
afterward — the missing piece is bootstrap, not CRUD.

**Same question applies to** `asset_type` and `catalog_option`, which are also
empty on a fresh company. Neither currently blocks the client (`asset_type` has
no frontend reader; `catalog_option` rows are created on demand), so they're
lower priority — but the bootstrap story should cover them deliberately rather
than by accident.

---

## 4. Replaced/cleared uploads leak objects ⬜ open

Setting a new `photo`/`logo`/`file` overwrites the string column; the previously
referenced object is never removed from storage. `storage.Storage` has
`Delete(ctx, key)` and **no caller anywhere in the codebase**.

Storage therefore grows monotonically, including for records that were themselves
deleted. Needs a product/infra decision:

- delete-on-replace inside the owning handler (simple, but risks deleting an
  object another row still references — nothing enforces single-ownership today), or
- a sweeper that lists bucket keys and drops those unreferenced by any of the five
  columns in #2, or
- a bucket lifecycle policy, if orphan retention is acceptable.

Not urgent at current volume, but it never self-corrects.

---

## 5. `POST /api/v1/uploads` does no validation ⬜ open

`UploadHandler.Upload` (`internal/http/handler/upload.go`) accepts any multipart
`file` of any size and any content type, and passes
`header.Header.Get("Content-Type")` — **client-supplied, unverified** — straight
through to storage as the object's content type.

The type/size limits users experience (PNG/JPEG/WebP, 5 MB) are enforced only in
the frontend and are trivially bypassed by calling the endpoint directly. Since
the bucket is public-read (#7), a stored object is then served to anyone with the
URL under an attacker-chosen content type.

**Suggested:** enforce a max size (e.g. `http.MaxBytesReader`), sniff the real
type with `http.DetectContentType` on the first 512 bytes, and reject anything
outside an allowlist rather than trusting the header.

---

## 6. `/api/v1/companies` is neither tenant-scoped nor admin-gated ⬜ open

Every other `/api/v1/*` route scopes off the JWT's `company_id` claim, but
`/api/v1/companies` does not: any authenticated employee, from any company, can
list, read, update, and **delete** any company row.

This was found during the frontend's admin work and is documented in detail at
`fleet-admin-app/docs/backend-migration/companies/README.md`. Repeated here
because it is a backend authorization gap, not a frontend concern — and the
frontend cannot paper over it.

Note the blast radius: `asset_status`, `asset_type` and others are
`FK … ON DELETE CASCADE` to `company`, so an unauthorized `DELETE
/api/v1/companies/{id}` destroys another tenant's data.

---

## 7. Public bucket: uploads are world-readable across tenants ⬜ decision

With `STORAGE_MINIO_PUBLIC=true` an anonymous read policy is applied to the whole
bucket, so any object is readable by anyone with the URL, regardless of tenant.
The `uploads/{company_id}/` key namespacing is organizational, not a security
boundary; the random token makes URLs unguessable, not private.

That is a reasonable trade for vehicle photos and company logos. It is **not**
reasonable for inspection submissions or fuel receipts, which may contain
personal or commercially sensitive information — and those columns (#2) point at
the same bucket.

`MinIO.PresignedGetURL()` already exists and is unused. Recommendation: decide
per-bucket (public assets vs. private documents) before the inspection and fuel
features go live, rather than after.
