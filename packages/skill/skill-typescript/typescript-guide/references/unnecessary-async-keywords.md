# unnecessary-async-keywords: Return Promises Directly Instead of Wrapping in Async

When a function's body just forwards a Promise-producing call (no `await`), drop `async` and return the Promise directly — handlers (e.g. Hono) accept a returned Promise. If the body actually uses `await`, keep `async`. Same ESLint rule as async-without-await: `@typescript-eslint/require-await`.

```typescript
// ❌ async wrapper around a Promise-returning call
async function getUser(id: string) {
  return database.users.find(id);
}
app.get("/users", async (c) => c.json(getUsers()));
// ✅ return the Promise directly, no async
function getUser(id: string) {
  return database.users.find(id);
}
app.get("/users", (c) => c.json(getUsers()));
```
