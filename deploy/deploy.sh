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
