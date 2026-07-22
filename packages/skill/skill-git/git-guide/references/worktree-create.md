# worktree-create: Create Feature Worktree

Create a sibling worktree with a feature branch for isolated development, deriving names from the current worktree.

## Naming convention

- Worktree directory: `<worktree>-feature-<feature-name>` (kebab-case name)
- Branch: `<worktree>/feature/<feature-name>`
- Use a type-descriptive prefix conveying the kind of change (`feature/`, `fix/`, `docs/`, `hotfix/`), not a single generic `feature` segment

| In worktree | Feature    | Directory                   | Branch                      |
| ----------- | ---------- | --------------------------- | --------------------------- |
| `services`  | `auth-fix` | `services-feature-auth-fix` | `services/feature/auth-fix` |

## Steps

```bash
# worktree name = basename of pwd; source = specified or `git branch --show-current`
git worktree add ../<worktree>-feature-<name> -b <worktree>/feature/<name> <source-branch>
git -C ../<worktree>-feature-<name> config branch.<branch>.mergeBackTo <source-branch>
```

## Gotchas

- The `<worktree>-feature-<name>` dir pattern is what merge/abandon/cleanup detect — non-conforming names break the workflow
- `branch.<branch>.mergeBackTo` is a custom key this workflow sets and reads; Git ignores it and provides no built-in behavior
- `mergeBackTo` is the only record of the source branch — without it `worktree-merge` can't find where to merge back
- A branch can't be checked out in two worktrees at once — use a different source branch or move the existing worktree
