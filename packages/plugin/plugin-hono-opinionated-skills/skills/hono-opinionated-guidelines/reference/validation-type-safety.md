# validation-type-safety: Request Validation and Type Safety with Zod

**Guideline:** Cast `c.req.valid` to specify return types and cast validation errors to `z.ZodError` to restore type safety when using Hono's base Context type.

**Rationale:** Hono's Context type without generic validation types returns `any` from `c.req.valid()`, losing type safety. Type assertions restore compile-time type checking without complex nested generics, enabling autocomplete and error detection while keeping signatures simple.

**Example:**

```typescript
import {zValidator} from "@hono/zod-validator";
import type {Context} from "hono";
import type {z} from "zod";
import {CreateUserSchema, type CreateUser} from "../schemas/users.js";

// Controller with type-safe validation
export function createUser(c: Context) {
  // Cast c.req.valid to specify return type
  const data = (c.req.valid as (target: string) => CreateUser)("json");

  // Now `data` has full type information
  const user = userService.create(data);
  return c.json(user, 201);
}

usersRouter.post(
  "/",
  zValidator("json", CreateUserSchema, (result, c) => {
    if (!result.success) {
      // Cast to z.ZodError for type-safe error processing
      return badRequest(c, result.error as z.ZodError);
    }
  }),
  controller.createUser,
);
```

**Techniques:**

- Define Zod schemas for request payloads with TypeScript type exports
- In controllers using base `Context`, cast `c.req.valid` with target parameter
- Specify the return type in the cast (e.g., `CreateUser`)
- In zValidator error handlers, cast `result.error` to `z.ZodError`
- Use typed validation errors in error response functions
- Apply same pattern to query params and path params
