# worktree-create: Create Feature Worktree

Create a sibling worktree with a feature branch for isolated development, deriving names from the current worktree.

## Naming convention

- Worktree directory: `<worktree>-feature-<feature-name>` (kebab-case name)
- Branch: `<worktree>/feature/<feature-name>`

| In worktree | Feature        | Directory                   | Branch                      |
| ----------- | -------------- | --------------------------- | --------------------------- |
| `services`  | `auth-fix`     | `services-feature-auth-fix` | `services/feature/auth-fix` |
| `api`       | `new-endpoint` | `api-feature-new-endpoint`  | `api/feature/new-endpoint`  |

## Steps

```bash
# worktree name = basename of pwd; source = specified or `git branch --show-current`
git worktree add ../<worktree>-feature-<name> -b <worktree>/feature/<name> <source-branch>
git -C ../<worktree>-feature-<name> config branch.<branch>.mergeBackTo <source-branch>
```

Optionally associate a plan via `git config branch.<branch>.plan <path>`.

## Gotchas

- The `<worktree>-feature-<name>` dir pattern is what merge/abandon/cleanup detect — non-conforming names break the workflow
- `branch.<branch>.mergeBackTo` and `branch.<branch>.plan` are custom keys this workflow sets/reads; git ignores them, no built-in behavior
- `mergeBackTo` is the only record of the source branch — without it `worktree-merge` can't find where to merge back
- A branch can't be checked out in two worktrees at once — use a different source branch or move the existing worktree
