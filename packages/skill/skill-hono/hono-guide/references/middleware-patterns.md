# middleware-patterns: CORS Configuration and Custom Middleware

Configure `cors()` (from `hono/cors`) per environment: `origin: "*"` in dev; an explicit origin list plus `credentials: true` in prod. Write custom middleware as a factory returning `async (c, next) => {...}`, and register it by calling the factory (`app.use("*", requestId())`) so config is closed over and `this`-dependent helpers keep their binding.

```typescript
import type {Context, Next} from "hono";
import {cors} from "hono/cors";

app.use(
  "*",
  process.env.NODE_ENV === "development"
    ? cors({origin: "*", maxAge: 86_400})
    : cors({
        origin: ["https://app.example.com"],
        credentials: true,
        maxAge: 600,
      }),
);

function requestId() {
  return async (c: Context, next: Next) => {
    const id = crypto.randomUUID();
    c.set("requestId", id);
    c.header("X-Request-ID", id);
    await next();
  };
}
app.use("*", requestId());
```
