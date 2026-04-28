# unbound-method-references: Keep Object References for Method Calls

**Guideline:** Keep object reference when calling methods, don't destructure them.

**Rationale:** Destructuring methods breaks `this` binding and causes runtime errors.

**Example:**

```typescript
// ❌ Bad: Destructuring breaks this
const {injectWebSocket} = createNodeWebSocket({app});
injectWebSocket(server); // May fail

// ✅ Good: Keep object reference
const wsHelpers = createNodeWebSocket({app});
wsHelpers.injectWebSocket(server); // Maintains this binding

// ❌ Bad: DOM API destructuring
const {addEventListener} = document;
addEventListener("click", handler); // Fails

// ✅ Good: Use object reference
document.addEventListener("click", handler);
```

**Techniques:**

- Never destructure methods: avoid `const {method} = object`
- Keep object reference: `const helpers = object`
- Call methods on object: `helpers.method()`
- Use object reference for DOM API calls
- Run ESLint to find @typescript-eslint/unbound-method violations
- Refactor destructured methods to keep object references
