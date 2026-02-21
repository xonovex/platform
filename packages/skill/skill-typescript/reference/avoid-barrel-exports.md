# avoid-barrel-exports: Import Directly from Specific Files

**Guideline:** Avoid `index.ts` barrel exports in subdirectories; import directly from specific files.

**Rationale:** Subdirectory barrels add indirection, obscure source, and cause circular dependency issues.

**Example:**

```typescript
// ✅ Good: Direct imports, no index.ts
import {csrfProtection} from "./middlewares/csrf.js";
import {authMiddleware} from "./middlewares/jwt.js";

// ❌ Bad: Subdirectory barrel exports
// src/middlewares/index.ts
export {authMiddleware} from "./jwt.js";
export {csrfProtection} from "./csrf.js";
```

**Techniques:**

- Remove `index.ts` from subdirectories (middlewares, utils, services, controllers)
- Update imports to specific files directly
- Use IDE refactoring tools to update all imports automatically
- Run tests to verify no broken imports
- Never create new subdirectory barrel exports
