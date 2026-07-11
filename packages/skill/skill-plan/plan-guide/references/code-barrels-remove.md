# code-barrels-remove: Analyze Sub-Directory Barrel Exports for Removal

Identify sub-directory barrel files for removal and rewrite their imports to direct file paths. Read-only: produces a removal report, does not edit.

## Technique

- Find `index.ts`/`index.js` in subdirectories; **exclude the root `src/index.ts`**
- Confirm each is a barrel — it contains only re-exports
- Find imports from sub-barrel paths (`from "./subdir"`) and map each to its actual source files
- Report line savings, import count, and the deletion + import-migration plan

```typescript
import {Role, User} from "./auth"; // ❌ barrel (src/auth/index.ts)
import {Role} from "./auth/role"; // ✅ direct
import {User} from "./auth/user";
```
