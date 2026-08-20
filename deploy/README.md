# Production runbook

The API runs on an IONOS VPS at **74.208.73.82**, served at
**https://go-logistics.jesuslab135.com**. Push to `main` and GitHub Actions
tests, builds an image, and rolls the stack. This file is what you need when
something goes wrong at 2am.

## The one thing to know first

**This VPS also hosts an unrelated production site**, `app.jesuslab135.com`,
from the `sga` Compose project in `/opt/sga`. Its nginx owns ports 80 and 443
and therefore terminates TLS for *both* sites. Nothing here may take those
ports, and edits under `/opt/sga` affect a live site that is not this one.

```
                       :80 / :443
                     ┌──────────────┐
   internet ────────▶│  sga_nginx   │  (project sga, /opt/sga)
                     └──────┬───────┘
            app.jesuslab135.com │ go-logistics.jesuslab135.com
              ┌────────────────┴────────────────┐
              ▼                                 ▼
      sga frontend/backend            /fleet/ ─▶ fleet-minio:9000
         (untouched)                  /      ─▶ fleet-api:8080

                                      project fleet, /opt/fleet
                                      api · db · minio · migrate · cli
                                      (no published ports except the
                                       loopback MinIO console)
```

The two projects meet on one Docker network, `edge`, which is external to both
and was created once by hand (`docker network create edge`). `sga_nginx` and
the fleet `api`/`minio` services all attach to it; the fleet services are
aliased `fleet-api` and `fleet-minio` there so the shared namespace cannot
collide with a future third stack.

## Server layout

```
/opt/fleet/
├── docker-compose.yml     shipped by the deploy (repo: deploy/docker-compose.prod.yml)
├── .env                   SERVER-OWNED, chmod 600, never deployed, never in git
├── migrations/            shipped by the deploy (repo: internal/db/migrations)
├── deploy.sh              shipped by the deploy
├── backup.sh              shipped by the deploy
└── backups/               nightly dumps + MinIO tarballs, 7-day rotation

/opt/sga/nginx/app.conf    BOTH vhosts; ours sits between the
                           "# >>> BEGIN go-logistics" / "# <<< END" markers
                           (versioned copy: deploy/nginx/go-logistics.conf)
```

Secrets live in exactly two places: `/opt/fleet/.env` at runtime, and GitHub
Actions secrets for the pipeline. The deploy bundle deliberately excludes
`.env`, so a deploy can never overwrite a secret.

## Pipeline

| Kind | Name | Value |
|---|---|---|
| Secret | `DEPLOY_HOST` | `74.208.73.82` |
| Secret | `DEPLOY_USER` | `deploy` |
| Secret | `DEPLOY_SSH_KEY` | private half of the `fleet-deploy` ed25519 key |
| Variable | `API_DOMAIN` | `go-logistics.jesuslab135.com` |

There is deliberately **no registry PAT**. The deploy job holds `packages:
read` on its own `GITHUB_TOKEN` and pipes that to `docker login` on the server;
it expires with the run, so there is nothing on the VPS to rotate or leak.

A deploy is: `main` → `test` → build and push
`ghcr.io/jesuslab135/go-logistics:<sha>` (plus `:latest`) → scp the bundle to
`/opt/fleet` → `deploy.sh <sha-tag>` → the job fails unless
`https://go-logistics.jesuslab135.com/healthz` answers within 60s.

`deploy.sh` writes the SHA into `.env`, so a bare `docker compose up -d` after
a reboot re-runs the same release rather than drifting to whatever `:latest`
points at.

### Rollback

```sh
ssh deploy@74.208.73.82 '/opt/fleet/deploy.sh ghcr.io/jesuslab135/go-logistics:<old-sha>'
curl -fsS https://go-logistics.jesuslab135.com/healthz
```

Get `<old-sha>` from `git log` or the Actions run list.

The registry credential the pipeline leaves behind expires with its run, so a
by-hand rollback cannot pull. `deploy.sh` handles that: a denied pull is only
fatal when the image is not already on disk, and for a rollback it usually is.
Only the three most recent tags are kept, though — going further back needs
`docker login ghcr.io -u <user>` with an ad-hoc classic PAT scoped
`read:packages`. Do not store one on the server.

