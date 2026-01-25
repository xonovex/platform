# code-barrels-remove: Analyze Sub-Directory Barrel Exports for Removal

**Guideline:** Identify sub-directory barrel files for removal and convert imports to direct file paths.

**Rationale:** Sub-barrels add indirection without clear benefits. Removing them simplifies navigation, makes imports explicit, and reduces circular dependency risks.

**Example:**
```typescript
// Before: Barrel indirection
import { User, Role } from './auth'  // From src/auth/index.ts

// After: Direct imports
import { User } from './auth/user'
import { Role } from './auth/role'
```

**Techniques:**
- Find all `index.ts`/`index.js` files in subdirectories (exclude root `src/index.ts`)
- Confirm each is a barrel by verifying it contains only re-exports
- Search for all imports from sub-barrel paths (pattern: `from "./subdir"`)
- Map each import to its actual source files to understand impact
- Update root barrel to convert relative paths to direct source context
- Track all usages across codebase and plan import migrations
- Generate report with line savings, import count, and deletion plan
