# docker-multistage: Moon Docker Scaffold Pattern

**Guideline:** Use moon's docker scaffold command in multistage Dockerfiles for efficient monorepo builds.

**Rationale:** Moon's scaffold command extracts only the necessary workspace files and dependencies, reducing build context and enabling efficient layer caching.

**Example:**

```dockerfile
# Base stage
FROM node:22-alpine AS base
WORKDIR /app
RUN npm install -g @moonrepo/cli

# Skeleton stage
FROM base AS skeleton
COPY . .
RUN moon docker scaffold my-service

# Build stage
FROM base AS build
COPY --from=skeleton /app/.moon/docker/workspace .
RUN npm ci
COPY --from=skeleton /app/.moon/docker/sources .
RUN moon run my-service:build
RUN moon docker prune

# Runtime stage
FROM gcr.io/distroless/nodejs22-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build --chown=65532:65532 /app/node_modules ./node_modules
COPY --from=build --chown=65532:65532 /app/packages ./packages
WORKDIR /app/packages/my-service
CMD ["dist/src/server.js"]
```

**Techniques:**
- moon docker scaffold: Extracts workspace config and package dependencies
- moon docker prune: Removes dev dependencies and build artifacts
- Skeleton stage: Separates dependency resolution from source copying
- Distroless runtime: Minimal production image with nonroot user
- Layer ordering: Dependencies before sources for cache efficiency
