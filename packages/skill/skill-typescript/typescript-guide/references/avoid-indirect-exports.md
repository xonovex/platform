# avoid-indirect-exports: Import Directly from the Defining Source

Import from the file/package that defines a symbol; never add an indirection layer. Two forms to avoid: (a) `index.ts` barrel exports in subdirectories — import from the specific file instead, since barrels obscure source and cause circular-dependency issues; (b) re-exporting external-package utilities through your own modules.

```typescript
// ✅ direct imports
import {getUserId, type JwtContext} from "@acme/shared-utils";
import {csrfProtection} from "./middlewares/csrf.js";

// ❌ subdirectory barrel: src/middlewares/index.ts
export {csrfProtection} from "./csrf.js";
// ❌ re-export of an external package: src/middlewares/jwt.ts
export {getUserId, type JwtContext} from "@acme/shared-utils";
```
