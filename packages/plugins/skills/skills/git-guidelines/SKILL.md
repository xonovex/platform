---
name: git-guidelines
description: >-
  Trigger on git operations and merge conflict situations. Use when working with git operations. Apply for conventional commits, merge conflict resolution, worktree management. Keywords: git, conventional commits, feat/fix/chore/docs, merge conflicts, worktrees, commit messages, history rewrite.
---

# Git Guidelines

## Core Principles

- **Conventional Commits** - Use type prefixes (feat, fix, chore, docs, refactor, test, ci), see [reference/commit.md](reference/commit.md)
- **Auto-Generate Messages** - Analyze changed files and context, see [reference/commit.md](reference/commit.md)
- **Isolated Development** - Use worktrees for feature branches, see [reference/worktree-create.md](reference/worktree-create.md)
- **Validate Before Merge** - Run typecheck/lint/build/test, see [reference/worktree-validate.md](reference/worktree-validate.md)

## Commit Operations

- **Auto-commit** - Analyze changes, infer type, generate message, optional push, see [reference/commit.md](reference/commit.md)

## Conflict Resolution

- **Detect and classify** - Find conflicts, suggest strategy (ours/theirs/merge), see [reference/merge-resolve.md](reference/merge-resolve.md)
- **Validate** - Run typecheck/lint after resolution before staging, see [reference/merge-resolve.md](reference/merge-resolve.md)

## Worktree Operations

- **Create** - `<worktree>-feature-<name>` directory with branch, see [reference/worktree-create.md](reference/worktree-create.md)
- **Commit** - Auto-commit with plan context, see [reference/worktree-commit.md](reference/worktree-commit.md)
- **Validate** - Pre-merge validation checkpoint, see [reference/worktree-validate.md](reference/worktree-validate.md)
- **Merge** - Merge feature back to source branch, see [reference/worktree-merge.md](reference/worktree-merge.md)
- **Cleanup** - Remove stale and merged worktrees, see [reference/worktree-cleanup.md](reference/worktree-cleanup.md)
- **Abandon** - Document and remove failed feature, see [reference/worktree-abandon.md](reference/worktree-abandon.md)

## Progressive Disclosure

### Commit Operations

- Read [reference/commit.md](reference/commit.md) - Commit with auto-generated conventional messages
- Read [reference/merge-resolve.md](reference/merge-resolve.md) - Detect and resolve merge conflicts with AI assistance

### Worktree Operations

- Read [reference/worktree-create.md](reference/worktree-create.md) - Create feature worktree with branch
- Read [reference/worktree-commit.md](reference/worktree-commit.md) - Auto-commit with plan context
- Read [reference/worktree-validate.md](reference/worktree-validate.md) - Pre-merge validation checkpoint
- Read [reference/worktree-merge.md](reference/worktree-merge.md) - Merge feature back to source branch
- Read [reference/worktree-cleanup.md](reference/worktree-cleanup.md) - Remove stale and merged worktrees
- Read [reference/worktree-abandon.md](reference/worktree-abandon.md) - Document and remove failed feature
