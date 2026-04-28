---
description: "Create or update a guideline skill by extracting patterns from the codebase and AGENTS.md/CLAUDE.md. Use when the user asks to extract a skill from existing code, distill patterns into a SKILL.md, or capture project conventions. Keywords: skill extraction, SKILL.md, codebase patterns, AGENTS.md, distill conventions, guideline skill."
---

# /xonovex-utility:skill-guide-extract – Extract Skill from Codebase

Creates or updates a skill by analyzing codebase patterns, AGENTS.md/CLAUDE.md instructions, and source files.

## Core Workflow

1. **Discover Sources**: Find AGENTS.md/CLAUDE.md in source-path; glob for source files by extension
2. **Analyze Instructions**: Extract patterns, conventions, file naming, architecture from project docs
3. **Analyze Code**: Sample source files for common patterns, types, naming conventions
4. **Categorize Patterns**: Group into categories (architecture, types, testing, safety, etc.)
5. **Present Patterns**: If interactive, ask user which categories to include
6. **Check Existing**: If skill exists and user didn't ask to update, error; if updating, read existing skill
7. **Generate Skill**: Create SKILL.md with essentials, examples, progressive disclosure
8. **Generate Reference Files**: Create reference files for each included category
9. **Write Files**: Create `.claude/skills/{skill-name}/` structure
10. **Report Summary**: List created files and pattern counts

## Pattern Extraction

**From AGENTS.md/CLAUDE.md:** Section headers → categories, tables → structured patterns, code blocks → examples

**From Source Files:** File naming conventions, type definitions, function signatures, comment conventions

## Skill Structure

```
.claude/skills/{skill-name}/
├── SKILL.md           # Main skill file
└── reference/           # Progressive disclosure
    └── pattern-*.md
```

**SKILL.md Format:**

```markdown
---
name: {skill-name}
description: "{one-line description}"
---

# {Title}

## Requirements

- {version/dependency requirements}

## Essentials

- {3-5 key guidelines as bullets}

## Example

{short code example}

## Progressive disclosure

- **reference/{file}.md** - When {trigger condition}
```

**Reference File Format:**

```markdown
# {pattern-name}: {Title}

**Guideline:** {one sentence}

**Rationale:** {why this pattern}

**Example:**
{code block}
```

## Implementation Details

**Discovery**: `Glob` for `**/AGENTS.md`, `**/CLAUDE.md`; detect language from file extensions (`.c/.h` → C99, `.ts` → TypeScript, `.py` → Python, `.lua` → Lua)

**Pattern Deduplication**: Skip patterns already in base skill (e.g., `c99-guide` for `c99-game-guide`)

**Naming**: Skill names use kebab-case; reference files use kebab-case matching pattern name

**Style Matching**: Follow existing skill style in `.claude/skills/` (read 2-3 for reference)

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
- No AGENTS.md/CLAUDE.md found → warn, continue with source analysis only
- Skill exists without update requested → error, ask user to confirm update
- No patterns extracted → error, path may not contain relevant code
- Language not detected → ask user or default to generic
