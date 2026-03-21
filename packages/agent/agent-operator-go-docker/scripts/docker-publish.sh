#!/bin/bash
set -e
WORKSPACE_ROOT="$1"
IMAGE="ghcr.io/xonovex/agent-operator-go"
DOCKERFILE="packages/agent/agent-operator-go/Dockerfile"
SHORT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIMESTAMP=$(date +%s)
TAG="${SHORT_SHA}-${BUILD_TIMESTAMP}"

docker buildx inspect xonovex-builder >/dev/null 2>&1 || \
  docker buildx create --name xonovex-builder --use

# Login to GHCR
echo "${GITHUB_TOKEN}" | docker login ghcr.io -u "${GITHUB_ACTOR:-deorder}" --password-stdin

# Build each architecture with registry layer caching
for ARCH in amd64 arm64; do
  docker buildx build \
    --builder xonovex-builder \
    --platform "linux/${ARCH}" \
    --build-arg BUILDKIT_INLINE_CACHE=1 \
    -f "$DOCKERFILE" \
    --cache-from "type=registry,ref=${IMAGE}:cache-${ARCH}" \
    --cache-to "type=registry,ref=${IMAGE}:cache-${ARCH},mode=max" \
    -t "${IMAGE}:${TAG}-${ARCH}" \
    --push \
    "$WORKSPACE_ROOT"
done

# Merge per-arch images into multi-arch manifest
docker buildx imagetools create \
  -t "${IMAGE}:${TAG}" \
  -t "${IMAGE}:latest" \
  "${IMAGE}:${TAG}-amd64" \
  "${IMAGE}:${TAG}-arm64"
