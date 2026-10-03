# worktree-validate: Pre-Merge Validation Checkpoint

Before merging a feature worktree, discover the repository's required validation from
its project instructions, task runner, and toolchain configuration. Run every
applicable typecheck, lint, build, and test task, and require each to pass.

```bash
# Example only; use the repository-owned task runner and target set.
npx moon run <project>:typecheck <project>:lint <project>:build <project>:test
```

Report the exact commands and revisions validated. Functional acceptance criteria are
owned by the selected validation or planning capability, not Git metadata.
