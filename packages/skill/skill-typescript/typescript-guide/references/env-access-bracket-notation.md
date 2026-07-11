# env-access-bracket-notation: Use Dot Notation for process.env

Access static `process.env` keys with dot notation; keep brackets only for dynamic keys or names with special characters. Caught by ESLint `@typescript-eslint/dot-notation`.

```typescript
const secret = process.env.JWT_SECRET; // ✅ static
const value = process.env[configKey]; // ✅ dynamic key
const dashed = process.env["MY-VAR-WITH-DASHES"]; // ✅ special chars
const bad = process.env["JWT_SECRET"]; // ❌ static via brackets
```
