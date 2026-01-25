# code-comments-remove: Identify Non-Essential Comments for Removal

**Guideline:** Analyze and remove non-essential comments while preserving functional directives.

**Rationale:** Self-documenting code preferred over comments. Remove explanatory, obvious, and commented-out code while preserving functional directives (eslint-disable, @ts-expect-error, TODO).

**Example:**
```typescript
// REMOVE: Redundant comment
// Get the user ID
const userId = user.id

// KEEP: Functional directive
// eslint-disable-next-line no-unsafe-optional-chaining
const role = user?.profile?.role?.name

// REMOVE: Commented-out code
// const oldUserId = user.legacy_id
```

**Techniques:**
- Detect comment syntax by language (C-style: `//`, `/* */`; hash-style: `#`)
- Distinguish functional directives (eslint, prettier, type hints) from explanatory comments
- Identify and categorize comments: explanatory, obvious, or commented-out code blocks
- Preserve critical directives: TODO, FIXME, @ts-expect-error, eslint-disable, go:generate, etc.
- Flag explanatory comments that duplicate clear code logic
- Find commented-out code blocks for removal (often outdated)
- Generate removal report with counts by category and impact summary
