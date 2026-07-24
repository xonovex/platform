---
name: git-guide
description: "Use when running git operations or resolving repo-state issues. Triggers on prompts about commit messages, conventional commits (feat/fix/chore/docs), merge conflicts, rebases, worktrees, feature-worktree create / merge / abandon / cleanup, branch cleanup, or history rewrites, even when the user doesn't say 'git'."
---

# Git Guidelines

## Core Principles

- **Conventional Commits** - Use type prefixes (feat, fix, chore, docs, refactor, test, ci), see [references/commit.md](references/commit.md)
- **Auto-Generate Messages** - Analyze changed files and context, see [references/commit.md](references/commit.md)
- **Isolated Development** - Use worktrees for feature branches, see [references/worktree-create.md](references/worktree-create.md)
- **Validate Before Merge** - Run typecheck/lint/build/test, see [references/worktree-validate.md](references/worktree-validate.md)
- **Publish and Rebase** - Push a branch upstream and rebase onto the base before opening a PR / MR, see [references/push.md](references/push.md)
- **Explicit Effects** - Worktree create, merge, and cleanup default to preview and mutate only for an explicit apply

## Commit Operations

- **Auto-commit** - Analyze changes, infer type, generate message, optional push, see [references/commit.md](references/commit.md)

## Branch Operations

- **Push + rebase-onto-base** - Publish a branch upstream and keep its PR / MR diff just this change; the host skill (`github-guide` / `gitlab-guide`) drives the PR itself, see [references/push.md](references/push.md)

## Conflict Resolution

- **Detect and classify** - Find conflicts, suggest strategy (ours/theirs/merge), see [references/merge-resolve.md](references/merge-resolve.md)
- **Validate** - Run typecheck/lint after resolution before staging, see [references/merge-resolve.md](references/merge-resolve.md)

## Worktree Operations

- **Create** - Preview or apply creation of a `<worktree>-feature-<name>` directory and branch, see [references/worktree-create.md](references/worktree-create.md)
- **Validate** - Pre-merge validation checkpoint, see [references/worktree-validate.md](references/worktree-validate.md)
- **Merge** - Preview or apply integration into the source branch without cleanup, see [references/worktree-merge.md](references/worktree-merge.md)
- **Cleanup** - Preview or apply removal of exact stale or merged targets, see [references/worktree-cleanup.md](references/worktree-cleanup.md)
- **Abandon** - Inspect and document a stopped feature without mutation, see [references/worktree-abandon.md](references/worktree-abandon.md)

## Gotchas

- `git pull` is `fetch` + `merge` — on a shared branch this creates spurious merge commits; prefer `pull --rebase` or `fetch` then explicit merge
- Detached HEAD: committing in this state silently loses commits when you `checkout` away — note the SHA or branch immediately
- Hooks in `.git/hooks/` are not version-controlled — share via `core.hooksPath` pointing at a tracked directory
- A feature `worktree-merge` integrates a branch into its **parent** branch; landing on the mainline goes through push + PR + CI review, never a direct local merge to `main`
- Merge, abandon, and cleanup are separate operations — none may silently perform another
- Long-lived feature branches drift and conflict — keep them short-lived; integrate work spanning sessions incrementally behind feature flags / branch-by-abstraction
- Deleting a worktree directory by hand leaves a stale admin entry — run `git worktree prune`; and each worktree needs its own dependency install (`node_modules` is not shared across worktrees)

## Progressive Disclosure

### Commit Operations

- Read [references/commit.md](references/commit.md) - Load when committing changes with auto-generated conventional messages
- Read [references/push.md](references/push.md) - Load when publishing a branch upstream and rebasing it onto the base before opening a PR / MR
- Read [references/merge-resolve.md](references/merge-resolve.md) - Load when detecting and resolving merge conflicts

### Worktree Operations

- Read [references/worktree-create.md](references/worktree-create.md) - Load when creating a feature worktree with branch
- Read [references/worktree-validate.md](references/worktree-validate.md) - Load when running pre-merge validation in a feature worktree
- Read [references/worktree-merge.md](references/worktree-merge.md) - Load when merging a feature worktree back to its source branch
- Read [references/worktree-cleanup.md](references/worktree-cleanup.md) - Load when removing stale or merged worktrees
- Read [references/worktree-abandon.md](references/worktree-abandon.md) - Load when documenting and removing an abandoned feature worktree
