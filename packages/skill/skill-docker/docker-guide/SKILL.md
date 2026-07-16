---
name: docker-guide
description: "Use when writing or editing Docker images and Compose files for production. Triggers on `Dockerfile`, `docker-compose.yml`, `.dockerignore`, and prompts about multi-stage builds, alpine/distroless base images, layer caching, non-root users, BuildKit, secrets, health checks, or image slimming, even when the user doesn't say 'Docker'."
---

# Docker Coding Guidelines

## Essentials

- **Build optimization** - Multi-stage builds, small base images (alpine/distroless), see [references/multi-stage-builds.md](references/multi-stage-builds.md)
- **Security** - Least privilege, externalize secrets (never in images), see [references/security.md](references/security.md)
- **Configuration** - Externalize config via env/volumes, pin versions, see [references/production-config.md](references/production-config.md)
- **Quality** - Use BuildKit and hadolint, see [references/production-config.md](references/production-config.md)

## Gotchas

- Multi-stage builds need explicit `COPY --from=<stage>` — forgetting it copies the whole heavy stage into the final image
- `ENTRYPOINT` + `CMD` interact: `CMD` provides default args to `ENTRYPOINT`; overriding `CMD` from `docker run` drops them, not appends

## Progressive disclosure

- Read [references/multi-stage-builds.md](references/multi-stage-builds.md) - Load when creating production images or optimizing build process
- Read [references/security.md](references/security.md) - Load when hardening images or handling secrets
- Read [references/production-config.md](references/production-config.md) - Load when adding health checks or configuring for orchestration
- Read [references/docker-compose.md](references/docker-compose.md) - Load when defining multi-service local development environments
