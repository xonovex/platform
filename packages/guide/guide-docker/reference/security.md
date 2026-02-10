# security: Security Best Practices

**Guideline:** Pin versions, run non-root, externalize secrets to limit attack surface.

**Rationale:** Multiple defense layers: pinned versions prevent supply chain attacks, non-root limits breach damage, externalized secrets prevent leaks.

**Example:**

```dockerfile
FROM node:22.1.0-alpine3.19

RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001
USER nodejs:nodejs

ENV DATABASE_URL_FILE=/run/secrets/db_url

RUN mkdir /tmp/app-cache && chown nodejs:nodejs /tmp/app-cache
VOLUME /tmp/app-cache
```

**Techniques:**

- Version pinning: Pin base image version including OS version for supply chain security
- Non-root user: Create and use non-root USER to limit breach damage
- Secret externalization: Never copy secrets into image; use environment or mounts
- Read-only filesystem: Use read-only root with writable VOLUME for security
- Defense layers: Combine multiple tactics (pinning, user, secrets) for defense-in-depth
