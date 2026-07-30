# validation-type-safety: Request Validation and Type Safety with Zod

## Guideline

Chain `zValidator` inline on the route so `c.req.valid('json')` is typed by inference through the route generics, no cast. Only when a controller is imported from a separate file and receives the base `Context` (where inference cannot flow) fall back to casting `c.req.valid`, and treat that cast as an unchecked assertion that can hide schema/type drift.

Chaining threads the schema's inferred type into the handler's `Context` generics; a separately-imported controller sees only the base `Context`, whose `c.req.valid()` returns `any`. Schema design (`z.infer`, `safeParse`, transforms, refinements) belongs to **zod-guide**. This file covers only the Hono glue.

## Example (preferred: inline method-chaining, inferred types)

```typescript
import {zValidator} from "@hono/zod-validator";
import {Hono} from "hono";
import {CreateUserSchema} from "../schemas/users.js";

const usersRouter = new Hono();

usersRouter.post("/", zValidator("json", CreateUserSchema), (c) => {
  // Typed by inference through the route generics, no cast needed
  const data = c.req.valid("json");

  const user = userService.create(data);
  return c.json(user, 201);
});
```

## Example (fallback: base `Context` controller in a separate file)

```typescript
import {zValidator} from "@hono/zod-validator";
import type {Context} from "hono";
import type {z} from "zod";
import {CreateUserSchema, type CreateUser} from "../schemas/users.js";

// Controller imported separately receives the base Context, where inference
// cannot flow. The cast is an UNCHECKED assertion. It can hide schema drift.
export function createUser(c: Context) {
  const data = (c.req.valid as (target: string) => CreateUser)("json");

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

Apply the same inline-chaining preference to query and path params. In `zValidator` error handlers, cast `result.error` to `z.ZodError`.
