---
description: Commit and push changes in the current or a specified directory
model: haiku
allowed-tools:
  - Bash
  - Read
argument-hint: >-
  [message] [--type <feat|fix|docs|chore|refactor|test|ci>] [--path <path>]
  [--remote <remote>] [--branch <branch>] [--push] [--dry-run] [--interactive]
---

# /git-commit – Auto-Commit with Smart Messages

Commits changes in a directory with conventional commit format. Automatically generates commit messages based on changed files and plan context.

## Goal

- Auto-commit changes with conventional format
- Auto-generate messages from changed files and plan context
- Support interactive mode and optional push

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

## Core Workflow

**IMPORTANT: Always auto-commit immediately without prompting. Never ask the user to select a suggestion unless `--interactive` flag is explicitly provided.**

1. **Navigate**: Change to specified `--path` (if provided)
2. **Check Status**: Detect uncommitted changes and unpushed commits
3. **Analyze Changes**: Examine changed files to determine type and message
4. **Generate Message**: Pick the best commit message automatically
5. **Commit Immediately**: Stage and commit without asking - do NOT show suggestions or ask for confirmation
6. **Push**: Optionally push to specified remote and branch

Exception: Only show suggestions and prompt for selection when `--interactive` flag is explicitly provided.

## Smart Suggestion Logic

**Type Detection** – Analyze changed files:

- `test`: `*.test.ts`, `*.spec.ts`
- `docs`: `*.md` in docs/
- `ci`: CI config files
- `chore`: `package.json`, config files
- `feat`: New files in `src/`
- `fix`: Small modifications in `src/`
- `refactor`: Large changesets

**Message Generation** – Based on:

1. File patterns (extract package/feature names)
2. Diff stats (gauge scope)
3. Plan context (from git config)
4. Common operations (add/update/fix/remove)

## Implementation Steps

1. **Change Directory**: `cd <path>` (if `--path` specified)
2. **Check Git Status**: `git status --porcelain` for changes
3. **Analyze Changes** (if no message provided):
   - Run: `git diff --stat HEAD`, `git status --porcelain`
   - Detect type from file patterns
   - Generate the best commit message
4. **If Message Provided**: Use directly (skip generation)
5. **Determine Type**:
   - Use `--type` if provided
   - Use detected type from file analysis
   - Default to "chore"
6. **Commit Immediately**: `git add -A && git commit -m "<type>: <message>"`
   - Do NOT prompt or show suggestions (unless `--interactive`)
   - Just commit with the best generated message
7. **Push** (if --push):
   - Get remote: `git config branch.<branch>.remote` or use `--remote` or default to "origin"
   - Get branch: Use `--branch` or `git branch --show-current`
   - Push: `git push -o ci.skip <remote> HEAD:<branch>`

## Commit Format

**Format**: `<type>: <description>` (lowercase, no footers)

**Types**: `chore`, `feat`, `fix`, `docs`, `refactor`, `test`, `ci`

## Output

**Default mode (auto-commit, no prompts):**

```
Commit: /path/to/repo
Changed: 14 files (+588, -210)
Type: chore

Committed: chore: game-common consistency fixes (a3b2c1d)
```

**Interactive mode (`--interactive` flag):**

```
Suggestions:
  1. docs: add implementation guide
  2. feat: implement core logic
Select [1-2]: 1

Committed: docs: add implementation guide (a3b2c1d)
```

## Error Handling

- Error if not in a git repository
- Warning if no changes to commit (working tree clean)
- Error if commit fails (show git error)
- Warning if push fails (show git error, commit still succeeded)
- Error if `--path` directory doesn't exist
- Warning if large number of files changed (suggest splitting commit)

## Examples

```bash
# Auto-commit with generated message
/git-commit

# Interactive selection
/git-commit --interactive

# Explicit message and push
/git-commit "fix validation bug" --type fix --push

# Auto-commit in specific directory
/git-commit --path ./services

# Preview what would be committed
/git-commit --dry-run

# Commit and push to specific branch
/git-commit --branch develop --remote origin --push
```
