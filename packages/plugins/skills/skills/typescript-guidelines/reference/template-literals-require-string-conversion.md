# template-literals-require-string-conversion: Explicitly Convert Numbers in Template Literals

**Guideline:** Explicitly convert numbers to strings in templates using `String(value)`, not implicit coercion.

**Rationale:** Explicit conversion is clearer and avoids relying on implicit type coercion.

**Example:**

```typescript
// ❌ Bad: Implicit number coercion
const port = 3000;
console.log(`Server running on port: ${port}`);

// ✅ Good: Explicit string conversion
const port = 3000;
console.log(`Server running on port: ${String(port)}`);

throw new Error(`Timeout after ${String(5000)}ms`);
```

**Techniques:**

- Wrap numeric values in `String()` when using template literals
- Apply to all primitive types in templates (numbers, booleans)
- Run ESLint to find @typescript-eslint/restrict-template-expressions violations
- Replace `` `text: ${value}` `` with `` `text: ${String(value)}` ``
- Verify tests pass after adding explicit conversions
