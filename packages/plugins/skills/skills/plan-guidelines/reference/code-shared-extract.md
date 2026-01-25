# code-shared-extract: Identify Common Patterns for Shared Library Extraction

**Guideline:** Identify common code patterns across packages for extraction to shared libraries.

**Rationale:** Duplicated code increases maintenance burden and inconsistency risks. Extracting to shared libraries promotes DRY, ensures consistent behavior, simplifies updates.

**Example:**
```typescript
// Found in 3 packages: auth validation logic
// packages/api/src/utils/validateToken.ts
// packages/web/src/utils/validateToken.ts (duplicate!)
// packages/admin/src/utils/validateToken.ts (slight variation)

// Extract to: packages/shared-auth/src/validateToken.ts
export function validateToken(token: string): Result<Payload, Error>
```

**Techniques:**
- Scan packages for repeated patterns: functions, components, hooks, middleware, types, constants
- Use semantic analysis to find similar logic even with different names
- Group patterns by similarity: identical code, equivalent logic, same interface
- Rank candidates by impact: occurrences × complexity × cross-package count
- Determine target: same-package `utils/`, cross-package `shared-*`, or new library
- Plan phased extraction with minimal breaking changes
- Validate extraction with typecheck, lint, and test suite
