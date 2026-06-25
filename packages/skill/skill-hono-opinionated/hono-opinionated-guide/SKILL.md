---
name: hono-opinionated-guide
description: "Use when editing Hono APIs that follow the opinionated style — inline OpenAPI handlers, explicit router selection (LinearRouter / RegExpRouter), sync handlers where possible, bodyLimit middleware. An overlay on hono-guide: covers only the opinionated decisions, not generic Hono usage. Triggers on `.ts` files with `OpenAPIHono`, inline route schemas, and prompts about router perf, payload limits, or sync-vs-async handlers, even when the user doesn't say 'opinionated'."
---

# Hono Opinionated Guidelines

An overlay on **hono-guide**. Apply **hono-guide** for all generic Hono work — application structure, validation/type safety, middleware patterns and combination, error handling, cookies, security, WebSockets, context storage, platform runtimes. This skill adds only the opinionated decisions on top.

## Requirements

- Hono ≥ 4.0, @hono/node-server, @hono/zod-openapi, TypeScript ≥ 5.8

## Opinionated patterns

- **Foundation** - Everything generic lives in **hono-guide**; this skill layers OpenAPI-first conventions on top

- **Async controllers** - Remove unnecessary `async` from synchronous handlers, see [references/controllers.md](references/controllers.md)
- **OpenAPI type inference** - Use inline handlers for type safety, explicit status codes, OpenAPIHono hierarchy, see [references/openapi-inline-handlers.md](references/openapi-inline-handlers.md), [references/openapi-explicit-status-codes.md](references/openapi-explicit-status-codes.md), [references/openapi-router-hierarchy.md](references/openapi-router-hierarchy.md)
- **OpenAPI documentation** - Use `app.doc()` for automatic spec generation, see [references/openapi-spec-generation.md](references/openapi-spec-generation.md)
- **Router selection** - LinearRouter for serverless, RegExpRouter for high-throughput, see [references/router-selection.md](references/router-selection.md)
- **Request limits** - Use `bodyLimit` middleware to prevent DoS, see [references/body-limit.md](references/body-limit.md)

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

## Gotchas

- Inline OpenAPI handlers tie validation, status codes, and types together — splitting them later is painful, write them inline from day one
- Explicit status codes (`c.json(body, 201)`) are required by the OpenAPI generator — implicit 200 produces wrong specs
- Router hierarchy: parent routes inherit middleware; mounting a sub-router with `.route()` runs the parent middleware first
- Picking the router (RegExpRouter / SmartRouter / TrieRouter / PatternRouter) is a startup decision — switching requires testing all routes

## Progressive disclosure

- Read [references/controllers.md](references/controllers.md) - Load when seeing unnecessary async functions
- Read [references/openapi-inline-handlers.md](references/openapi-inline-handlers.md) - Load when OpenAPI loses type inference
- Read [references/openapi-explicit-status-codes.md](references/openapi-explicit-status-codes.md) - Load when defining OpenAPI response schemas
- Read [references/openapi-router-hierarchy.md](references/openapi-router-hierarchy.md) - Load when composing multiple routers
- Read [references/openapi-spec-generation.md](references/openapi-spec-generation.md) - Load when generating OpenAPI documentation
- Read [references/router-selection.md](references/router-selection.md) - Load when optimizing for serverless/edge or high-throughput
- Read [references/body-limit.md](references/body-limit.md) - Load when preventing oversized request payloads
