# env-access-bracket-notation: Use Dot Notation for process.env

**Guideline:** Use dot notation for `process.env` (not bracket notation) unless dynamic or special chars.

**Rationale:** Dot notation is more idiomatic and improves readability.

**Example:**

```typescript
// ❌ Bad: Bracket notation for static
const secret = process.env["JWT_SECRET"];

// ✅ Good: Dot notation for static
const secret = process.env.JWT_SECRET;

// OK: Bracket for dynamic keys
const configKey = "MY_VAR";
const value = process.env[configKey];

// OK: Bracket for special characters
const value = process.env["MY-VAR-WITH-DASHES"];
```

**Techniques:**
- Replace all `process.env["VAR"]` with `process.env.VAR` for static keys
- Keep bracket notation only for dynamic keys or special characters
- Run ESLint to find @typescript-eslint/dot-notation violations
- Verify runtime behavior unchanged after conversion
- Update all affected import statements and config files
