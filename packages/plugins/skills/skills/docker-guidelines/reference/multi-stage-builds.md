# multi-stage-builds: Multi-Stage Build Pattern

**Guideline:** Separate build and runtime stages to minimize production image size and attack surface.

**Rationale:** Multi-stage builds use full-featured build envs while keeping runtime images minimal; reduces size, attack surface, deployment time.

**Example:**

```dockerfile
# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm i --frozen-lockfile

FROM node:22-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules node_modules
COPY . .
RUN pnpm build

FROM gcr.io/distroless/nodejs22
WORKDIR /app
USER nonroot:nonroot
COPY --from=build /app/dist dist
ENV NODE_ENV=production
EXPOSE 3000
CMD ["dist/server.js"]
```

**Techniques:**
- Stage separation: Use distinct dependencies, build, and runtime stages
- Artifact copying: Copy only necessary outputs between stages
- Minimal runtime: Use distroless or alpine for production image
- Named stages: Use AS for clarity in multi-stage builds
- Size optimization: Reduces image size by excluding build tools from runtime
