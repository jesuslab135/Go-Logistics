# Go-Logistics IONOS Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Continuous deployment of the Go-Logistics API (Go + Postgres + MinIO) to a provisioned IONOS VPS: push to `main` → GitHub Actions tests, builds and pushes a Docker image to GHCR, then SSHes into the VPS and rolls the stack with automatic HTTPS.

**Architecture:** One Docker Compose project on the VPS at `/opt/fleet`: Caddy (ports 80/443, Let's Encrypt) reverse-proxies `api.<domain>` → the API container and `media.<domain>` → MinIO; Postgres and MinIO are internal-only with named volumes; a `migrate` one-shot job runs golang-migrate before the API starts (same pattern the dev compose already uses). GitHub Actions builds the image once, tags it with the commit SHA, and the server always runs an immutable SHA tag — rollback is redeploying the previous SHA.

**Tech Stack:** Docker + Compose v2, Caddy 2, GHCR, GitHub Actions (plain `ssh`/`scp`, no third-party deploy actions), golang-migrate, Ubuntu LTS on IONOS VPS.

**Spec:** User decisions recorded 2026-08-14: IONOS VPS already provisioned with root SSH · backend only (frontend deploys separately) · a domain is available → Caddy + Let's Encrypt · production only, deployed from `main`.

## Global Constraints

- Repo: `github.com/jesuslab135/Go-Logistics`, module `fleet`, Go from `go.mod` (1.26.x). Image: `ghcr.io/jesuslab135/go-logistics`.
- Commits authored solely by the user (`jefeing.esteban@gmail.com`), NO co-author trailers; per-commit gate `go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...`.
- The 22 local commits (`9d201e5..eb9ea6a`) are **not pushed**. The pipeline builds from GitHub, so pushing `main` is a hard prerequisite for Task 6 — and pushing is the user's explicit call.
- Server `.env` is created once by hand and **never written by the deploy** — the deploy bundle deliberately excludes it. Secrets live in two places only: the server `.env` (runtime) and GitHub Actions secrets (pipeline).
- `APP_ENV=production` already disables Swagger UI and sets Gin release mode; prod uses `JWT_ACCESS_TTL=15m` per `.env.example`'s PROD block.
- `STORAGE_MINIO_PUBLIC_URL` must end with the bucket name (`/fleet`) — startup validates this.
- Compose interpolation builds derived values (`DATABASE_URL`) in the compose file from `.env` raw secrets; never put `${VAR}` references *inside* `.env` values (env_file passes them as literals).
- Placeholders the executor must substitute everywhere they appear: `example.com` → the real domain; `203.0.113.10` → the VPS IP.

## File Structure

```
Dockerfile                              (modify — also ship /cli)
deploy/docker-compose.prod.yml          (new — caddy, api, db, minio, migrate)
deploy/Caddyfile                        (new — two vhosts, env-substituted domains)
deploy/.env.production.example          (new — server .env template)
deploy/deploy.sh                        (new — pin image SHA, pull, up)
deploy/backup.sh                        (new — nightly pg_dump + minio tar, 7-day rotation)
deploy/README.md                        (new — runbook: secrets, first boot, rollback, restore)
.github/workflows/ci.yml                (new — PR gate: build/vet/fmt/test + generated-code drift)
.github/workflows/deploy.yml            (new — main: test → build+push → ssh deploy → health gate)
README.md                               (modify — deployment section pointer)
```

Tasks 1–4 are repo commits and need no server. Tasks 5–7 touch the VPS. Task 6 requires `main` pushed.

---

### Task 1: Ship the admin CLI in the production image

The distroless image contains only `/api`. On the server there is no Go toolchain and no source, so `cmd/cli` (`setpass`, `bootstrap`) — required to create the first login — must ride in the same image as a second static binary.

**Files:**
- Modify: `Dockerfile`

**Interfaces:**
- Produces: image entrypoints `/api` (default) and `/cli` (via `--entrypoint /cli`). Task 2's compose `cli` service and Task 6's bootstrap commands rely on `/cli` existing at that exact path.

- [ ] **Step 1: Add the second binary to the build**

Replace the build `RUN` and final-stage `COPY` in `Dockerfile` so it reads:

```dockerfile
# Build static binaries, then ship them on a minimal base image.
FROM golang:1.26 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/cli ./cmd/cli

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
# Admin CLI (setpass, bootstrap). The server has no Go toolchain, so the only
# way to run it in production is from this image: --entrypoint /cli
COPY --from=build /out/cli /cli
EXPOSE 8080
ENTRYPOINT ["/api"]
```

- [ ] **Step 2: Verify both binaries run**

```sh
docker build -t fleet-test .
docker run --rm --entrypoint /cli fleet-test; echo "exit=$?"
```
Expected: the CLI usage text (`usage: fleet-cli setpass ...`) and `exit=2` (its no-args exit code). Then `docker run --rm -e APP_ENV=x fleet-test` should fail fast complaining about `DATABASE_URL` — proving `/api` still runs.

- [ ] **Step 3: Commit**

```bash
git add Dockerfile
git commit -m "build: ship the admin CLI beside the API in the production image"
```

---

### Task 2: Production stack — compose, Caddy, env template, deploy/backup scripts

**Files:**
- Create: `deploy/docker-compose.prod.yml`, `deploy/Caddyfile`, `deploy/.env.production.example`, `deploy/deploy.sh`, `deploy/backup.sh`

**Interfaces:**
- Consumes: image `${API_IMAGE}` with `/api` + `/cli` (Task 1).
- Produces: the deploy bundle layout Task 4's workflow scp's to `/opt/fleet` — `docker-compose.prod.yml` is copied **as `docker-compose.yml`** on the server so bare `docker compose` works; `deploy.sh <image-ref>` is the only entrypoint the pipeline calls. `.env` keys consumed here: `API_DOMAIN`, `MEDIA_DOMAIN`, `ACME_EMAIL`, `API_IMAGE`, `POSTGRES_PASSWORD`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`, plus every app var.

- [ ] **Step 1: Write `deploy/docker-compose.prod.yml`**

```yaml
# Production stack for the IONOS VPS. Lives at /opt/fleet/docker-compose.yml.
#
# Only Caddy touches the host network (80/443). Postgres and MinIO are
# reachable solely on the compose network; admin access goes through
# `docker exec` or an SSH tunnel to the loopback-bound MinIO console.
#
# Derived values (DATABASE_URL, the MinIO endpoint) are assembled HERE from the
# raw secrets in .env — compose interpolates ${...} in this file; it does NOT
# expand references inside .env values.
name: fleet

services:
  caddy:
    image: caddy:2
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    environment:
      API_DOMAIN: ${API_DOMAIN}
      MEDIA_DOMAIN: ${MEDIA_DOMAIN}
      ACME_EMAIL: ${ACME_EMAIL}
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data      # certificates — losing this re-issues, rate limits apply
      - caddy_config:/config

  db:
    image: postgres:16
    restart: unless-stopped
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: fleet
    volumes:
      - pg_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d fleet"]
      interval: 5s
      timeout: 5s
      retries: 10

  minio:
    image: minio/minio:latest
    restart: unless-stopped
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD}
      # Presigned/console URLs advertise the public hostname, not "minio".
      MINIO_SERVER_URL: https://${MEDIA_DOMAIN}
    command: server /data --console-address ":9001"
    ports:
      # Console on loopback only — reach it with:
      #   ssh -L 9001:127.0.0.1:9001 deploy@<vps>   then http://localhost:9001
      - "127.0.0.1:9001:9001"
    volumes:
      - minio_data:/data
    healthcheck:
      test: ["CMD-SHELL", "curl -sf http://localhost:9000/minio/health/live || exit 1"]
      interval: 5s
      timeout: 5s
      retries: 10

  migrate:
    image: migrate/migrate:v4.17.1
    depends_on:
      db:
        condition: service_healthy
    volumes:
      # The deploy bundle ships the repo's internal/db/migrations as ./migrations.
      - ./migrations:/migrations:ro
    entrypoint:
      [
        "migrate",
        "-path", "/migrations",
        "-database", "postgres://postgres:${POSTGRES_PASSWORD}@db:5432/fleet?sslmode=disable",
        "up",
      ]
    restart: "no"

  api:
    image: ${API_IMAGE}
    restart: unless-stopped
    env_file:
      - .env
    environment:
      DATABASE_URL: postgres://postgres:${POSTGRES_PASSWORD}@db:5432/fleet?sslmode=disable
      STORAGE_MINIO_ENDPOINT: minio:9000
      STORAGE_MINIO_ACCESS_KEY: ${MINIO_ROOT_USER}
      STORAGE_MINIO_SECRET_KEY: ${MINIO_ROOT_PASSWORD}
      STORAGE_MINIO_USE_SSL: "false"   # internal hop; TLS terminates at Caddy
      # Must end with the bucket name: the public-read policy covers /fleet/*.
      STORAGE_MINIO_PUBLIC_URL: https://${MEDIA_DOMAIN}/fleet
    depends_on:
      db:
        condition: service_healthy
      minio:
        condition: service_healthy
      migrate:
        condition: service_completed_successfully

  # Admin CLI from the same image (distroless, so no shell — entrypoint only):
  #   docker compose run --rm cli setpass --email a@b.c --password secret123
  cli:
    image: ${API_IMAGE}
    entrypoint: ["/cli"]
    env_file:
      - .env
    environment:
      DATABASE_URL: postgres://postgres:${POSTGRES_PASSWORD}@db:5432/fleet?sslmode=disable
    depends_on:
      db:
        condition: service_healthy
    profiles: ["tools"]

