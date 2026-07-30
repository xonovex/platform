# openapi-router-hierarchy: OpenAPIHono at Every Level

Use `OpenAPIHono` (not plain `Hono`) for root, intermediate, and leaf routers. A single plain `Hono` anywhere in the path drops its children's routes from the generated spec. Mount with `.route()` as usual.

```typescript
import {OpenAPIHono} from "@hono/zod-openapi";

const app = new OpenAPIHono(); // root
app.route("/api/v1", v1Router);

export const v1Router = new OpenAPIHono(); // intermediate
v1Router.route("/items", itemsRouter);

export const itemsRouter = new OpenAPIHono(); // leaf
itemsRouter.openapi(listItemsRoute, (c) => c.json(itemsService.list(), 200));
```
