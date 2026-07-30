# unnecessary-async-keywords: Return Promises Directly Instead of Wrapping in Async

When a body just forwards a Promise-producing call (no `await`), drop `async` and return the Promise directly: handlers (e.g. Hono) accept a returned Promise. Same ESLint rule as async-without-await: `@typescript-eslint/require-await`.

```typescript
// ❌ async wrapper around a Promise-returning call
app.get("/users", async (c) => c.json(getUsers()));
// ✅ return the Promise directly, no async
app.get("/users", (c) => c.json(getUsers()));
```
