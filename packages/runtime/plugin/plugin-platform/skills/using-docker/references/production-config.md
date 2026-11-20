# production-config: Production Configuration

Externalize deployment config via `ENV` and declare a `VOLUME` for data that must survive restarts. The `# syntax=` directive enables BuildKit features.

```dockerfile
# syntax=docker/dockerfile:1.7
FROM node:22-alpine
WORKDIR /app

ENV NODE_ENV=production PORT=3000 LOG_LEVEL=info

VOLUME /app/data

EXPOSE 3000
USER node:node
CMD ["node", "dist/server.js"]
```
