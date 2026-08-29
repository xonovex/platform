# Agent Operator Go Docker

Use this package to build or publish the [agent-operator-go](../agent-operator-go/) image for `linux/amd64` and `linux/arm64`. Published builds create a multi-platform manifest at `ghcr.io/xonovex/agent-operator-go` and reuse registry layer caching.

## Usage

Choose a local build for verification or the publish task to push both platforms to GitHub Container Registry.

```bash
# Local multi-arch build (no push)
npx moon run agent-operator-go-docker:docker-build

# Build and publish to GHCR
npx moon run agent-operator-go-docker:docker-publish
```

## Tags

Each publication writes the immutable build tag and updates `latest`.

- `<sha>-<timestamp>` multi-platform manifest
- `latest` multi-platform manifest
