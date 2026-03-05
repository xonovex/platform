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

# /xonovex-utility:instructions-consolidate – Consolidate project instruction files

Reduce AGENTS.md file count by deleting those with no unique content and standardizing the rest to a consistent bullet-list format.

## Arguments

- `--dry-run` (optional): Preview without modifying
- `--path <directory>` (optional): Root directory to scan (defaults to workspace root)

## Core Workflow

1. Use TodoWrite to track steps
2. Discover all AGENTS.md files via Glob
3. Classify each as unique or redundant → delete redundant → standardize unique → report

## Classification

**Redundant (delete):** Title restating dir name, one-line description inferable from dir name, or only guideline links predictable from package type

**Unique (keep):** Contains caveats, gotchas, non-obvious config, testing commands, workflow pipelines, style rules, or architectural decisions not inferable from directory structure

## Standardization

**Format:** `# Title` (humanized dir name) + flat bullet list — no `##` headings, no prose paragraphs

- Convert prose → bullet points; remove redundant descriptions and predictable guideline links
- Consolidate related bullets into groups separated by blank lines
- Keep all non-obvious technical details

## Implementation

- Skip root AGENTS.md (always unique)
- Redundant = removing title, description, and guideline links leaves nothing
- When unsure, keep the file

## Examples

```bash
/xonovex-utility:instructions-consolidate
/xonovex-utility:instructions-consolidate --dry-run
/xonovex-utility:instructions-consolidate --path packages/
```

## Error Handling

- No AGENTS.md found: report and exit
- All already consolidated: report and exit
- File permissions: warn and skip

## Safety

Use `--dry-run` to preview, never delete root AGENTS.md, never modify CLAUDE.md, report deleted and standardized files separately.
