# async-without-await: Only Use Async When Function Contains Await

**Guideline:** Only use `async` when function contains `await` or must return Promise.

**Rationale:** Unnecessary `async` adds overhead and misleads readers about async behavior.

**Example:**

```typescript
// ❌ Bad: async without await
const middleware = async (c, next) => {
  c.set("user", getUser());
};

// ✅ Good: Remove async
const middleware = (c, next) => {
  c.set("user", getUser());
};

// ✅ Good: Return Promise explicitly if needed
const middleware = (c, next) => {
  c.set("user", getUser());
  return Promise.resolve();
};
```

**Techniques:**

- Review each `async` function for `await` expressions
- Remove `async` keyword if no `await` found
- Return `Promise.resolve()` explicitly if signature requires Promise
- Run ESLint to find @typescript-eslint/require-await violations
- Verify function behavior unchanged after removal
