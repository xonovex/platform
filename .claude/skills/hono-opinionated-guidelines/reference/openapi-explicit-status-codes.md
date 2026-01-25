# openapi-explicit-status-codes: Always Provide Explicit Status Codes in Responses

**Guideline:** Always provide explicit status codes in all `c.json()` calls within OpenAPI routes, as discriminated unions require them for proper TypeScript type narrowing.

**Rationale:** OpenAPI routes define multiple possible response types, each with a specific status code. TypeScript uses discriminated unions to represent these possibilities. Without explicit status codes, TypeScript cannot narrow the response type at compile time, causing type inference to break and losing type safety guarantees.

**Example:**

```typescript
import {createRoute, z} from "@hono/zod-openapi";

const getItemRoute = createRoute({
  method: "get",
  path: "/items/{id}",
  responses: {
    200: {
      content: {"application/json": {schema: ItemSchema}},
      description: "Item found",
    },
    404: {
      content: {"application/json": {schema: ProblemDetailsSchema}},
      description: "Item not found",
    },
  },
});

// ✅ CORRECT - Explicit status codes
itemsRouter.openapi(getItemRoute, (c) => {
  const {id} = c.req.valid("param");
  const item = itemsService.findById(id);

  if (!item) {
    return c.json(
      {
        type: "about:blank#not-found",
        title: "Not Found",
        status: 404,
        detail: "Item not found",
      },
      404,
    ); // Explicit status code
  }

  return c.json(item, 200); // Explicit status code
});

// ✅ CORRECT - Explicit 201 for creation
itemsRouter.openapi(createItemRoute, (c) => {
  const data = c.req.valid("json");
  const item = itemsService.create(data);
  return c.json(item, 201); // Explicit 201 for created resource
});
```

**Techniques:**
- Review all `c.json()` calls in OpenAPI handlers and add status code parameter
- Use 200 for success (OK) and 201 for creation (Created)
- Use 400 (Bad Request), 404 (Not Found), 500 (Server Error) for error cases
- Match status codes to those defined in route's `responses` object
- For Problem Details errors, include status in both response object and HTTP status
- Never rely on implicit default status codes in OpenAPI routes
