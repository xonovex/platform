# application-structure: Application Factory and Router Organization

Return the app from a `createApp()` factory instead of exporting a module-level instance (fresh instance per test, supports DI). Configure global middleware inside the factory; mount domain routers with `app.route()`; keep controllers in dedicated files.

```typescript
// src/app.ts
import {Hono} from "hono";
import {cors} from "hono/cors";
import {logger} from "hono/logger";
import {v1Router} from "./routes/v1/index.js";

export function createApp() {
  const app = new Hono();
  app.use("*", logger());
  app.use("*", cors());
  app.route("/api/v1", v1Router);
  return app;
}
```
