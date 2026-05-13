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

## Commit Operations

- **Auto-commit** - Analyze changes, infer type, generate message, optional push, see [references/commit.md](references/commit.md)

## Conflict Resolution

- **Detect and classify** - Find conflicts, suggest strategy (ours/theirs/merge), see [references/merge-resolve.md](references/merge-resolve.md)
- **Validate** - Run typecheck/lint after resolution before staging, see [references/merge-resolve.md](references/merge-resolve.md)

## Worktree Operations

- **Create** - `<worktree>-feature-<name>` directory with branch, see [references/worktree-create.md](references/worktree-create.md)
- **Commit** - Auto-commit with plan context, see [references/worktree-commit.md](references/worktree-commit.md)
- **Validate** - Pre-merge validation checkpoint, see [references/worktree-validate.md](references/worktree-validate.md)
- **Merge** - Merge feature back to source branch, see [references/worktree-merge.md](references/worktree-merge.md)
- **Cleanup** - Remove stale and merged worktrees, see [references/worktree-cleanup.md](references/worktree-cleanup.md)
- **Abandon** - Document and remove failed feature, see [references/worktree-abandon.md](references/worktree-abandon.md)

## Gotchas

- `git pull` is `fetch` + `merge` — on a shared branch this creates spurious merge commits; prefer `pull --rebase` or `fetch` then explicit merge
- Detached HEAD: committing in this state silently loses commits when you `checkout` away — note the SHA or branch immediately
- `git rebase` rewrites history; force-pushing to a shared branch overwrites teammates' work — never force-push to `main`/`master`
- Hooks in `.git/hooks/` are not version-controlled — share via `core.hooksPath` pointing at a tracked directory
- `.gitignore` only ignores untracked files; already-tracked files need `git rm --cached` to stop tracking

## Progressive Disclosure

### Commit Operations

- Read [references/commit.md](references/commit.md) - Load when committing changes with auto-generated conventional messages
- Read [references/merge-resolve.md](references/merge-resolve.md) - Load when detecting and resolving merge conflicts

### Worktree Operations

- Read [references/worktree-create.md](references/worktree-create.md) - Load when creating a feature worktree with branch
- Read [references/worktree-commit.md](references/worktree-commit.md) - Load when committing inside a feature worktree with plan context
- Read [references/worktree-validate.md](references/worktree-validate.md) - Load when running pre-merge validation in a feature worktree
- Read [references/worktree-merge.md](references/worktree-merge.md) - Load when merging a feature worktree back to its source branch
- Read [references/worktree-cleanup.md](references/worktree-cleanup.md) - Load when removing stale or merged worktrees
- Read [references/worktree-abandon.md](references/worktree-abandon.md) - Load when documenting and removing an abandoned feature worktree
