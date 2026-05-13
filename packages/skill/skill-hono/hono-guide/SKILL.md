---
name: hono-guide
description: "Use when editing or scaffolding Hono 4.0+ API servers in TypeScript. Triggers on `.ts` files with `hono` or `@hono/*` imports, route handlers, middleware, and prompts about validation, error handling, WebSockets, CORS, or building an HTTP API in Hono — even when the user doesn't say 'Hono'. Skip Express (use express.js-guide), the opinionated inline-OpenAPI style (use hono-opinionated-guide), and non-HTTP TypeScript work."
---

# Hono Coding Guidelines

## Requirements

- Hono ≥ 4.0, @hono/node-server, @hono/zod-validator, TypeScript ≥ 5.8

## Essentials

- **Factory functions** - Use for testability and domain organization, see [references/application-structure.md](references/application-structure.md)
- **Type-safe validation** - Cast `c.req.valid` properly, handle errors with Zod, see [references/validation-type-safety.md](references/validation-type-safety.md)
- **Middleware configuration** - Use factories for CORS, composition with `some()`/`every()`/`except()`, see [references/middleware-patterns.md](references/middleware-patterns.md), [references/middleware-combine.md](references/middleware-combine.md)
- **WebSocket helpers** - Keep object references to maintain `this` binding, see [references/websocket-support.md](references/websocket-support.md)
- **Error responses** - Use RFC 7807 Problem Details format, see [references/error-handling.md](references/error-handling.md)
- **Security middleware** - Apply `secureHeaders()` for security headers, see [references/security-middleware.md](references/security-middleware.md)
- **Cookie handling** - Set secure options explicitly, use signed cookies, see [references/cookie-handling.md](references/cookie-handling.md)
- **Platform portability** - Use `env(c)` for environment, `getRuntimeKey()` for detection, see [references/platform-runtime.md](references/platform-runtime.md)

## Example

```typescript
import {Hono} from "hono";
import {secureHeaders} from "hono/secure-headers";

export function createApp() {
  const app = new Hono();
  app.use("*", secureHeaders());
  app.route("/api/v1", v1Router);
  return app;
}
```

## Gotchas

- `c.req.valid('json')` returns `any` unless you cast or use a typed factory — type safety leaks at the boundary
- Middleware runs in the order it's registered — auth before routes, error handlers last; ordering bugs cause silent 200s on protected routes
- `env(c)` is the portable way to read env vars across runtimes — `process.env` works on Node but not Workers/Deno
- WebSocket helpers capture `this` from the closing object — assigning the handler to a variable loses the binding

## Progressive disclosure

- Read [references/application-structure.md](references/application-structure.md) - Load when organizing a new Hono application
- Read [references/validation-type-safety.md](references/validation-type-safety.md) - Load when losing type safety after validation
- Read [references/middleware-patterns.md](references/middleware-patterns.md) - Load when creating reusable middleware or configuring CORS
- Read [references/websocket-support.md](references/websocket-support.md) - Load when implementing WebSocket endpoints
- Read [references/error-handling.md](references/error-handling.md) - Load when standardizing error responses across routes
- Read [references/security-middleware.md](references/security-middleware.md) - Load when configuring auth, CSRF, or security headers
- Read [references/cookie-handling.md](references/cookie-handling.md) - Load when managing sessions or sensitive cookies
- Read [references/context-storage.md](references/context-storage.md) - Load when accessing Context outside handlers
- Read [references/middleware-combine.md](references/middleware-combine.md) - Load when composing complex middleware logic
- Read [references/platform-runtime.md](references/platform-runtime.md) - Load when deploying across multiple runtimes
