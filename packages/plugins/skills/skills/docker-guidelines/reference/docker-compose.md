# docker-compose: Docker Compose Configuration

**Guideline:** Use Docker Compose for development with proper secrets, volumes, and health checks.

**Rationale:** Simplifies multi-container orchestration during development; secrets, volumes, and health checks mirror production while maintaining security.

**Example:**

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

**Techniques:**

- Build configuration: Specify build context and target stage in compose
- Secrets management: Use secrets section for sensitive data, never environment vars
- Volume mounting: Mount volumes for persistent data matching production layout
- Health checks: Configure health checks matching Dockerfile definitions
- Restart policy: Set `unless-stopped` for development persistence
- Port mapping: Expose ports for local development access
