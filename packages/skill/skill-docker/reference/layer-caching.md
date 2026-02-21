# layer-caching: Layer Caching Strategy

**Guideline:** Order Dockerfile commands least-to-most frequently changing to maximize cache reuse.

**Rationale:** Docker caches each layer; changes invalidate subsequent layers. Proper ordering minimizes rebuild time.

**Example:**

```dockerfile
FROM node:22-alpine
WORKDIR /app
COPY package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build
```

**Techniques:**

- Copy lockfiles first: package-lock.json, poetry.lock rarely change
- Install dependencies: Run before copying source to cache dependency layer
- Copy source last: src/, config files change frequently
- Build commands: Run after source copy, invalidates only build layer
- Order principle: Static → dependencies → source → build
- Cache validation: Check Docker BUILDKIT with --progress=plain for cache hits
- Multi-stage: Use separate stages for different cache patterns
