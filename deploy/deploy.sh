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

# Reclaim disk WITHOUT a blanket `docker image prune`: this host also runs the
# unrelated `sga` project, and a global prune would eat its images too. Keep the
# three most recent tags of our own repository and drop the rest; docker refuses
# to remove an image still in use, so the running release is safe either way.
docker images --filter "reference=${IMAGE%:*}" --format '{{.Repository}}:{{.Tag}}' \
    | grep -v ':latest$' \
    | tail -n +4 \
    | xargs -r docker rmi >/dev/null 2>&1 || true

echo "deployed ${IMAGE}"
