# Agent Operator Go Docker

Multi-arch Docker image build and GHCR publish for [agent-operator-go](../agent-operator-go/).

Builds `linux/amd64` and `linux/arm64` images, pushes per-arch images with registry layer caching, then merges into a multi-arch manifest at `ghcr.io/xonovex/agent-operator-go`.

## Usage

```bash
# Local multi-arch build (no push)
npx moon run agent-operator-go-docker:docker-build

# Build and publish to GHCR
npx moon run agent-operator-go-docker:docker-publish
```

## Tags

- `<sha>-<timestamp>-<arch>` per architecture
- `<sha>-<timestamp>` merged multi-arch manifest
- `latest` merged multi-arch manifest
