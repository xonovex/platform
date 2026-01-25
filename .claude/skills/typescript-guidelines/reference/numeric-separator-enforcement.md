# numeric-separator-enforcement: Use Underscores in Large Numbers

**Guideline:** Use underscores in large numeric literals (e.g., `30_000`) for improved readability.

**Rationale:** Numeric separators improve readability and prevent misreading similar values.

**Example:**

```typescript
// ❌ Bad: Hard to read
const timeout = 30000;
const secondsPerDay = 86400;
const maxUsers = 1000000;

// ✅ Good: Clear and readable
const timeout = 30_000;
const secondsPerDay = 86_400;
const maxUsers = 1_000_000;

// OK: Smaller numbers don't need separators
const smallTimeout = 5000;
const port = 3000;
```

**Techniques:**
- Add underscores for numbers >= 10_000 grouped by thousands
- Do not add separators for smaller numbers (< 10_000)
- Run ESLint to find numeric-separators-style violations
- Use consistent grouping by thousands (e.g., 1_000_000 not 100_0000)
- Verify numeric values are correct after adding separators
