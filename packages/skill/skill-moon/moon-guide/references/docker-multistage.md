# docker-multistage: Moon Docker Scaffold Pattern

Use `moon docker scaffold <project>` in a skeleton stage, then `moon docker prune` after the build. Scaffold output lands in `.moon/docker/workspace` (config/manifests) and `.moon/docker/sources` (source files) — copy `workspace` before `npm ci`, `sources` after, so dependency layers cache independently of source changes.

```dockerfile
FROM base AS skeleton
COPY . .
RUN moon docker scaffold my-service

FROM base AS build
COPY --from=skeleton /app/.moon/docker/workspace .
RUN npm ci
COPY --from=skeleton /app/.moon/docker/sources .
RUN moon run my-service:build && moon docker prune

FROM gcr.io/distroless/nodejs22-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build --chown=65532:65532 /app/node_modules ./node_modules
COPY --from=build --chown=65532:65532 /app/packages ./packages
```

The runtime stage runs distroless nonroot (uid/gid `65532`); copy build artifacts with `--chown=65532:65532`.