volumes:
  caddy_data:
  caddy_config:
  pg_data:
  minio_data:
```

- [ ] **Step 2: Write `deploy/Caddyfile`**

```caddyfile
# Caddy substitutes {$VAR} from the container environment at startup.
{
	email {$ACME_EMAIL}
}

{$API_DOMAIN} {
	encode gzip
	reverse_proxy api:8080
}

# Public object storage. App-stored URLs look like
# https://{$MEDIA_DOMAIN}/fleet/uploads/<company>/<token>-<name>.
{$MEDIA_DOMAIN} {
	reverse_proxy minio:9000
}
```

- [ ] **Step 3: Write `deploy/.env.production.example`**

```sh
# Template for /opt/fleet/.env on the VPS. Copy, fill, chmod 600.
# This file is created ONCE by hand and never touched by deploys.
#
# Generate secrets:  openssl rand -hex 32   (JWT)
#                    openssl rand -hex 24   (passwords)

# ---- deployment ----------------------------------------------------------
API_DOMAIN=api.example.com
MEDIA_DOMAIN=media.example.com
ACME_EMAIL=jefeing.esteban@gmail.com
# Written by deploy.sh on every deploy — leave the placeholder on first boot.
API_IMAGE=ghcr.io/jesuslab135/go-logistics:latest

