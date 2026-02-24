---
description: Consolidate project instructions by removing redundant files and standardizing format
model: sonnet
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - AskUserQuestion
argument-hint: "[--dry-run] [--path <directory>]"
---

# /xonovex-utility:project-instructions-consolidate – Consolidate project instruction files

Reduce the number of AGENTS.md files by deleting those with no unique content and standardizing the remaining files to a consistent bullet-list format.

## Arguments

- `--dry-run` (optional): Preview changes without modifying
- `--path <directory>` (optional): Root directory to scan (defaults to workspace root)

## Core Workflow

1. Use TodoWrite to track steps
2. Discover all AGENTS.md files via Glob
3. Read each file and classify as unique or redundant
4. Delete redundant files
5. Standardize remaining files to consistent format
6. Report metrics

## Classification Rules

**Redundant (delete):** Files that only contain a title restating the directory name, a one-line description inferable from the directory name, or guideline links predictable from the package type (e.g., TypeScript packages linking to typescript-guidelines)

**Unique (keep):** Files containing caveats, gotchas, non-obvious configuration details, testing commands, workflow pipelines, style rules, or architectural decisions that cannot be inferred from the directory name or project structure

## Standardization Rules

**Format:** `# Title` followed by a flat bullet list — no `##` headings, no prose paragraphs

**Title:** Match the directory name (humanized)

**Content:**

- Convert prose paragraphs to bullet points
- Remove redundant descriptions that restate the directory name
- Remove guideline links predictable from package type/tags
- Consolidate related sections into single bullet groups separated by blank lines
- Keep all non-obvious technical details

## Example

**Before (redundant — delete):**

```markdown
# Shared Core

Core library for Xonovex TypeScript scripts (Node.js).

## Guidelines

- See [typescript-guidelines](../../skill/skill-typescript/guide/SKILL.md)
- See [vitest-guidelines](../../skill/skill-vitest/guide/SKILL.md)
```

**Before (unique — keep and standardize):**

```markdown
# ESLint Config Base

`"import"` must appear before `"node"` in `package.json` exports — jiti resolves
conditions in key order. `"import"` → `src/index.ts` (no build); `"node"` →
`dist/src/index.js` (requires build).
```

**After (standardized):**

```markdown
# ESLint Config Base

- `"import"` must appear before `"node"` in `package.json` exports — jiti resolves conditions in key order
- `"import"` → `src/index.ts` (no build); `"node"` → `dist/src/index.js` (requires build)
```

## Implementation

- Skip root AGENTS.md (always unique)
- A file is redundant if removing the title, description, and guideline links leaves nothing
- When unsure, keep the file
- Use blank lines between bullet groups to separate distinct topics

## Error Handling

- No AGENTS.md files found: report and exit
- All files already consolidated: report and exit
- File permissions: warn and skip

## Safety

- Recommend git commit before running
- Use `--dry-run` to preview
- Never delete root AGENTS.md
- Never modify CLAUDE.md files
- Report deleted files and standardized files separately

## Examples

```bash
/xonovex-utility:project-instructions-consolidate
/xonovex-utility:project-instructions-consolidate --dry-run
/xonovex-utility:project-instructions-consolidate --path packages/
```
