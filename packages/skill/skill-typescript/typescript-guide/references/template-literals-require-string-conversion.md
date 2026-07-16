# template-literals-require-string-conversion: Convert Numbers with String()

Wrap non-string primitives in `String(value)` inside template literals. Caught by ESLint `@typescript-eslint/restrict-template-expressions`.

```typescript
const port = 3000;
console.log(`Server running on port: ${String(port)}`); // ✅
console.log(`Server running on port: ${port}`); // ❌ implicit coercion
```
