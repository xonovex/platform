# controllers: Drop Async from Synchronous Controllers

Remove `async` from controller functions with no `await` in the body; keep it only when awaiting DB/API/other async work. `c.json()` returns fine from a sync function. Caught by ESLint `@typescript-eslint/require-await`.

```typescript
import type {Context} from "hono";

// ✅ sync body — no async
export function getUser(c: Context) {
  const {id} = c.req.valid("param");
  return c.json(userService.getById(id));
}
// ✅ keep async — body awaits
export async function createUser(c: Context) {
  const user = await userService.create(c.req.valid("json"));
  return c.json(user, 201);
}
```
