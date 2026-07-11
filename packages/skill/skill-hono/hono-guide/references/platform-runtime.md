# platform-runtime: Platform-Specific Runtime Detection

Read env with `env<T>(c)` from `hono/adapter` (not `process.env`/`Deno.env`), branch on `getRuntimeKey()` for platform code (Workers returns `"workerd"`), and import `getConnInfo` from the platform-specific subpath. Platform quirks: Deno cache needs `wait: true`; Node needs explicit compression middleware.

```typescript
import {env, getRuntimeKey} from "hono/adapter";
import {getConnInfo} from "hono/cloudflare-workers";

app.get("/config", (c) => {
  const {DATABASE_URL} = env<{DATABASE_URL: string}>(c);
  return c.json({configured: !!DATABASE_URL});
});

app.use("*", async (c, next) => {
  const runtime = getRuntimeKey();
  c.header("X-Runtime", runtime === "workerd" ? "cloudflare" : runtime);
  await next();
});

app.get("/ip", (c) => c.json({ip: getConnInfo(c).remote.address}));
```
