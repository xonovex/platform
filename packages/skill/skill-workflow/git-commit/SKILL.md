---
description: "Commit and push changes with conventional-commit messages in the current or a specified directory. Use when the user asks to commit, push, save changes, or wrap up a piece of work. Keywords: commit, push, conventional commits, git commit, save changes, ship work."
---

# /xonovex-workflow:git-commit – Auto-Commit with Smart Messages

Commits changes in a directory with conventional commit format. Automatically generates commit messages based on changed files and plan context.

## Goal

- Auto-commit changes with conventional format
- Auto-generate messages from changed files and plan context
- Support interactive mode and optional push

## Core Workflow

**IMPORTANT: Always auto-commit immediately without prompting. Never ask the user to select a suggestion unless interactive mode was explicitly requested.**

1. **Navigate**: Change to specified path (if provided)
2. **Check Status**: Detect uncommitted changes and unpushed commits
3. **Analyze Changes**: Examine changed files to determine type and message
4. **Generate Message**: Pick the best commit message automatically
5. **Commit Immediately**: Stage and commit without asking - do NOT show suggestions or ask for confirmation
6. **Push**: Optionally push to specified remote and branch

Exception: Only show suggestions and prompt for selection when interactive mode is explicitly requested.

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

1. **Change Directory**: `cd <path>` (if path was specified)
2. **Check Git Status**: `git status --porcelain` for changes
3. **Analyze Changes** (if no message provided):
   - Run: `git diff --stat HEAD`, `git status --porcelain`
   - Detect type from file patterns
   - Generate the best commit message
4. **If Message Provided**: Use directly (skip generation)
5. **Determine Type**:
   - Use specified type if provided
   - Use detected type from file analysis
   - Default to "chore"
6. **Commit Immediately**: `git add -A && git commit -m "<type>: <message>"`
   - Do NOT prompt or show suggestions (unless interactive mode was requested)
   - Just commit with the best generated message
7. **Push** (if user requested push):
   - Get remote: `git config branch.<branch>.remote` or use specified remote, default to "origin"
   - Get branch: Use specified branch or `git branch --show-current`
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

**Interactive mode (when requested):**

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
- Error if specified path directory doesn't exist
- Warning if large number of files changed (suggest splitting commit)
