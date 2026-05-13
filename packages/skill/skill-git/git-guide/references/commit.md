# commit: Auto-Commit with Conventional Messages

Commit changes with auto-generated conventional-commit messages based on changed files and (optionally) plan context. Default behaviour: commit immediately without prompting.

## Goal

- Auto-commit changes with conventional format `<type>: <description>`
- Auto-generate messages from changed files (and plan context where available)
- Support interactive mode (suggestions + selection) and optional push

## Core Workflow

**Default: always auto-commit immediately without prompting.** Show suggestions only when interactive mode is explicitly requested.

1. **Navigate** — change to the specified path if one was provided
2. **Check Status** — detect uncommitted changes and unpushed commits
3. **Analyze Changes** — examine changed files to determine type and message
4. **Generate Message** — pick the best conventional-commit message
5. **Commit Immediately** — stage and commit; do not prompt unless interactive mode is requested
6. **Push** (optional) — push to specified remote and branch

## Type Detection

Analyze changed files:

| Pattern                                   | Type       |
| ----------------------------------------- | ---------- |
| `*.test.ts`, `*.spec.ts`                  | `test`     |
| `*.md` under `docs/`                      | `docs`     |
| CI config files                           | `ci`       |
| `package.json`, lockfiles, project config | `chore`    |
| New files in `src/`                       | `feat`     |
| Small modifications in `src/`             | `fix`      |
| Large changesets across many files        | `refactor` |

Default: `chore` if nothing else matches.

## Message Generation

Derive the description from:

1. File patterns (extract package / feature names)
2. Diff stats (gauge scope)
3. Plan context (from `git config branch.<branch>.plan` if set)
4. Common operations (add / update / fix / remove)

## Implementation Steps

1. `cd <path>` (if specified)
2. `git status --porcelain` to detect changes
3. If no message provided:
   - `git diff --stat HEAD` and `git status --porcelain` to analyze
   - Detect type from file patterns
   - Generate the best description
4. If message provided: use it directly
5. Determine type: specified > detected > `chore`
6. Stage and commit: `git add -A && git commit -m "<type>: <description>"`
7. Push if requested:
   - Remote: `git config branch.<branch>.remote` or specified, default `origin`
   - Branch: specified or `git branch --show-current`
   - `git push -o ci.skip <remote> HEAD:<branch>`

## Commit Format

- `<type>: <description>` — lowercase, no footers
- Types: `chore`, `feat`, `fix`, `docs`, `refactor`, `test`, `ci`

## Output

**Default mode (no prompts):**

```
Commit: /path/to/repo
Changed: 14 files (+588, -210)
Type: chore

Committed: chore: game-common consistency fixes (a3b2c1d)
```

**Interactive mode:**

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
- Warning if push fails (show git error; commit already succeeded)
- Error if specified path doesn't exist
- Warning if a very large number of files changed (suggest splitting)

## Gotchas

- Auto-detecting `refactor` when the changeset spans many files often misses a more specific intent (`feat` / `fix`) — interactive mode is safer for very large diffs
- Pushing with `-o ci.skip` skips CI on the push; remove the option when you actually want CI to run
- Default mode commits without asking — risky on dirty working trees with mixed-intent changes; either pre-stage or use interactive mode
- A stale `branch.<branch>.plan` git config can inject irrelevant plan context into the message — clear it when starting unrelated work
