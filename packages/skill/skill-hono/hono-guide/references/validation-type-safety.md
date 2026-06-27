# validation-type-safety: Request Validation and Type Safety with Zod

## Guideline

Chain `zValidator` inline on the route so `c.req.valid('json')` is typed by inference through the route generics — no cast. Only when a controller is imported from a separate file and receives the base `Context` (where inference cannot flow) fall back to casting `c.req.valid`, and treat that cast as an unchecked assertion that can hide schema/type drift.

## Rationale

When the validator is chained on the route, Hono threads the schema's inferred type into the handler's `Context` generics, so `c.req.valid('json')` is fully typed with autocomplete and no assertion. A controller imported separately sees only the base `Context`, whose `c.req.valid()` returns `any`; there a cast restores compile-time types but is unchecked, so it must be reserved for that case and called out as a risk.

Schema design — defining schemas, `z.infer` types, `safeParse`, transforms, refinements — belongs to **zod-guide**. This file covers only the Hono glue: wiring a schema through `zValidator` and recovering types at the `c.req.valid` boundary.

## Example (preferred — inline method-chaining, inferred types)

```typescript
import {zValidator} from "@hono/zod-validator";
import {Hono} from "hono";
import {CreateUserSchema} from "../schemas/users.js";

const usersRouter = new Hono();

usersRouter.post("/", zValidator("json", CreateUserSchema), (c) => {
  // Typed by inference through the route generics — no cast needed
  const data = c.req.valid("json");

  const user = userService.create(data);
  return c.json(user, 201);
});
```

## Example (fallback — base `Context` controller in a separate file)

```typescript
import {zValidator} from "@hono/zod-validator";
import type {Context} from "hono";
import type {z} from "zod";
import {CreateUserSchema, type CreateUser} from "../schemas/users.js";

// Controller imported separately receives the base Context, where inference
// cannot flow. The cast is an UNCHECKED assertion — it can hide schema drift.
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

## Techniques

- Define Zod schemas for request payloads with TypeScript type exports
- Prefer chaining `zValidator` inline on the route so `c.req.valid('json')` is typed by inference — no cast
- Only when a controller is imported separately and gets the base `Context`, cast `c.req.valid` with the target parameter, specifying the return type (e.g., `CreateUser`)
- Treat such casts as unchecked assertions that can hide schema/type drift; reserve them for the base-`Context` limitation
- In zValidator error handlers, cast `result.error` to `z.ZodError`
- Use typed validation errors in error response functions
- Apply the same inline-chaining preference to query params and path params
