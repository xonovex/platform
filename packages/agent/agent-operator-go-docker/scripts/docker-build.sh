#!/bin/bash
set -e
WORKSPACE_ROOT="$1"
IMAGE="ghcr.io/xonovex/agent-operator-go"
DOCKERFILE="packages/agent/agent-operator-go/Dockerfile"

docker buildx inspect xonovex-builder >/dev/null 2>&1 || \
  docker buildx create --name xonovex-builder --use

docker buildx build \
  --builder xonovex-builder \
  --platform linux/amd64,linux/arm64 \
  -f "$DOCKERFILE" \
  -t "$IMAGE:latest" \
  "$WORKSPACE_ROOT"
