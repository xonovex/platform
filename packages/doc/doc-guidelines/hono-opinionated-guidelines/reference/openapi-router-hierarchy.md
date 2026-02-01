# openapi-router-hierarchy: Use OpenAPIHono Throughout Router Hierarchy

**Guideline:** Use `OpenAPIHono` for all routers in the hierarchy (root, intermediate, and leaf) to ensure OpenAPI routes appear in the generated specification.

**Rationale:** OpenAPI metadata propagation works through the router chain only when all routers use `OpenAPIHono`. Mixing regular `Hono` with `OpenAPIHono` breaks metadata propagation, causing child OpenAPI routes to be "lost" and resulting in incomplete API documentation.

**Example:**

```typescript
// src/app.ts - Root router
// src/routes/v1/index.ts - Intermediate router

// src/routes/v1/items.ts - Leaf router
import {
  createRoute,
  OpenAPIHono,
  OpenAPIHono,
  OpenAPIHono,
  z,
} from "@hono/zod-openapi";

export function createApp() {
  // ✅ Use OpenAPIHono for root
  const app = new OpenAPIHono();

  app.use("*", logger());
  app.route("/api/v1", v1Router);

  return app;
}

// ✅ Use OpenAPIHono for intermediate router
export const v1Router = new OpenAPIHono();

v1Router.route("/items", itemsRouter);
v1Router.route("/users", usersRouter);

// ✅ Use OpenAPIHono for leaf router
export const itemsRouter = new OpenAPIHono();

const listItemsRoute = createRoute({
  method: "get",
  path: "/",
  responses: {
    200: {
      content: {"application/json": {schema: ItemListSchema}},
      description: "List of items",
    },
  },
});

itemsRouter.openapi(listItemsRoute, (c) => {
  // Handler implementation
});
```

**Techniques:**

- Import `OpenAPIHono` from `@hono/zod-openapi`
- Use `OpenAPIHono` for root application router
- Use `OpenAPIHono` for all intermediate routers (version routers, feature groupings)
- Use `OpenAPIHono` for all leaf routers (domain-specific routes)
- Mount routers normally with `.route()` method
- Verify all routes appear in generated `/openapi.json` spec
- Regular non-OpenAPI routes can coexist on OpenAPIHono routers
