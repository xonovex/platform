# Sources

## Docker documentation

- **URLs:**
  - https://docs.docker.com/build/building/best-practices/
  - https://docs.docker.com/build/building/multi-stage/
  - https://docs.docker.com/build/building/secrets/
- **Last reviewed:** 2026-07-19
- **Used for:** `SKILL.md` and all `references/`; layer caching, multi-stage images, non-root runtime users, BuildKit secrets, health checks, Compose, and `.dockerignore`.
- **Aspects extracted:** Official build-cache ordering, stage-copy behavior, secret mounts, image minimization, and runtime hardening guidance.