Rollback runs the old image against the **current** schema. Migrations are
forward-only, so a rollback across a destructive migration needs a restore, not
a redeploy.

## Day-to-day

```sh
# logs
ssh deploy@74.208.73.82 'cd /opt/fleet && docker compose logs -f api'
ssh deploy@74.208.73.82 'cd /opt/fleet && docker compose logs migrate'

# state
ssh deploy@74.208.73.82 'cd /opt/fleet && docker compose ps'

# psql
ssh deploy@74.208.73.82 'cd /opt/fleet && docker compose exec db psql -U postgres -d fleet'

# change a config value (CORS_ORIGINS, TTLs, …)
ssh deploy@74.208.73.82
nano /opt/fleet/.env && cd /opt/fleet && docker compose up -d api

# MinIO console — never exposed publicly; tunnel to the loopback bind
ssh -L 9001:127.0.0.1:9001 deploy@74.208.73.82
#   then open http://localhost:9001 with MINIO_ROOT_USER / MINIO_ROOT_PASSWORD
```

The admin CLI ships in the same image as a second binary, because the server
has no Go toolchain:

```sh
cd /opt/fleet
docker compose run --rm cli setpass --email <email> --password <password>
docker compose run --rm cli bootstrap --email <email> --company-id <id>
```

### First-boot bootstrap (already done — kept for a rebuild)

`POST /companies` requires an authenticated account-owner, so the very first
identity cannot come from the API. Insert it once, then hand it to the CLI:

```sh
cd /opt/fleet
docker compose exec db psql -U postgres -d fleet -c "
INSERT INTO company (name,tax_id,address,created_at,phone,email,website,city,region,postal_code,country,timezone,currency,system_of_measurement)
VALUES ('Go Logistics','','',now(),'','','','','','','MX','America/Mexico_City','MXN','metric');
INSERT INTO employee (default_company_id,first_name,last_name,employee_id,is_active,email,mobile_phone,work_phone,job_title,is_technician,is_vehicle_operator,is_account_owner,license_class,license_number,license_state,street_address,city,region,postal_code,country,updated_at,password_hash)
VALUES (1,'Esteban','Olmos','E1',true,'<admin-email>','','','Administrator',false,false,true,'','','','','','','','',now(),'');
"
docker compose run --rm cli bootstrap --email '<admin-email>' --company-id 1
docker compose run --rm cli setpass  --email '<admin-email>' --password '<password>'
```

## Backups

`backup.sh` runs nightly at 03:15 from the `deploy` user's crontab: a
`pg_dump -Fc` plus a tar of the MinIO volume, both rotated after 7 days, into
`/opt/fleet/backups/`.

**They sit on the same disk as the thing they protect.** A failed volume takes
both. Copy them off the box — even a laptop cron running
`scp deploy@74.208.73.82:/opt/fleet/backups/\* .` — before treating them as a
real backup.

### Restore: database

```sh
cd /opt/fleet
# always rehearse into a throwaway database first
docker compose exec -T db createdb -U postgres fleet_restore_test
docker compose exec -T db pg_restore -U postgres -d fleet_restore_test < backups/fleet-<stamp>.dump
docker compose exec db psql -U postgres -d fleet_restore_test -c 'SELECT count(*) FROM employee;'
docker compose exec db dropdb -U postgres fleet_restore_test

# for real, over the live database
docker compose stop api
docker compose exec -T db pg_restore -U postgres --clean --if-exists -d fleet < backups/fleet-<stamp>.dump
docker compose start api
```

### Restore: MinIO objects

```sh
cd /opt/fleet
docker compose stop api minio
docker run --rm -v fleet_minio_data:/data -v "$(pwd)/backups:/backup" busybox \
    sh -c 'rm -rf /data/* && tar xzf /backup/minio-<stamp>.tar.gz -C /data'
docker compose start minio api
```

## TLS

