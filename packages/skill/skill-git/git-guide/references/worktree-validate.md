# worktree-validate: Pre-Merge Validation Checkpoint

Before merging a feature worktree, run in sequence and require each to pass:

```bash
npm run typecheck && npm run lint && npm run build && npm run test
```

If a plan is associated (`git config branch.<branch>.plan`), verify each plan success criterion is met before declaring READY TO MERGE.
