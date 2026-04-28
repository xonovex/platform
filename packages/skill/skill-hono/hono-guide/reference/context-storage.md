# context-storage: Access Context Globally with AsyncLocalStorage

**Guideline:** Use `contextStorage()` middleware with `getContext()` to access Hono Context outside route handlers, enabling cleaner service layer code without parameter drilling.

**Rationale:** AsyncLocalStorage provides global access to the current request context without parameter drilling through the call stack, enabling clean separation between routing and business logic while automatically cleaning up context after requests.

**Example:**

```typescript
import {Hono} from "hono";
import {contextStorage, getContext} from "hono/context-storage";

const app = new Hono();

// Enable context storage early
app.use(contextStorage());

// Set request metadata
app.use("*", async (c, next) => {
  c.set("requestId", crypto.randomUUID());
  c.set("tenantId", c.req.header("X-Tenant-ID"));
  await next();
});

app.get("/items", (c) => {
  const items = itemsService.list(); // No context passed
  return c.json(items);
});

// In service layer (separate file)
export function list() {
  const c = getContext();
  const tenantId = c.get("tenantId");
  const requestId = c.get("requestId");

  logger.info(`[${requestId}] Listing items for tenant ${tenantId}`);
  return db.items.findMany({where: {tenantId}});
}
```

**Techniques:**

- Import `contextStorage` and `getContext` from `hono/context-storage`
- Apply `contextStorage()` middleware early in middleware chain
- Call `getContext()` anywhere during request handling to access context
- For Cloudflare Workers enable `nodejs_compat` flag
- Use sparingly as explicit parameters are more testable
