# template-literals-require-string-conversion: Convert Numbers with String()

Wrap non-string primitives (numbers, booleans) in `String(value)` inside template literals rather than relying on implicit coercion. Caught by ESLint `@typescript-eslint/restrict-template-expressions`.

```typescript
const port = 3000;
console.log(`Server running on port: ${String(port)}`); // ✅
throw new Error(`Timeout after ${String(5000)}ms`); // ✅
console.log(`Server running on port: ${port}`); // ❌ implicit coercion
```
