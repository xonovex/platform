# merge-resolve: Detect and Resolve Merge Conflicts

**Guideline:** Detect, classify, and resolve merge conflicts with auto-resolution for simple cases and strategies for complex conflicts.

**Rationale:** Simple conflicts like import statements and dependencies can be auto-resolved, while complex conflicts require manual decision-making based on strategies like keeping ours/theirs/merging intelligently.

**Example:**

```bash
# Detecting conflicts
$ git status --porcelain
UU src/services/auth.ts
AA package.json
DD src/utils/old-helper.ts

# Auto-resolve simple import statement conflicts
# package.json: merge both dependency versions

# Manual resolution for function body conflict
# Review auth.ts and choose strategy (ours/theirs/merge)

# Validate after resolution
$ npm run typecheck && npm run lint
TypeScript: PASS
ESLint: PASS

# Stage and complete merge
$ git add .
$ git commit -m "merge: resolve conflicts with feature branch"
```

**Techniques:**

- Detect conflicts via `git status --porcelain` (UU, AA, DD markers)
- Classify as simple (auto-resolvable: imports, dependencies, additions) or complex (manual: functions, types, logic)
- Auto-resolve imports by merging lists, dependencies by newer version, additions by including both
- For complex conflicts prompt for strategy: ours, theirs, or manual merge
- Validate with `npm run typecheck && npm run lint` before staging
- Stage resolved files with `git add` only after validation passes
