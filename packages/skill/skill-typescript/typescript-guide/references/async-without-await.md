# async-without-await: Only Use Async When Function Contains Await

Mark a function `async` only when it contains `await`; otherwise drop the keyword. Keep `async` wherever the body awaits — never rewrite an awaiting function into a `.then()` chain to shed it. Caught by ESLint `@typescript-eslint/require-await`.

```typescript
// ❌ async, no await
const middleware = async (c, next) => {
  c.set("user", getUser());
};
// ✅ drop async (sync body)
const middleware = (c, next) => {
  c.set("user", getUser());
};
```
