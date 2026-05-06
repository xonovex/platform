---
name: docker-guide
description: "Use when writing or editing Docker images and Compose files for production. Triggers on `Dockerfile`, `docker-compose.yml`, `.dockerignore`, and prompts about multi-stage builds, alpine/distroless base images, layer caching, non-root users, BuildKit, secrets, health checks, or image slimming, even when the user doesn't say 'Docker'. Skip Kubernetes manifests (use kubernetes-guide) and Podman/containerd-only runtime config."
---

# Docker Coding Guidelines

## Essentials

- **Build optimization** - Multi-stage builds, small base images (alpine/distroless), see [reference/multi-stage-builds.md](reference/multi-stage-builds.md)
- **Layer caching** - Order layers for caching, copy lockfiles before source, see [reference/layer-caching.md](reference/layer-caching.md)
- **Security** - Run as non-root, least privilege, no secrets in images, see [reference/security.md](reference/security.md)
- **Configuration** - Externalize config via env/volumes, pin versions, see [reference/production-config.md](reference/production-config.md)
- **Quality** - Use BuildKit and hadolint, add health endpoints, see [reference/production-config.md](reference/production-config.md)

## Progressive disclosure

- Read [reference/multi-stage-builds.md](reference/multi-stage-builds.md) - When creating production images or optimizing build process
- Read [reference/layer-caching.md](reference/layer-caching.md) - When builds are slow or cache invalidates frequently
- Read [reference/security.md](reference/security.md) - When hardening images or handling secrets
- Read [reference/production-config.md](reference/production-config.md) - When adding health checks or configuring for orchestration
- Read [reference/docker-compose.md](reference/docker-compose.md) - When defining multi-service local development environments
