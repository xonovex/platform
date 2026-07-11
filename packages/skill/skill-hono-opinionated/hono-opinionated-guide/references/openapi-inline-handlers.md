# openapi-inline-handlers: Write OpenAPI Handlers Inline

Register handlers inline in `router.openapi(route, (c) => {...})` — never extract them to separate controller functions. Inference of `c.req.valid(...)` and the response types flows from the route schema and does not cross a function boundary; an extracted controller collapses to a generic `Response`. Keep the handler to request/response mapping and call service functions for logic.

```typescript
import {createRoute, z} from "@hono/zod-openapi";

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
      content: {"application/json": {schema: ItemListSchema}},
      description: "List of items",
    },
  },
});

itemsRouter.openapi(listItemsRoute, (c) => {
  const {page, limit} = c.req.valid("query"); // typed from the route
  const {items, total} = itemsService.listItems(page, limit);
  return c.json({items, pagination: {page, limit, total}}, 200);
});
```
