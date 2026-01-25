# controllers: Remove Unnecessary Async from Synchronous Controllers

**Guideline:** Remove the `async` keyword from controller functions that don't use `await`, as unnecessary async adds overhead without providing benefits.

**Rationale:** Marking a function `async` without awaiting anything creates unnecessary Promise wrapping overhead, adds microtask queue delays, and misleads developers about async operations. Synchronous controllers are faster, clearer in intent, and produce simpler stack traces.

**Example:**

```typescript
import type {Context} from "hono";

// ✅ CORRECT - No async needed
export function getUser(c: Context) {
  const {id} = (c.req.valid as (target: string) => {id: string})("param");
  const user = userService.getById(id); // Synchronous call
  return c.json(user);
}

// ✅ CORRECT - Async needed for await
export async function createUser(c: Context) {
  const data = (c.req.valid as (target: string) => CreateUser)("json");
  const user = await userService.create(data); // Awaiting async operation
  return c.json(user, 201);
}

// ✅ CORRECT - No async, just returns response
export function listUsers(c: Context) {
  const users = userService.list();
  return c.json(users);
}
```

**Techniques:**
- Review controller function body for `await` keyword usage
- Remove `async` keyword if no `await` exists in function
- Ensure return statements work with or without async (c.json() compatible)
- Keep `async` only when awaiting DB queries, external APIs, or other async operations
- Check for performance improvements after removing unnecessary async
