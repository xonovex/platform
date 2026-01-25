# application-structure: Application Factory and Router Organization

**Guideline:** Use factory functions to create Hono applications instead of exporting instances, and organize routes into separate router files by domain.

**Rationale:** Factory functions enable testability by creating fresh instances per test, support dependency injection, and improve isolation. Domain-based router organization provides clear separation of concerns and reduces merge conflicts in team environments.

**Example:**

```typescript
// src/app.ts - Main application factory
import {Hono} from "hono";
import {cors} from "hono/cors";
import {logger} from "hono/logger";
import {v1Router} from "./routes/v1/index.js";
import {v2Router} from "./routes/v2/index.js";

export function createApp() {
  const app = new Hono();

  // Global middleware
  app.use("*", logger());
  app.use("*", cors());

  // Mount routers by version
  app.route("/api/v1", v1Router);
  app.route("/api/v2", v2Router);

  return app;
}

// src/routes/v1/users.ts - Domain router
import {Hono} from "hono";
import {zValidator} from "@hono/zod-validator";
import * as controller from "../../controllers/users.controller.js";
import {CreateUserSchema} from "../../schemas/users.js";

export const usersRouter = new Hono();

usersRouter.post(
  "/",
  zValidator("json", CreateUserSchema),
  controller.createUser,
);
```

**Techniques:**
- Create `createApp()` factory function instead of exporting instances
- Configure global middleware inside the factory
- Mount domain routers using `app.route()`
- Organize routers in separate files by domain (users, items, etc.)
- Import controllers from dedicated controller files
- Keep route definitions focused on routing logic only
- Export factory functions for testability
