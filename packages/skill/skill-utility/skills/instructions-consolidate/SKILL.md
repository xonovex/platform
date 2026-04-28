---
description: "Consolidate scattered AGENTS.md/CLAUDE.md files, remove redundancy, standardize format. Use when the user asks to clean up instructions, deduplicate AGENTS files, or unify project instruction style. Keywords: AGENTS.md, CLAUDE.md, consolidate, dedupe, cleanup, standardize, instructions."
---

# /xonovex-utility:instructions-consolidate – Consolidate project instruction files

Reduce AGENTS.md file count by deleting those with no unique content and standardizing the rest to a consistent bullet-list format.

## Core Workflow

1. Use TodoWrite to track steps
2. Discover all AGENTS.md files via Glob
3. Classify each as unique or redundant → delete redundant → standardize unique → report

## Classification

**Redundant (delete):** Title restating dir name, one-line description inferable from dir name, only guideline links predictable from package type, or code summaries that restate what is already in the code

Code summaries include:

- Bin entries, dependencies, exports, engines derivable from package.json
- Build commands, env vars, task definitions derivable from moon.yml
- Module listings, function names, config values, CLI options derivable from source files
- Directory structure derivable from file listing

**Unique (keep):** Contains caveats, gotchas, non-obvious constraints, style rules, or architectural decisions not inferable from reading the code — things an agent would miss even after reading package.json, moon.yml, and source files

## Standardization

**Format:** `# Title` (humanized dir name) + flat bullet list — no `##` headings, no prose paragraphs

- Convert prose → bullet points; remove redundant descriptions and predictable guideline links
- Remove bullets that summarize code (an agent can read the code itself)
- Consolidate related bullets into groups separated by blank lines
- Keep only non-obvious technical details: gotchas, caveats, non-obvious constraints

## Implementation

- Skip root AGENTS.md (always unique)
- Redundant = removing title, description, guideline links, and code summaries leaves nothing
- When unsure, keep the file

## Error Handling

- No AGENTS.md found: report and exit
- All already consolidated: report and exit
- File permissions: warn and skip

## Safety

Preview before writing, never delete root AGENTS.md, never modify CLAUDE.md, report deleted and standardized files separately.
