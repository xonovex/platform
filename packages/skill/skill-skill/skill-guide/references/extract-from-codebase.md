# extract-from-codebase: Distill a Skill from Codebase Patterns

## Contents

[Spec Constraints](#spec-constraints) · [Core Workflow](#core-workflow) · [Pattern Extraction](#pattern-extraction) · [Skill Structure](#skill-structure) · [Implementation Details](#implementation-details) · [Interactive Mode](#interactive-mode) · [Error Handling](#error-handling)

## Spec Constraints

Spec limits (name / description / body / optional-field rules) live in the parent SKILL.md — re-check after edits.

## Core Workflow

1. **Discover Sources**: Find AGENTS.md in source-path; glob for source files by extension; scan git history for recurring corrections
2. **Analyze Instructions**: Extract patterns, conventions, file naming, architecture from project docs
3. **Analyze Code**: Sample source files for common patterns, types, naming conventions
4. **Mine Gotchas**: Collect non-obvious env-specific facts from PR comments, fix commits, and code review history
5. **Categorize Patterns**: Group into categories (architecture, types, testing, safety, etc.)
6. **Present Patterns**: If interactive, ask user which categories to include
7. **Check Existing**: If skill exists and user didn't ask to update, error; if updating, read existing skill
8. **Generate Skill**: Create SKILL.md with essentials, gotchas, examples, progressive disclosure
9. **Generate Reference Files**: Create reference files for each included category; pair each with a load-when trigger
10. **Write Files**: Create `<skills-dir>/{skill-name}/` structure (path depends on the agent harness)
11. **Report Summary**: List created files and pattern counts

## Pattern Extraction

### From AGENTS.md

Section headers → categories, tables → structured patterns, code blocks → examples

### From Source Files

File naming conventions, type definitions, function signatures, comment conventions

### From git history

Fix commits → gotchas; PR review comments → recurring corrections; patches → real-world failure modes

## Skill Structure

```
<skills-dir>/{skill-name}/
├── SKILL.md             # Main skill file
└── references/          # Progressive disclosure
    └── pattern-*.md
```

### SKILL.md Format

```markdown
---
name: {skill-name}
description: "{one-line description}"
---

# {Title}

## Requirements

- {version/dependency requirements}

## Essentials

- {3-5 key points as bullets — add only what the agent lacks}

## Gotchas

- {Non-obvious env-specific facts that defy reasonable assumptions}

## Example

{short code example showing the default approach}

## Progressive disclosure

- **references/{file}.md** - Load when {specific trigger condition}
```

### Reference File Format

```markdown
# {pattern-name}: {Title}

## {Facet}

{statement and rationale, then a bad -> good example}
```

## Implementation Details

**Discovery**: Glob for `**/AGENTS.md`; detect language from file extensions (`.c/.h` → C99, `.ts` → TypeScript, `.py` → Python, `.lua` → Lua)

**Pattern Deduplication**: Skip patterns already in base skill (e.g., `c99-guide` for `c99-game-guide`)

**Naming**: Skill names use kebab-case; reference files use kebab-case matching pattern name

**Style Matching**: Follow existing skill style in `<skills-dir>/` (read 2-3 for reference)

## Interactive Mode

Present categories with descriptions, allow multi-select:

```
Which pattern categories should I include?
[ ] Architecture (file naming, directory structure)
[ ] Types (type definitions, patterns)
[ ] Testing (assertions, test patterns)
[ ] Safety (validations, error handling)
```

## Error Handling

- Source path not found → error with suggestion
- No agent instructions found → warn, continue with source analysis only
- Skill exists without update requested → error, ask user to confirm update
- No patterns extracted → error, path may not contain relevant code
- Language not detected → ask user or default to generic
