#!/usr/bin/env bash
# Roll the stack to a specific image. The pipeline calls this with an
# immutable SHA tag; recording it in .env means a plain `docker compose up -d`
# after a reboot re-runs the same release, and rollback is re-running this
# script with the previous SHA.
set -euo pipefail

IMAGE="${1:?usage: deploy.sh <image-ref>}"
cd "$(dirname "$0")"

# Get the image BEFORE recording it: a failed pull must not leave .env pointing
# at something the host does not have, or the next reboot comes up broken.
#
# The registry credential the pipeline installs is scoped to its own run and
# expires with it, so a rollback done by hand hours later WILL be denied. That
# is fine as long as the image is already on disk — which for a rollback it
# normally is — so a denied pull is only fatal when there is no local copy.
if ! docker pull --quiet "${IMAGE}" >/dev/null 2>&1; then
    if docker image inspect "${IMAGE}" >/dev/null 2>&1; then
        echo "registry unavailable; using the local copy of ${IMAGE}"
    else
        echo "deploy.sh: cannot pull ${IMAGE} and there is no local copy." >&2
        echo "  Authenticate, then retry:" >&2
        echo "    docker login ghcr.io -u <github-user>   # PAT with read:packages" >&2
        exit 1
    fi
fi

if grep -q '^API_IMAGE=' .env; then
    sed -i "s|^API_IMAGE=.*|API_IMAGE=${IMAGE}|" .env
else
    echo "API_IMAGE=${IMAGE}" >> .env
fi

docker compose up -d --remove-orphans

# Reclaim disk WITHOUT a blanket `docker image prune`: this host also runs the
# unrelated `sga` project, and a global prune would eat its images too. Keep the
# three most recent tags of our own repository and drop the rest; docker refuses
# to remove an image still in use, so the running release is safe either way.
#
# Production (<sha>) and QA (qa-<sha>) tags share the repository, so each stack
# only counts and prunes its own kind — a burst of QA deploys must never evict
# production's rollback images.
case "${IMAGE##*:}" in
    qa-*) own_tags() { grep ':qa-'; } ;;
    *)    own_tags() { grep -v ':qa'; } ;;
esac
docker images --filter "reference=${IMAGE%:*}" --format '{{.Repository}}:{{.Tag}}' \
    | grep -v ':latest$' \
    | own_tags \
    | tail -n +4 \
    | xargs -r docker rmi >/dev/null 2>&1 || true

echo "deployed ${IMAGE}"
