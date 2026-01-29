# openapi-inline-handlers: Use Inline Handlers for OpenAPI Routes

**Guideline:** Implement handler logic inline within OpenAPI route definitions instead of separate controller functions to enable proper TypeScript type inference from route schemas.

**Rationale:** Hono's OpenAPI integration uses TypeScript type inference to match route schemas with handlers. This only works with inline handlers because TypeScript cannot infer types across function boundaries, and separate controllers return generic `Response` types instead of being automatically typed by route definitions.

**Example:**

```typescript
import {createRoute, OpenAPIHono, z} from "@hono/zod-openapi";

// Define route with schemas
const listItemsRoute = createRoute({
  method: "get",
  path: "/items",
  request: {
    query: z.object({
      page: z.coerce.number().default(1),
      limit: z.coerce.number().default(10),
    }),
  },
  responses: {
    200: {
      content: {
        "application/json": {
          schema: z.object({
            items: z.array(ItemSchema),
            pagination: z.object({
              page: z.number(),
              limit: z.number(),
              total: z.number(),
            }),
          }),
        },
      },
      description: "List of items",
    },
  },
});

// ✅ CORRECT - Inline handler with type inference
itemsRouter.openapi(listItemsRoute, (c) => {
  // Types automatically inferred from route definition
  const {page, limit} = c.req.valid("query");
  const {items, total} = itemsService.listItems(page, limit);
  return c.json({items, pagination: {page, limit, total}}, 200);
});
```

**Techniques:**

- Define route using `createRoute()` with full request and response schemas
- Register route with `router.openapi(route, (c) => {...})` with inline handler
- Access validated data via `c.req.valid()` to get automatically typed request data
- Return responses with `c.json(data, statusCode)` and explicit status codes
- Keep route handler focused on request/response mapping
- Extract complex business logic to service functions called from handler