# ---- infrastructure secrets ---------------------------------------------
POSTGRES_PASSWORD=<openssl rand -hex 24>
MINIO_ROOT_USER=fleet-admin
MINIO_ROOT_PASSWORD=<openssl rand -hex 24>

# ---- application (mirrors .env.example PROD block) -----------------------
APP_ENV=production
HTTP_ADDR=:8080
# The frontend's exact origin(s), comma-separated — NOT "*" in production.
CORS_ORIGINS=https://app.example.com

JWT_SECRET=<openssl rand -hex 32>
JWT_ISSUER=fleet
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

STORAGE_BACKEND=minio
STORAGE_MINIO_BUCKET=fleet
STORAGE_MINIO_REGION=us-east-1
STORAGE_MINIO_PUBLIC=true

UPLOAD_MAX_BYTES=5242880
UPLOAD_ALLOWED_TYPES=image/jpeg,image/png,image/gif,image/webp,application/pdf
UPLOAD_THUMBNAIL_MAX_DIM=320

PAGE_DEFAULT_SIZE=25
PAGE_MAX_SIZE=100
DASHBOARD_UPCOMING_DAYS=30
```

(`DATABASE_URL` and the `STORAGE_MINIO_ENDPOINT/ACCESS_KEY/SECRET_KEY/USE_SSL/PUBLIC_URL` group are intentionally absent — the compose file derives them, and its `environment:` block overrides `env_file` anyway.)

- [ ] **Step 4: Write `deploy/deploy.sh`**

```bash
#!/usr/bin/env bash
# Roll the stack to a specific image. The pipeline calls this with an
# immutable SHA tag; recording it in .env means a plain `docker compose up -d`
# after a reboot re-runs the same release, and rollback is re-running this
# script with the previous SHA.
set -euo pipefail

