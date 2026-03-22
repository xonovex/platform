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

# Build multi-arch image with registry layer caching
docker buildx build \
  --builder xonovex-builder \
  --platform linux/amd64,linux/arm64 \
  -f "$DOCKERFILE" \
  --cache-from "type=registry,ref=${IMAGE}:cache" \
  --cache-to "type=registry,ref=${IMAGE}:cache,mode=max" \
  -t "${IMAGE}:${TAG}" \
  -t "${IMAGE}:latest" \
  --push \
  "$WORKSPACE_ROOT"
