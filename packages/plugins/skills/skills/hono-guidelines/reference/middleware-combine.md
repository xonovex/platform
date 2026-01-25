# middleware-combine: Composing Middleware with some, every, except

**Guideline:** Use the `combine` module (`some`, `every`, `except`) to compose complex middleware logic for conditional execution and access control.

**Rationale:** The combine module provides declarative composition of complex requirements: `some()` implements OR logic for alternative auth methods, `every()` implements AND logic for layered checks, and `except()` skips middleware for specific paths, enabling cleaner code than nested conditionals.

**Example:**

```typescript
import {basicAuth} from "hono/basic-auth";
import {bearerAuth} from "hono/bearer-auth";
import {every, except, some} from "hono/combine";

// Accept either basic auth OR bearer token
app.use(
  "/api/*",
  some(
    basicAuth({verifyUser: verifyBasicAuth}),
    bearerAuth({verifyToken: verifyBearerToken}),
  ),
);

// Must be authenticated AND have admin role
app.use(
  "/admin/*",
  every(bearerAuth({verifyToken: verifyToken}), async (c, next) => {
    const user = c.get("user");
    if (user.role !== "admin") {
      return c.json({error: "Forbidden"}, 403);
    }
    await next();
  }),
);

// Apply rate limiting to all routes EXCEPT health checks
app.use("*", except("/health", rateLimiter({max: 100, window: 60})));

// Multiple path exclusions
app.use("*", except(["/health", "/metrics", "/ready"], authMiddleware()));
```

**Techniques:**
- Import `some`, `every`, `except` from `hono/combine`
- Use `some()` for alternative auth methods (OAuth OR API key)
- Use `every()` for layered requirements (authenticated AND has role)
- Use `except()` for path-based middleware exclusions
- Combine module handlers with other Hono middleware seamlessly
