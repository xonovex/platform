# multi-stage-builds: Multi-Stage Build Pattern

Separate deps/build/runtime into named stages (`AS <name>`); `COPY --from=<stage>` only the artifacts into a minimal runtime base (distroless or alpine). Keeps build tools out of the final image.

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
