# async-without-await: Only Use Async When Function Contains Await

Mark a function `async` only when it contains `await` or must return a Promise — otherwise drop the keyword. Conversely, keep `async` wherever the body uses `await`; never rewrite an awaiting function into a `.then()` chain just to shed the keyword. Caught by ESLint `@typescript-eslint/require-await`.

```typescript
// ❌ async, no await
const middleware = async (c, next) => {
  c.set("user", getUser());
};
// ✅ drop async (sync body)
const middleware = (c, next) => {
  c.set("user", getUser());
};
// ✅ keep async when the body awaits — don't convert to a .then() chain
const loadConfig = async () =>
  JSON.parse(await readFile("config.json", "utf8"));
```
