#!/bin/bash
set -euo pipefail
WORKSPACE_ROOT="$1"
IMAGE="ghcr.io/xonovex/agent-operator-go"
DOCKERFILE="packages/agent/agent-operator-go/Dockerfile"
BUILDER="xonovex-builder"
SHORT_SHA=$(git rev-parse --short HEAD 2> /dev/null || echo "dev")
BUILD_TIMESTAMP=$(date +%s)
TAG="${SHORT_SHA}-${BUILD_TIMESTAMP}"

# The named builder stays isolated because each operation selects it explicitly.
docker buildx inspect "$BUILDER" > /dev/null 2>&1 \
  || docker buildx create --name "$BUILDER"

# Login to GHCR
echo "${GITHUB_TOKEN}" | docker login ghcr.io -u "${GITHUB_ACTOR:-deorder}" --password-stdin

# Build multi-arch image with registry layer caching
docker buildx build \
  --builder "$BUILDER" \
  --platform linux/amd64,linux/arm64 \
  -f "$DOCKERFILE" \
  --cache-from "type=registry,ref=${IMAGE}:cache" \
  --cache-to "type=registry,ref=${IMAGE}:cache,mode=max" \
  -t "${IMAGE}:${TAG}" \
  -t "${IMAGE}:latest" \
  --push \
  "$WORKSPACE_ROOT"

# Cap the persistent BuildKit cache: the docker-container builder keeps its own
# cache volume that is never auto-pruned and balloons unbounded otherwise.
docker buildx prune --builder "$BUILDER" --keep-storage "${BUILDX_CACHE_KEEP:-20GB}" --force > /dev/null 2>&1 || true
