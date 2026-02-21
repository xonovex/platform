# production-config: Production Configuration

**Guideline:** Configure with health checks, externalized config, proper lifecycle management.

**Rationale:** Health checks enable failure detection, env vars allow deployment-specific config, volumes preserve data.

**Example:**

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

**Techniques:**

- Health checks: Add HEALTHCHECK for orchestration failure detection
- Environment config: Use ENV for deployment-specific configuration
- Persistent volumes: Define VOLUME for data that must survive restarts
- Port exposure: Use EXPOSE to document container ports
- Non-root user: Set USER for security and orchestration requirements
- BuildKit syntax: Use syntax directive for advanced features
