# docker-multistage: Moon Docker Scaffold Pattern

Use `moon docker scaffold <project>` in a skeleton stage to extract only that project's workspace config + dependencies, then `moon docker prune` after build to strip dev deps and artifacts. Scaffold output lands in `.moon/docker/workspace` (config/manifests) and `.moon/docker/sources` (source files) — copy `workspace` before `npm ci`, `sources` after, so dependency layers cache independently of source changes.

```dockerfile
FROM base AS skeleton
COPY . .
RUN moon docker scaffold my-service

FROM base AS build
COPY --from=skeleton /app/.moon/docker/workspace .
RUN npm ci
COPY --from=skeleton /app/.moon/docker/sources .
RUN moon run my-service:build && moon docker prune
```
