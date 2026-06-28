#!/bin/bash
set -e
WORKSPACE_ROOT="$1"
IMAGE="ghcr.io/xonovex/agent-operator-go"
DOCKERFILE="packages/agent/agent-operator-go/Dockerfile"

docker buildx inspect xonovex-builder >/dev/null 2>&1 || \
  docker buildx create --name xonovex-builder --use

# Read-only registry layer cache, shared with docker-publish.sh. A miss or missing
# GHCR access is non-fatal; buildx falls back to a full build.
if [ -n "${GITHUB_TOKEN:-}" ]; then
  echo "${GITHUB_TOKEN}" | docker login ghcr.io -u "${GITHUB_ACTOR:-deorder}" --password-stdin
fi

docker buildx build \
  --builder xonovex-builder \
  --platform linux/amd64,linux/arm64 \
  -f "$DOCKERFILE" \
  --cache-from "type=registry,ref=${IMAGE}:cache" \
  -t "$IMAGE:latest" \
  "$WORKSPACE_ROOT"

# Cap the persistent BuildKit cache: the docker-container builder keeps its own
# cache volume that is never auto-pruned and balloons unbounded otherwise.
docker buildx prune --builder xonovex-builder --keep-storage "${BUILDX_CACHE_KEEP:-20GB}" --force >/dev/null 2>&1 || true
