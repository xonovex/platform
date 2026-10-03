# middleware-combine: Composing Middleware with some, every, except

From `hono/combine`: `some()` = OR (alternative auth methods, runs until one passes), `every()` = AND (layered checks), `except(path, mw)` = run `mw` on all paths except the given one(s). `except` accepts a string or string array.

```typescript
import {basicAuth} from "hono/basic-auth";
import {bearerAuth} from "hono/bearer-auth";
import {every, except, some} from "hono/combine";

// basic auth OR bearer token
app.use("/api/*", some(basicAuth({verifyUser}), bearerAuth({verifyToken})));

// authenticated AND admin
app.use(
  "/admin/*",
  every(bearerAuth({verifyToken}), async (c, next) => {
    if (c.get("user").role !== "admin")
      return c.json({error: "Forbidden"}, 403);
    await next();
  }),
);

// rate-limit everything except health/metrics
app.use("*", except(["/health", "/metrics"], authMiddleware()));
```