IMAGE="${1:?usage: deploy.sh <image-ref>}"
cd "$(dirname "$0")"

if grep -q '^API_IMAGE=' .env; then
    sed -i "s|^API_IMAGE=.*|API_IMAGE=${IMAGE}|" .env
else
    echo "API_IMAGE=${IMAGE}" >> .env
fi

docker compose pull --quiet
docker compose up -d --remove-orphans
docker image prune -f >/dev/null

echo "deployed ${IMAGE}"
```

- [ ] **Step 5: Write `deploy/backup.sh`**

```bash
#!/usr/bin/env bash
# Nightly backups: a compressed pg_dump plus a tar of the MinIO volume,
# rotated after 7 days. Run from cron as the deploy user (docker group).
# Restore commands are in deploy/README.md.
set -euo pipefail

cd "$(dirname "$0")"
BACKUP_DIR="./backups"
STAMP="$(date +%Y%m%d-%H%M%S)"
mkdir -p "${BACKUP_DIR}"

docker compose exec -T db pg_dump -U postgres -Fc fleet \
    > "${BACKUP_DIR}/fleet-${STAMP}.dump"

docker run --rm \
    -v fleet_minio_data:/data:ro \
    -v "$(pwd)/${BACKUP_DIR}:/backup" \
    busybox tar czf "/backup/minio-${STAMP}.tar.gz" -C /data .

find "${BACKUP_DIR}" -name 'fleet-*.dump'    -mtime +7 -delete
find "${BACKUP_DIR}" -name 'minio-*.tar.gz' -mtime +7 -delete

echo "backup ${STAMP} done: $(du -sh ${BACKUP_DIR} | cut -f1) total"
```

- [ ] **Step 6: Verify the compose file renders**

```sh
cd deploy && cp .env.production.example .env \
 && sed -i 's/<openssl rand -hex 24>/x/; s/<openssl rand -hex 32>/x/' .env 2>/dev/null; \
 docker compose -f docker-compose.prod.yml config >/dev/null && echo RENDER-OK; \
 rm .env; cd ..
```
Expected: `RENDER-OK` (interpolation resolves, no undefined-variable warnings). Also `git update-index --chmod=+x deploy/deploy.sh deploy/backup.sh` so the scripts are executable after scp.

- [ ] **Step 7: Commit**

```bash
git add deploy/docker-compose.prod.yml deploy/Caddyfile deploy/.env.production.example deploy/deploy.sh deploy/backup.sh
git update-index --chmod=+x deploy/deploy.sh deploy/backup.sh
git commit -m "deploy: production compose stack with Caddy TLS and admin tooling"
```

---

### Task 3: CI workflow (PR gate)

**Files:**
- Create: `.github/workflows/ci.yml`

**Interfaces:**
- Produces: a `test` job shape Task 4 duplicates in its own workflow (kept as two files so the PR gate and the deploy pipeline read independently).

- [ ] **Step 1: Write `.github/workflows/ci.yml`**

```yaml
name: CI

on:
  pull_request:
  push:
    branches-ignore: [main]   # main is covered by deploy.yml's test job

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Build
        run: go build ./...

      - name: Vet
        run: go vet ./...

      - name: Format
        run: test -z "$(gofmt -l .)" || { gofmt -l .; exit 1; }

      - name: Test
        run: go test ./...

      # The repo commits generated code; regen must be byte-identical or
      # someone hand-edited generated output / forgot to regenerate.
      - name: Generated-code drift (sqlc)
        run: |
          go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0
          sqlc generate
          git diff --exit-code internal/db/gen

      - name: Generated-code drift (swag)
        run: |
          go install github.com/swaggo/swag/cmd/swag@v1.16.4
          swag init -g main.go \
            -d ./cmd/api,./internal/http/handler,./internal/http/dto,./internal/auth,./internal/platform/storage \
            --parseDependency --parseInternal -o docs
          git diff --exit-code docs
