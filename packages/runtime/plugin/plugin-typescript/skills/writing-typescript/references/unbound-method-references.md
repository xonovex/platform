# unbound-method-references: Keep the Object Reference

Call methods through their object; destructuring detaches them from `this` and breaks at runtime. Caught by ESLint `@typescript-eslint/unbound-method`.

```typescript
// ❌ destructuring breaks this-binding
const {injectWebSocket} = createNodeWebSocket({app});
injectWebSocket(server);
const {addEventListener} = document;

// ✅ keep the object reference
const wsHelpers = createNodeWebSocket({app});
wsHelpers.injectWebSocket(server);
document.addEventListener("click", handler);
```
