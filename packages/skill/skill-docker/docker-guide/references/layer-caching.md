# layer-caching: Layer Caching Strategy

Order layers static → dependencies → source → build so a source edit only invalidates the build layer. Copy lockfiles (`package-lock.json`, `poetry.lock`) and install deps before copying source. Verify cache hits with `--progress=plain` under BuildKit.

```dockerfile
FROM node:22-alpine
WORKDIR /app
COPY package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build
```