```

- [ ] **Step 2: Verify locally what CI will run**

```sh
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./... && echo LOCAL-GATE-OK
```
Expected: `LOCAL-GATE-OK`. (The drift steps were proven byte-stable during the BR work; the live run happens when a branch is pushed after Task 6's push.)

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: build, vet, format, test and generated-code drift gate"
```

---

### Task 4: Deploy workflow (main → GHCR → VPS)

**Files:**
- Create: `.github/workflows/deploy.yml`

**Interfaces:**
- Consumes: `deploy/deploy.sh <image-ref>` and the bundle layout (Task 2); GitHub **secrets** `DEPLOY_SSH_KEY`, `DEPLOY_HOST`, `DEPLOY_USER`, `GHCR_PULL_TOKEN` and **variable** `API_DOMAIN` (created in Task 5).
- Produces: images `ghcr.io/jesuslab135/go-logistics:{<sha>,latest}`; a deploy that ends red unless `https://<API_DOMAIN>/healthz` answers.

- [ ] **Step 1: Write `.github/workflows/deploy.yml`**

```yaml
name: Deploy

on:
  push:
    branches: [main]
  workflow_dispatch:

# Never two deploys interleaved; queue instead of cancel so every main commit lands.
concurrency:
  group: deploy-production
  cancel-in-progress: false

env:
  IMAGE: ghcr.io/jesuslab135/go-logistics

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - run: go build ./...
      - run: go vet ./...
      - run: test -z "$(gofmt -l .)"
      - run: go test ./...

  build-push:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - uses: docker/setup-buildx-action@v3

      - uses: docker/build-push-action@v6
        with:
          context: .
          platforms: linux/amd64
          push: true
          tags: |
            ${{ env.IMAGE }}:${{ github.sha }}
            ${{ env.IMAGE }}:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy:
    needs: build-push
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4

      - name: Set up SSH
        run: |
          mkdir -p ~/.ssh
          printf '%s\n' "${{ secrets.DEPLOY_SSH_KEY }}" > ~/.ssh/id_ed25519
          chmod 600 ~/.ssh/id_ed25519
          ssh-keyscan -H "${{ secrets.DEPLOY_HOST }}" >> ~/.ssh/known_hosts

      - name: Ship the deploy bundle
        # Everything the server needs EXCEPT .env, which is server-owned.
        run: |
          HOST="${{ secrets.DEPLOY_USER }}@${{ secrets.DEPLOY_HOST }}"
          ssh "$HOST" 'mkdir -p /opt/fleet/migrations'
          scp deploy/docker-compose.prod.yml "$HOST:/opt/fleet/docker-compose.yml"
          scp deploy/Caddyfile deploy/deploy.sh deploy/backup.sh "$HOST:/opt/fleet/"
          scp internal/db/migrations/* "$HOST:/opt/fleet/migrations/"

      - name: Roll the stack
        run: |
          HOST="${{ secrets.DEPLOY_USER }}@${{ secrets.DEPLOY_HOST }}"
          ssh "$HOST" "
            set -euo pipefail
            echo '${{ secrets.GHCR_PULL_TOKEN }}' | docker login ghcr.io -u jesuslab135 --password-stdin
            chmod +x /opt/fleet/deploy.sh /opt/fleet/backup.sh
            /opt/fleet/deploy.sh '${{ env.IMAGE }}:${{ github.sha }}'
          "

      - name: Health gate
        run: |
          for i in $(seq 1 30); do
            if curl -fsS "https://${{ vars.API_DOMAIN }}/healthz" >/dev/null 2>&1; then
              echo "healthy after ${i} attempts"; exit 0
            fi
            sleep 2
          done
          echo "API never became healthy — check: ssh ${{ secrets.DEPLOY_USER }}@${{ secrets.DEPLOY_HOST }} 'cd /opt/fleet && docker compose logs api migrate'"
          exit 1
```

- [ ] **Step 2: Verify the workflow parses**

