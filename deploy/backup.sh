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

find "${BACKUP_DIR}" -name 'fleet-*.dump'   -mtime +7 -delete
find "${BACKUP_DIR}" -name 'minio-*.tar.gz' -mtime +7 -delete

echo "backup ${STAMP} done: $(du -sh ${BACKUP_DIR} | cut -f1) total"
