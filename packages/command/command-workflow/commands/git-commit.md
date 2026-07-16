---
description: Commit and push changes in the current or a specified directory
allowed-tools:
  - Bash
  - Read
  - Skill
argument-hint: >-
  [message] [--type <feat|fix|docs|chore|refactor|test|ci>] [--path <path>]
  [--remote <remote>] [--branch <branch>] [--push] [--dry-run] [--interactive]
---

# /xonovex-workflow:git-commit — Auto-Commit with Smart Messages

## Arguments

`/git-commit [message] [--type <type>] [--path <path>] [--remote <remote>] [--branch <branch>] [--push] [--dry-run] [--interactive]`

- `message` (optional): Commit description (if omitted, auto-generates and uses best suggestion)
- `--type <type>` (optional): Commit type (auto-detected if not provided)
- `--path` (optional): Directory path for git commands (defaults to current directory)
- `--remote` (optional): Git remote to push to (defaults to "origin")
- `--branch` (optional): Remote branch to push to (defaults to current branch)
- `--push` (optional): Push after committing
- `--dry-run` (optional): Preview without committing
- `--interactive` (optional): Show suggestions and prompt for selection instead of auto-committing

## Delegation

Load an installed capability that provides safe Git commit/push operations and perform
its commit workflow with these arguments. This is a soft dependency: if no compatible
Git capability is installed, stop and name the missing capability instead of improvising
or changing another provider.
