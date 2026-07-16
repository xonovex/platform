#!/bin/bash
set -euo pipefail
WORKSPACE_ROOT="$1"
IMAGE="ghcr.io/xonovex/agent-operator-go"
DOCKERFILE="packages/agent/agent-operator-go/Dockerfile"
BUILDER="xonovex-builder"

# The named builder stays isolated because each operation selects it explicitly.
docker buildx inspect "$BUILDER" > /dev/null 2>&1 \
  || docker buildx create --name "$BUILDER"

# Read-only registry layer cache, shared with docker-publish.sh. A miss or missing
# GHCR access is non-fatal; buildx falls back to a full build.
if [ -n "${GITHUB_TOKEN:-}" ]; then
  echo "${GITHUB_TOKEN}" | docker login ghcr.io -u "${GITHUB_ACTOR:-deorder}" --password-stdin
fi

docker buildx build \
  --builder "$BUILDER" \
  --platform linux/amd64,linux/arm64 \
  -f "$DOCKERFILE" \
  --cache-from "type=registry,ref=${IMAGE}:cache" \
  -t "$IMAGE:latest" \
  "$WORKSPACE_ROOT"

# Cap the persistent BuildKit cache: the docker-container builder keeps its own
# cache volume that is never auto-pruned and balloons unbounded otherwise.
docker buildx prune --builder "$BUILDER" --keep-storage "${BUILDX_CACHE_KEEP:-20GB}" --force > /dev/null 2>&1 || true
