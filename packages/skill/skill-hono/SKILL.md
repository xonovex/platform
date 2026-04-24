---
name: hono-guidelines
description: Trigger on `*.ts` files with Hono imports or `@hono/` packages. Use when building Hono 4.0+ API servers. Apply for validation, middleware, error handling, WebSockets. Keywords: Hono, middleware, Zod validator, error handling, WebSocket.
---

# Hono Coding Guidelines

## Requirements

- Hono ≥ 4.0, @hono/node-server, @hono/zod-validator, TypeScript ≥ 5.8

## Essentials

- **Factory functions** - Use for testability and domain organization, see [reference/application-structure.md](reference/application-structure.md)
- **Type-safe validation** - Cast `c.req.valid` properly, handle errors with Zod, see [reference/validation-type-safety.md](reference/validation-type-safety.md)
- **Middleware configuration** - Use factories for CORS, composition with `some()`/`every()`/`except()`, see [reference/middleware-patterns.md](reference/middleware-patterns.md), [reference/middleware-combine.md](reference/middleware-combine.md)
- **WebSocket helpers** - Keep object references to maintain `this` binding, see [reference/websocket-support.md](reference/websocket-support.md)
- **Error responses** - Use RFC 7807 Problem Details format, see [reference/error-handling.md](reference/error-handling.md)
- **Security middleware** - Apply `secureHeaders()` for security headers, see [reference/security-middleware.md](reference/security-middleware.md)
- **Cookie handling** - Set secure options explicitly, use signed cookies, see [reference/cookie-handling.md](reference/cookie-handling.md)
- **Platform portability** - Use `env(c)` for environment, `getRuntimeKey()` for detection, see [reference/platform-runtime.md](reference/platform-runtime.md)

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

## Progressive disclosure

- Read [reference/application-structure.md](reference/application-structure.md) - When organizing a new Hono application
- Read [reference/validation-type-safety.md](reference/validation-type-safety.md) - When losing type safety after validation
- Read [reference/middleware-patterns.md](reference/middleware-patterns.md) - When creating reusable middleware or configuring CORS
- Read [reference/websocket-support.md](reference/websocket-support.md) - When implementing WebSocket endpoints
- Read [reference/error-handling.md](reference/error-handling.md) - When standardizing error responses across routes
- Read [reference/security-middleware.md](reference/security-middleware.md) - When configuring auth, CSRF, or security headers
- Read [reference/cookie-handling.md](reference/cookie-handling.md) - When managing sessions or sensitive cookies
- Read [reference/context-storage.md](reference/context-storage.md) - When accessing Context outside handlers
- Read [reference/middleware-combine.md](reference/middleware-combine.md) - When composing complex middleware logic
- Read [reference/platform-runtime.md](reference/platform-runtime.md) - When deploying across multiple runtimes
