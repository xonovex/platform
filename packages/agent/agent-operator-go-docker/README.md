# Agent Operator Go Docker

Multi-arch Docker image build and GHCR publish for [agent-operator-go](../agent-operator-go/).

Builds `linux/amd64` and `linux/arm64` images as a multi-platform manifest at `ghcr.io/xonovex/agent-operator-go`, with registry layer caching for published builds.

## Usage

```bash
# Local multi-arch build (no push)
npx moon run agent-operator-go-docker:docker-build

# Build and publish to GHCR
npx moon run agent-operator-go-docker:docker-publish
```

## Tags

- `<sha>-<timestamp>` multi-platform manifest
- `latest` multi-platform manifest