```sh
docker run --rm -v "$(pwd)":/repo -w /repo rhysd/actionlint:latest .github/workflows/deploy.yml .github/workflows/ci.yml || echo "actionlint unavailable — rely on GitHub's parser on first push"
```
Expected: no findings (or the fallback message; GitHub validates on push either way).

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/deploy.yml
git commit -m "ci: continuous deployment to the IONOS VPS via GHCR and SSH"
```

---

### Task 5: One-time server preparation (manual, on the VPS)

No repo changes. Every command is run once; each step ends with its own check.

- [ ] **Step 1: DNS.** In the domain's DNS panel create two `A` records → the VPS IP: `api.example.com` and `media.example.com`. Verify: `nslookup api.example.com` and `nslookup media.example.com` both return `203.0.113.10`. (Caddy cannot obtain certificates until these resolve.)

- [ ] **Step 2: IONOS firewall.** IONOS VPS traffic passes an **external firewall policy in the IONOS Cloud Panel** in addition to the OS firewall — a classic gotcha. In the panel (Server → Network → Firewall Policies) allow inbound TCP **22, 80, 443** and nothing else. Verify later at Step 6.

- [ ] **Step 3: OS + Docker.** SSH in as root (assumes Ubuntu LTS):

```sh
apt-get update && apt-get -y upgrade
curl -fsSL https://get.docker.com | sh
docker --version && docker compose version
```
Expected: Docker ≥ 24 and Compose v2 print versions.

- [ ] **Step 4: Deploy user + SSH keys.**

```sh
adduser --disabled-password --gecos "" deploy
usermod -aG docker deploy
mkdir -p /home/deploy/.ssh && chmod 700 /home/deploy/.ssh
# Generate the PIPELINE's keypair locally (on the laptop), NOT on the server:
#   ssh-keygen -t ed25519 -C fleet-deploy -f fleet_deploy_key -N ""
# Paste fleet_deploy_key.pub (and your own personal public key) here:
nano /home/deploy/.ssh/authorized_keys
chmod 600 /home/deploy/.ssh/authorized_keys && chown -R deploy:deploy /home/deploy/.ssh
```
Then harden `sshd` — in `/etc/ssh/sshd_config` set `PasswordAuthentication no` and `PermitRootLogin prohibit-password`, then `systemctl restart ssh`. Verify **from the laptop before closing the root session**: `ssh -i fleet_deploy_key deploy@203.0.113.10 docker ps` prints an (empty) container table.

- [ ] **Step 5: OS firewall.**

```sh
ufw allow OpenSSH && ufw allow 80/tcp && ufw allow 443/tcp && ufw --force enable && ufw status
```
Expected: exactly 22, 80, 443 allowed. (Docker's published ports bypass ufw via iptables — that's why the compose file publishes nothing except Caddy and the loopback-bound MinIO console.)

- [ ] **Step 6: Server `.env`.** As `deploy`:

```sh
mkdir -p /opt/fleet   # (as root: mkdir -p /opt/fleet && chown deploy:deploy /opt/fleet)
cd /opt/fleet && nano .env      # paste deploy/.env.production.example, fill every value:
#   API_DOMAIN / MEDIA_DOMAIN / ACME_EMAIL / CORS_ORIGINS → real values
#   POSTGRES_PASSWORD / MINIO_ROOT_PASSWORD → openssl rand -hex 24
#   JWT_SECRET → openssl rand -hex 32
chmod 600 .env
```
Verify: `ls -l /opt/fleet/.env` shows `-rw-------` owned by `deploy`.

- [ ] **Step 7: GHCR pull token + GitHub secrets.** On github.com (account `jesuslab135`): Settings → Developer settings → Personal access tokens → **classic** token, scope `read:packages` only, no expiry shorter than you can live with. Then in the `Go-Logistics` repo → Settings → Secrets and variables → Actions create:

| Kind | Name | Value |
|---|---|---|
| Secret | `DEPLOY_HOST` | `203.0.113.10` |
| Secret | `DEPLOY_USER` | `deploy` |
| Secret | `DEPLOY_SSH_KEY` | contents of `fleet_deploy_key` (the private file) |
| Secret | `GHCR_PULL_TOKEN` | the `read:packages` PAT |
| Variable | `API_DOMAIN` | `api.example.com` |

Verify: all five listed in the repo's Actions settings.

---

### Task 6: First deployment + admin bootstrap

Prerequisite: the user pushes `main` (this is their explicit call — 22+ local commits plus Tasks 1–4).

- [ ] **Step 1: Push and watch.** `git push origin main`, then watch Actions → Deploy. Expected: `test` → `build-push` → `deploy` all green, health gate reports `healthy`. First TLS issuance can add ~10s.

- [ ] **Step 2: Verify HTTPS + migrations.** From the laptop: `curl -s https://api.example.com/healthz` → `{"status":"ok"}`. On the server: `cd /opt/fleet && docker compose logs migrate` ends with the 9 migrations applied; `docker compose ps` shows caddy/db/minio/api `running`, migrate `exited (0)`.

