# code-comments-remove: Identify Non-Essential Comments for Removal

Flag explanatory/obvious comments and commented-out code for removal; **preserve functional directives**. Read-only: produces a removal report grouped by category.

## Technique

- Detect comment syntax by language (C-style `//`, `/* */`; hash-style `#`)
- Preserve directives: `eslint-disable`, `prettier`, `@ts-expect-error`, `TODO`, `FIXME`, `go:generate`
- Remove: explanatory comments duplicating clear code, obvious comments, commented-out code blocks
- Report counts by category

```typescript
// Get the user ID        ← REMOVE (obvious)
const userId = user.id;
// eslint-disable-next-line no-unsafe-optional-chaining   ← KEEP (directive)
const role = user?.profile?.role?.name;
// const oldUserId = user.legacy_id   ← REMOVE (commented-out code)
```
