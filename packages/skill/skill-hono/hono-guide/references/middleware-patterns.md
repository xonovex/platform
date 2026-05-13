# middleware-patterns: CORS Configuration and Custom Middleware

**Guideline:** Configure CORS differently for development and production, and use middleware factory functions to ensure proper `this` binding and parameterization.

**Rationale:** Environment-specific CORS balances development convenience with production security. Middleware factories ensure correct context binding in async operations, enable parameterization, and provide consistent patterns across applications through proper closure over configuration values.

**Example:**

```typescript
import type {Context, Next} from "hono";
import {cors} from "hono/cors";

// Environment-specific CORS
const isDevelopment = process.env.NODE_ENV === "development";

if (isDevelopment) {
  // Development: Permissive
  app.use(
    "*",
    cors({
      origin: "*",
      allowMethods: ["GET", "POST", "PUT", "DELETE"],
      maxAge: 86_400, // 24 hours
    }),
  );
} else {
  // Production: Restrictive
  app.use(
    "*",
    cors({
      origin: ["https://app.example.com"],
      allowMethods: ["GET", "POST", "PUT", "DELETE"],
      credentials: true,
      maxAge: 600, // 10 minutes
    }),
  );
}

// Custom middleware factory with proper binding
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

**Techniques:**

- Import CORS from `hono/cors` and configure based on environment
- For development use `origin: "*"` for permissive access
- For production specify allowed origins explicitly
- Create middleware factory function that returns `async (c, next) => {...}`
- Call factory function when registering: `app.use("*", requestId())`
- Never destructure methods from helpers that rely on `this` binding
