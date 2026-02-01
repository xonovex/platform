# unnecessary-async-keywords: Remove Async Keyword Without Await Usage

**Guideline:** Remove `async` from functions that don't use `await`; frameworks allow returning Promises directly.

**Rationale:** Unnecessary `async` adds overhead and creates false async expectations.

**Example:**

```typescript
// ❌ Bad: async without await
async function getUser(id: string) {
  return database.users.find(id);
}

// ✅ Good: Remove async, return Promise directly
function getUser(id: string) {
  return database.users.find(id);
}

// ❌ Bad: Hono handler with async, no await
app.get("/users", async (c) => {
  return c.json(getUsers());
});

// ✅ Good: Remove async
app.get("/users", (c) => {
  return c.json(getUsers());
});
```

**Techniques:**

- Run ESLint to find @typescript-eslint/require-await violations
- Remove `async` keyword for each flagged function
- Verify return types still match (Promises ok without `async`)
- Return Promise directly without `async` wrapper
- Run tests to ensure behavior unchanged
- Commit with clear explanation of changes
