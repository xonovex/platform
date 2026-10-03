# docker-compose: Docker Compose Configuration

Point `build.target` at a Dockerfile stage, pass sensitive data through the `secrets:` section (as `*_FILE` env vars, not plain `environment`), mount named `volumes:` for persistence, and mirror the Dockerfile `healthcheck`. Use `restart: unless-stopped` for dev persistence.

```yaml
services:
  app:
    build: {context: ., target: runtime}
    environment:
      - NODE_ENV=production
      - DATABASE_URL_FILE=/run/secrets/db_url
    secrets: [db_url]
    volumes: [app-data:/app/data]
    ports: ["3000:3000"]
    healthcheck:
      test: ["CMD", "node", "healthcheck.js"]
      interval: 30s
      timeout: 3s
      retries: 3
    restart: unless-stopped

secrets:
  db_url:
    file: ./secrets/db_url.txt

volumes:
  app-data:
```
