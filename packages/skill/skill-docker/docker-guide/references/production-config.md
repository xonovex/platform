# production-config: Production Configuration

Add a `HEALTHCHECK` for orchestrator failure detection, externalize deployment config via `ENV`, declare a `VOLUME` for data that must survive restarts, and set a non-root `USER`. The `# syntax=` directive enables BuildKit features.

```dockerfile
# syntax=docker/dockerfile:1.7
FROM node:22-alpine
WORKDIR /app

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD node healthcheck.js || exit 1

ENV NODE_ENV=production PORT=3000 LOG_LEVEL=info

VOLUME /app/data

EXPOSE 3000
USER node:node
CMD ["node", "dist/server.js"]
```
