# platform-runtime: Platform-Specific Runtime Detection

**Guideline:** Use `env()` for unified environment variables and `getRuntimeKey()` for platform-specific code paths.

**Rationale:** Hono runs on different runtimes with different APIs and behaviors. Unified helpers enable portable code across all platforms.

**Example:**

```typescript
import {Hono} from "hono";
import {env, getRuntimeKey} from "hono/adapter";
import {getConnInfo} from "hono/cloudflare-workers";

const app = new Hono();

// Unified environment access
app.get("/config", (c) => {
  const {DATABASE_URL} = env<{DATABASE_URL: string}>(c);
  return c.json({configured: !!DATABASE_URL});
});

// Platform-specific logic
app.use("*", async (c, next) => {
  const runtime = getRuntimeKey();
  c.header("X-Runtime", runtime === "workerd" ? "cloudflare" : runtime);
  await next();
});

// Connection info
app.get("/ip", (c) => {
  const info = getConnInfo(c);
  return c.json({ip: info.remote.address});
});
```

**Techniques:**

- Import `env` and `getRuntimeKey` from `hono/adapter`
- Use `env(c)` instead of `process.env` or `Deno.env`
- Check runtime with `getRuntimeKey()` for platform-specific logic
- Import connection info from platform-specific paths
- Handle platform differences (Deno cache requires `wait: true`, Node.js needs compression middleware)
