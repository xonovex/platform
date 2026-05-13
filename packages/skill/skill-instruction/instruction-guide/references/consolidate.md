# consolidate: Dedupe and Standardize AGENTS.md Across a Repo

Reduce AGENTS.md file count by deleting those with no unique content and standardizing the rest to a consistent bullet-list format.

## Core Workflow

1. Track steps in a task list
2. Discover all AGENTS.md files (glob)
3. Classify each as unique or redundant → delete redundant → standardize unique → report

## Classification

**Redundant (delete):** title restating dir name, one-line description inferable from dir name, only guideline links predictable from package type, or code summaries that restate what is already in the code

Code summaries include:

- Bin entries, dependencies, exports, engines derivable from `package.json`
- Build commands, env vars, task definitions derivable from `moon.yml`
- Module listings, function names, config values, CLI options derivable from source files
- Directory structure derivable from a file listing

**Unique (keep):** caveats, gotchas, non-obvious constraints, style rules, or architectural decisions not inferable from reading the code — things an agent would miss even after reading `package.json`, `moon.yml`, and source files

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
- **Respect nested precedence:** the AGENTS.md closest to the edited file wins. Don't merge a subproject's AGENTS.md up into the root — its scope is intentionally narrower.

## Error Handling

- No AGENTS.md found: report and exit
- All already consolidated: report and exit
- File permissions: warn and skip

## Safety

Preview before writing; never delete root AGENTS.md; report deleted and standardized files separately.

## Gotchas

- A 100-line AGENTS.md that only restates code is _more_ redundant than a 5-line one with a real gotcha
- Don't merge AGENTS.md files across distant subdirs — colocate signal with the code it applies to
- Root AGENTS.md is the project's entry-point doc; never auto-delete even if it looks thin
- A subproject's AGENTS.md may _look_ redundant against the root, but the closest-wins precedence means it's an intentional override — read both before classifying
