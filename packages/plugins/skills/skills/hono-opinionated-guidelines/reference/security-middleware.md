# security-middleware: Built-in Security Middleware Configuration

**Guideline:** Use Hono's built-in security middleware for authentication, CSRF protection, and security headers, applying `secureHeaders()` first in the middleware chain.

**Rationale:** Hono provides battle-tested security middleware implementations for basic auth, bearer tokens, JWT verification, CSRF protection, and IP restriction with consistent configuration patterns and type safety, eliminating the need for external dependencies.

**Example:**

```typescript
import {Hono} from "hono";
import {basicAuth} from "hono/basic-auth";
import {bearerAuth} from "hono/bearer-auth";
import {csrf} from "hono/csrf";
import {ipRestriction} from "hono/ip-restriction";
import {secureHeaders} from "hono/secure-headers";

const app = new Hono();

// Apply secure headers first
app.use("*", secureHeaders());

// CSRF protection for form submissions
app.use("/forms/*", csrf());

// Bearer auth with custom verification
app.use(
  "/api/*",
  bearerAuth({
    verifyToken: async (token, c) => {
      return await tokenService.verify(token);
    },
  }),
);

// Basic auth for admin routes
app.use(
  "/admin/*",
  basicAuth({
    verifyUser: (username, password, c) => {
      return username === "admin" && password === process.env.ADMIN_PASSWORD;
    },
  }),
);

// IP restriction for internal routes
app.use(
  "/internal/*",
  ipRestriction({
    denyList: [],
    allowList: ["192.168.0.0/16", "10.0.0.0/8"],
  }),
);
```

**Techniques:**
- Import middleware from `hono/` subpaths (basicAuth, bearerAuth, csrf, secureHeaders, ipRestriction)
- Apply security headers first in middleware chain to protect all routes
- Use custom verifiers for dynamic authentication logic
- Configure CSRF protection for form-based applications
- Restrict IPs for admin and internal routes using CIDR notation
