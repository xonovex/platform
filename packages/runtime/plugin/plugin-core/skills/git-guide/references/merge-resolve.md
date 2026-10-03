# merge-resolve: Detect and Resolve Merge Conflicts

Detect conflicts via `git status --porcelain` markers (`UU`, `AA`, `DD`). Classify and resolve:

- **Simple (auto-resolvable)**: imports → merge both lists; dependencies → keep newer version; additions → include both.
- **Complex (manual)**: function bodies, types, logic → choose ours / theirs / manual merge.

Validate with `npm run typecheck && npm run lint` before `git add` on resolved files, then commit.

```bash
git status --porcelain          # UU src/auth.ts, AA package.json, DD src/old.ts
# resolve per classification above
npm run typecheck && npm run lint
git add . && git commit -m "merge: resolve conflicts with feature branch"
```
