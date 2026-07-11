# security: Security Best Practices

Pin the base image including OS version (`node:22.1.0-alpine3.19`), create and run as a non-root `USER`, and externalize secrets via env/mounts (never `COPY` them in). Use a read-only root filesystem with a writable `VOLUME` for caches.

```dockerfile
FROM node:22.1.0-alpine3.19

RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001
USER nodejs:nodejs

ENV DATABASE_URL_FILE=/run/secrets/db_url

RUN mkdir /tmp/app-cache && chown nodejs:nodejs /tmp/app-cache
VOLUME /tmp/app-cache
```
