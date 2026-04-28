---
description: "Convert extracted insights from a category into AGENTS.md/CLAUDE.md bullet points. Use when the user asks to integrate insights into project instructions, update AGENTS.md from lessons learned, or fold session insights into docs. Keywords: insights, AGENTS.md, CLAUDE.md, integrate, project instructions, lessons learned, update docs."
---

# /xonovex-utility:insights-instructions-integrate – Convert Insights to AGENTS.md

Convert insights from a category into concise bullet points and integrate them into the appropriate AGENTS.md file.

## Arguments

- `category` (required): Category to convert (e.g., `testing`, `typescript`, `workflow`)
- `--dry-run` (optional): Preview without modifying
- `--agents-file <path>` (optional): Target AGENTS.md (default: auto-detect from `applies_to`)

## Core Workflow

1. Search `insights/` for category files, extract Problem/Solution pairs, group by topic
2. Locate target AGENTS.md — use `--agents-file` or auto-detect from `applies_to` field
3. Convert each insight to a concise bullet point matching AGENTS.md style
4. Merge into existing file — append to relevant bullet group or create new group
5. Mark processed insights as `applied: true`
6. Preview or write → report

## Conversion Rules

**Format:** Each insight becomes 1-2 bullet points in AGENTS.md style — backtick-wrapped names, `—` descriptions, `→` chains

**Deduplication:** Skip insights already covered by existing bullets; merge related insights into single bullets

**Placement:** Insert near related existing bullets; if no related section exists, append a new bullet group with blank line separator

**Brevity:** Only non-obvious details — skip anything inferable from directory name or project structure

## Auto-Detection

When `--agents-file` is not provided:

- Use `applies_to` field to match directory names or package names
- Search for nearest AGENTS.md in the matching directory
- If ambiguous, ask user via AskUserQuestion

## Examples

```bash
/xonovex-utility:insights-instructions-integrate testing
/xonovex-utility:insights-instructions-integrate typescript --dry-run
/xonovex-utility:insights-instructions-integrate api --agents-file packages/api/AGENTS.md
```

## Error Handling

- Missing category: ask user
- No insights found: suggest `/xonovex-utility:insights-extract [category]`
- No matching AGENTS.md: ask user for target path
- AGENTS.md not found at path: verify and abort

## Safety

Use `--dry-run` to preview, preserve existing AGENTS.md content and structure, never remove existing bullets, only append or merge.