- [ ] **Step 3: Seed the first company + admin employee.** The API's `POST /companies` needs an authenticated account-owner, so the very first identity is seeded by SQL, then given a password by the CLI (exact minimal insert; the schema's NOT NULL text columns take `''`):

```sh
cd /opt/fleet
docker compose exec db psql -U postgres -d fleet -c "
INSERT INTO company (name,tax_id,address,created_at,phone,email,website,city,region,postal_code,country,timezone,currency,system_of_measurement)
VALUES ('<Company Name>','','',now(),'','','','','','','MX','America/Mexico_City','MXN','metric');
INSERT INTO employee (default_company_id,first_name,last_name,employee_id,is_active,email,mobile_phone,work_phone,job_title,is_technician,is_vehicle_operator,is_account_owner,license_class,license_number,license_state,street_address,city,region,postal_code,country,updated_at,password_hash)
VALUES (1,'<First>','<Last>','E1',true,'<admin-email>','','','Administrator',false,false,true,'','','','','','','','',now(),'');
"
# bootstrap: creates the Administrador/Almacén roles + default statuses,
# links the employee as admin — all transactionally (internal/bootstrap).
docker compose run --rm cli bootstrap --email '<admin-email>' --company-id 1
docker compose run --rm cli setpass --email '<admin-email>' --password '<strong password ≥8>'
```
Verify: `curl -s -X POST https://api.example.com/auth/login -H 'Content-Type: application/json' -d '{"email":"<admin-email>","password":"<password>"}'` returns a token pair.

- [ ] **Step 4: End-to-end smoke.** With `TOKEN` from Step 3:

```sh
curl -s https://api.example.com/api/v1/me/permissions -H "Authorization: Bearer $TOKEN" | head -c 200
curl -s -X POST https://api.example.com/api/v1/uploads -H "Authorization: Bearer $TOKEN" -F file=@photo.png
curl -fsS <the returned url>   # public object served via https://media.example.com/fleet/...
curl -s -o /dev/null -w '%{http_code}\n' https://api.example.com/swagger/index.html   # 404: prod hides docs
```
Expected: permissions JSON; a 201 whose `url` starts `https://media.example.com/fleet/` and serves over HTTPS; `404` for Swagger.

- [ ] **Step 5: Rollback drill (while the stakes are zero).**

```sh
ssh deploy@203.0.113.10 '/opt/fleet/deploy.sh ghcr.io/jesuslab135/go-logistics:<previous-sha>'
curl -fsS https://api.example.com/healthz
ssh deploy@203.0.113.10 '/opt/fleet/deploy.sh ghcr.io/jesuslab135/go-logistics:<current-sha>'
```
Expected: healthy on both, proving rollback is a one-liner.

---

### Task 7: Backups, runbook, README

**Files:**
- Create: `deploy/README.md`
- Modify: `README.md` (deployment pointer)

- [ ] **Step 1: Schedule the backup.** On the server as `deploy`: `crontab -e` →

```cron
15 3 * * * /opt/fleet/backup.sh >> /opt/fleet/backups/backup.log 2>&1
```

Then run it once by hand: `/opt/fleet/backup.sh` → expected `backup <stamp> done`, and `ls /opt/fleet/backups` shows a `.dump` and a `.tar.gz`.

