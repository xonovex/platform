# security: Security Best Practices

Externalize secrets via a file-based mechanism, never as a plain `ENV`/`environment` value: `ENV DATABASE_URL_FILE=/run/secrets/db_url` (the app reads the file, not the value). Pin the base image including OS version. Use a read-only root filesystem with a writable `VOLUME` for caches.

```dockerfile
FROM node:22.1.0-alpine3.19

ENV DATABASE_URL_FILE=/run/secrets/db_url

RUN mkdir /tmp/app-cache
VOLUME /tmp/app-cache
```
