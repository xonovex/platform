# context-storage: Access Context Globally with AsyncLocalStorage

Apply `contextStorage()` middleware early, then call `getContext()` from a service layer to read Context without parameter drilling. Use sparingly: explicit parameters are more testable. On Cloudflare Workers enable the `nodejs_compat` flag.

```typescript
import {contextStorage, getContext} from "hono/context-storage";

app.use(contextStorage());
app.use("*", async (c, next) => {
  c.set("tenantId", c.req.header("X-Tenant-ID"));
  await next();
});

// service layer (separate file)
export function list() {
  const c = getContext();
  return db.items.findMany({where: {tenantId: c.get("tenantId")}});
}
```
