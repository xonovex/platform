---
name: hono-opinionated-guidelines
description: >-
  Trigger on `*.ts` files with Hono imports or `@hono/` packages. Opinionated patterns: inline OpenAPI handlers, router selection, remove unnecessary async, bodyLimit. Keywords: Hono, OpenAPIHono, LinearRouter/RegExpRouter, inline handlers, bodyLimit.
---

# Hono Opinionated Guidelines

## Requirements

- Hono ≥ 4.0, @hono/node-server, @hono/zod-openapi, TypeScript ≥ 5.8

## Opinionated patterns

- **Async controllers** - Remove unnecessary `async` from synchronous handlers, see [reference/controllers.md](reference/controllers.md)
- **OpenAPI type inference** - Use inline handlers for type safety, explicit status codes, OpenAPIHono hierarchy, see [reference/openapi-inline-handlers.md](reference/openapi-inline-handlers.md), [reference/openapi-explicit-status-codes.md](reference/openapi-explicit-status-codes.md), [reference/openapi-router-hierarchy.md](reference/openapi-router-hierarchy.md)
- **OpenAPI documentation** - Use `app.doc()` for automatic spec generation, see [reference/openapi-spec-generation.md](reference/openapi-spec-generation.md)
- **Router selection** - LinearRouter for serverless, RegExpRouter for high-throughput, see [reference/router-selection.md](reference/router-selection.md)
- **Request limits** - Use `bodyLimit` middleware to prevent DoS, see [reference/body-limit.md](reference/body-limit.md)

## Example

```typescript
import {OpenAPIHono} from "@hono/zod-openapi";
import {bodyLimit} from "hono/body-limit";
import {secureHeaders} from "hono/secure-headers";

export function createApp() {
  const app = new OpenAPIHono();
  app.use("*", secureHeaders());
  app.use("*", bodyLimit({maxSize: 100 * 1024}));
  app.route("/api/v1", v1Router);
  app.doc("/openapi.json", {
    openapi: "3.1.0",
    info: {title: "API", version: "1.0.0"},
  });
  return app;
}
```

## Progressive disclosure

- Read [reference/controllers.md](reference/controllers.md) - When seeing unnecessary async functions
- Read [reference/openapi-inline-handlers.md](reference/openapi-inline-handlers.md) - When OpenAPI loses type inference
- Read [reference/openapi-explicit-status-codes.md](reference/openapi-explicit-status-codes.md) - When defining OpenAPI response schemas
- Read [reference/openapi-router-hierarchy.md](reference/openapi-router-hierarchy.md) - When composing multiple routers
- Read [reference/openapi-spec-generation.md](reference/openapi-spec-generation.md) - When generating OpenAPI documentation
- Read [reference/router-selection.md](reference/router-selection.md) - When optimizing for serverless/edge or high-throughput
- Read [reference/body-limit.md](reference/body-limit.md) - When preventing oversized request payloads