- [ ] **Step 2: Restore drill (prove the dump is a backup, not a hope).**

```sh
cd /opt/fleet
docker compose exec -T db pg_restore -U postgres --clean --if-exists -d fleet --schema-only --dry-run < backups/fleet-<stamp>.dump 2>/dev/null || true
docker compose exec -T db createdb -U postgres fleet_restore_test
docker compose exec -T db pg_restore -U postgres -d fleet_restore_test < backups/fleet-<stamp>.dump
docker compose exec db psql -U postgres -d fleet_restore_test -c 'SELECT count(*) FROM employee;'
docker compose exec db dropdb -U postgres fleet_restore_test
```
Expected: the count matches production's employee count.

- [ ] **Step 3: Write `deploy/README.md`** — the runbook, containing exactly: the secrets/variables table from Task 5 Step 7; the server layout (`/opt/fleet`: `docker-compose.yml`, `Caddyfile`, `.env` (server-owned, chmod 600, never deployed), `migrations/`, `deploy.sh`, `backup.sh`, `backups/`); how a deploy works (push to `main` → Actions → `deploy.sh <sha-tag>` → health gate); rollback (`/opt/fleet/deploy.sh ghcr.io/jesuslab135/go-logistics:<old-sha>`); first-boot bootstrap (Task 6 Step 3 commands verbatim); logs (`cd /opt/fleet && docker compose logs -f api`); MinIO console (`ssh -L 9001:127.0.0.1:9001 deploy@<vps>` → `http://localhost:9001`); restore (Step 2 commands, plus MinIO: stop stack, untar `minio-<stamp>.tar.gz` into the `fleet_minio_data` volume via the same busybox pattern, start stack); and the two off-server risks: backups live on the same VPS (copy `backups/` elsewhere — even a laptop cron pulling via `scp` — before trusting them) and the `GHCR_PULL_TOKEN`/certificates renew automatically but the PAT does not.

- [ ] **Step 4: README pointer.** In `README.md`, after the "Development" section, add:

```markdown
## Deployment

Production runs on an IONOS VPS behind Caddy (automatic HTTPS): push to `main`
and GitHub Actions tests, builds `ghcr.io/jesuslab135/go-logistics:<sha>`, and
rolls the stack over SSH. Operations — first boot, secrets, rollback, backups,
restore — are documented in [`deploy/README.md`](deploy/README.md).
```

- [ ] **Step 5: Gate + commit.**

```bash
go build ./... && go vet ./... && test -z "$(gofmt -l .)" && go test ./...
git add deploy/README.md README.md
git commit -m "docs: deployment runbook and README pointer"
git push origin main   # rides the pipeline it documents
```
Expected: the push itself triggers Deploy; green run = the pipeline redeployed unchanged code, which is the final regression test.

---

## Self-Review (done at plan time)

- **Coverage vs the decisions:** provisioned VPS → Task 5 starts at hardening, no provisioning steps. Backend only → no frontend tasks; `CORS_ORIGINS` carries the future frontend origin. Domain/TLS → Caddy + ACME (Tasks 2, 5.1, 6.2). Production-only from `main` → single stack, `deploy.yml` on `main` + dispatch.
- **Gaps closed during writing:** the CLI wasn't in the prod image (Task 1); migrations aren't in the image so the bundle ships them (Task 4 scp); MinIO console reachable only via loopback + SSH tunnel; IONOS panel firewall called out separately from ufw; first-identity seeding documented because `POST /companies` requires an authenticated owner.
- **Type/name consistency:** image name `ghcr.io/jesuslab135/go-logistics` everywhere; bundle file names in Task 4 match Task 2's outputs (`docker-compose.prod.yml` → server `docker-compose.yml` noted in both); secrets names in Task 4's workflow match Task 5's table; `deploy.sh <image-ref>` signature consistent across Tasks 2, 4, 6.5, and the runbook.
- **Known deviation from strict TDD:** infrastructure tasks verify with commands (`docker compose config`, `curl`, restore drills) rather than Go tests — each step still ends with an expected observable result.
