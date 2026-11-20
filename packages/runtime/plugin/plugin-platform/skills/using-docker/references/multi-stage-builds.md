# multi-stage-builds: Multi-Stage Build Pattern

Use a distroless runtime base (`gcr.io/distroless/nodejs22`, `USER nonroot:nonroot`) and a separate `deps` stage that installs only production dependencies before the `build` stage.

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
