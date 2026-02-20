# avoid-reexports: Import from Source Packages Directly

**Guideline:** Import directly from source packages, don't re-export external utilities through your modules.

**Rationale:** Re-exports obscure code origin and complicate dependency understanding.

**Example:**

```typescript
// ✅ Good: Import directly from source
import {getUserId, type JwtContext} from "@acme/shared-utils";

// ❌ Bad: Re-export from shared packages
// src/middlewares/jwt.ts
export {getUserId, type JwtContext} from "@acme/shared-utils";

// src/middlewares/jwt.ts - only local exports
```

**Techniques:**

- Identify files re-exporting from external/shared packages
- Remove re-export statements from module exports
- Update consuming code to import directly from source packages
- Use IDE find-and-replace for bulk import updates
- Verify imports resolve correctly with type checking
- Run tests to ensure behavior unchanged
