# openapi-spec-generation: Use app.doc() for Automatic OpenAPI Spec Generation

**Guideline:** Use the `app.doc()` method to automatically generate OpenAPI specifications from registered routes instead of maintaining static documents.

**Rationale:** Manual OpenAPI documents drift out of sync with implementation, require constant maintenance, and create a duplicate source of truth. The `app.doc()` method automatically extracts route metadata from route definitions, building complete documentation that stays synchronized with the actual implementation.

**Example:**

```typescript
import {createRoute, OpenAPIHono, z} from "@hono/zod-openapi";

const app = new OpenAPIHono();

// Define and register OpenAPI routes
const listItemsRoute = createRoute({
  method: "get",
  path: "/items",
  responses: {
    200: {
      content: {"application/json": {schema: ItemListSchema}},
      description: "List of items",
    },
  },
});

app.openapi(listItemsRoute, (c) => {
  // Handler implementation
});

// ✅ CORRECT - Auto-generate spec with app.doc()
app.doc("/openapi.json", {
  openapi: "3.1.0",
  info: {
    title: "Hono Backend API",
    version: "1.0.0",
    description: "API for managing items and uploads",
  },
  servers: [
    {
      url: "http://localhost:3000",
      description: "Development",
    },
  ],
});

// Access spec at: http://localhost:3000/openapi.json
// Spec automatically includes all registered OpenAPI routes
```

**Techniques:**

- Ensure your app uses `OpenAPIHono` (not regular `Hono`)
- Define routes using `createRoute()` with full request and response schemas
- Register routes using `router.openapi(route, handler)` method
- Call `app.doc(path, metadata)` to register spec endpoint
- Provide OpenAPI metadata: openapi version, info object, servers
- Access the spec at the registered path (e.g., `/openapi.json`)
- Never manually create or maintain the `paths` object