One Let's Encrypt certificate covers `go-logistics.jesuslab135.com`, issued
through the **sga stack's existing certbot** container (webroot
`/var/www/certbot`), which renews everything in `/etc/letsencrypt` on a 12h
loop. Nothing to do per-site.

Renewal writes new files but does not signal nginx. nginx only reloads
certificates on reload, so a renewed cert is not served until the next reload —
worth a `docker exec sga_nginx nginx -s reload` after a renewal window, or a
weekly cron doing it unconditionally.

To reissue by hand:

```sh
cd /opt/sga
docker compose run --rm --entrypoint certbot certbot certonly \
    --webroot -w /var/www/certbot -d go-logistics.jesuslab135.com \
    --email jefeing.esteban@gmail.com --agree-tos --no-eff-email --non-interactive
```

## Editing the edge vhost — read this before you do

`/opt/sga/nginx/app.conf` is bind-mounted into `sga_nginx` **as a single
file**, which means the container is bound to that file's *inode*, not its
path. Any editor that writes a replacement and renames it over the original —
`sed -i`, vim's default, most IDEs — silently leaves the container reading the
old, now-unlinked file. `nginx -t` passes, the reload succeeds, and nothing
changes.

Edit in place instead, preserving the inode:

```sh
# safe: truncates and rewrites the same inode
cat /tmp/new-app.conf > /opt/sga/nginx/app.conf
docker exec sga_nginx nginx -t && docker exec sga_nginx nginx -s reload

# if you already broke the binding, this is the only fix (≈2s of downtime
# on app.jesuslab135.com):
docker restart sga_nginx
```

Verify the container actually sees your edit:

```sh
stat -c '%i %s' /opt/sga/nginx/app.conf
docker exec sga_nginx stat -c '%i %s' /etc/nginx/conf.d/app.conf   # must match
```

The vhost resolves its upstreams through Docker's embedded DNS
(`resolver 127.0.0.11` plus a variable in `proxy_pass`) rather than at config
load. That is deliberate: every deploy replaces the `api` container with a new
IP, and a statically resolved upstream would 502 until someone reloaded nginx
by hand.

## SSH

Key-only, enforced by `/etc/ssh/sshd_config.d/00-hardening.conf`. The `00`
prefix is load-bearing: cloud-init ships `50-cloud-init.conf` with
`PasswordAuthentication yes`, OpenSSH honours the **first** value it reads, and
a `99-` file loses. Revert by deleting the drop-in and
`systemctl restart ssh`; the IONOS Cloud Panel console is the way back in if
SSH is ever lost entirely.

`deploy` is in the `docker` group, which is root-equivalent on this host — the
pipeline's SSH key is effectively a root key. Treat `DEPLOY_SSH_KEY` as such.

## Swagger

The interactive docs are served at
<https://go-logistics.jesuslab135.com/swagger/index.html>, enabled by
`SWAGGER_ENABLED=true` in `/opt/fleet/.env`. Production defaults to off, so
this is a deliberate opt-in.

It is unauthenticated and lists every route, model and field. That is not a
vulnerability on its own — the endpoints still require a token — but it does
hand an attacker a map. To close it again:

```sh
ssh deploy@74.208.73.82
sed -i 's/^SWAGGER_ENABLED=.*/SWAGGER_ENABLED=false/' /opt/fleet/.env
cd /opt/fleet && docker compose up -d api
```

The spec is generated with an empty `host`, so the UI targets whatever origin
serves it and "Try it out" hits production for real. Treat it as a live console,
not a sandbox.

## Known follow-ups

- `CORS_ORIGINS` is set to the backend's own origin. A browser never sends
  that as its `Origin`, so the middleware echoes no `Access-Control-Allow-Origin`
  today and every browser client is blocked — not cosmetic. Set it to the
  frontend's real origin (plus `http://localhost:5173` if browser-testing
  against this deployment), then `docker compose up -d api`.
- Backups are on-box only (see above).
- The IONOS Cloud Panel has its own firewall policy in front of the OS `ufw`.
  Both currently allow 22/80/443 only; a port opened in `ufw` alone will still
  be blocked upstream.
