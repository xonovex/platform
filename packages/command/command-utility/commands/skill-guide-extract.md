---
description: >-
  Create or update a skill by extracting patterns from codebase and project
  instructions
model: sonnet
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - AskUserQuestion
  - Task
argument-hint: "[skill-name] [source-path] [--update] [--interactive] [--dry-run]"
---

# /xonovex-utility:skill-guide-extract – Extract Skill from Codebase

Creates or updates a skill by analyzing codebase patterns, AGENTS.md instructions, and source files.

## Spec Constraints

- `name`: 1-64 chars, lowercase kebab-case, no consecutive/leading/trailing hyphens, not the reserved words `anthropic`/`claude`, no XML tags, must match parent dir
- `description`: 1-1024 chars, imperative "Use when..." + triggers, third person, covers what + when (incl. non-obvious triggers)
- Body: target <500 lines / ~5000 tokens; push detail to `references/`
- Optional: `license` (string), `compatibility` (≤500 chars), `metadata` (string→string map), `allowed-tools` (experimental, space-separated)

## Arguments

- `skill-name` (required): Name for skill (e.g., `example-guide`)
- `source-path` (required): Path to analyze (e.g., `packages/example`)
- `--update` (optional): Update existing skill instead of creating new
- `--interactive` (optional): Ask which patterns to include before writing
- `--dry-run` (optional): Preview without writing files

## Core Workflow

1. **Discover Sources**: Find AGENTS.md in source-path; glob for source files by extension; scan git history for recurring corrections
2. **Analyze Instructions**: Extract patterns, conventions, file naming, architecture from project docs
3. **Analyze Code**: Sample source files for common patterns, types, naming conventions
4. **Mine Gotchas**: Collect non-obvious env-specific facts from PR comments, fix commits, and code review history
5. **Categorize Patterns**: Group into categories (architecture, types, testing, safety, etc.)
6. **Present Patterns**: If `--interactive`, ask user which categories to include
7. **Check Existing**: If skill exists and no `--update`, error; if `--update`, read existing skill
8. **Generate Skill**: Create SKILL.md with essentials, gotchas, examples, progressive disclosure
9. **Generate Reference Files**: Create reference files for each included category; pair each with a load-when trigger
10. **Write Files**: Create `<skills-dir>/{skill-name}/` structure
11. **Report Summary**: List created files and pattern counts

## Pattern Extraction

**From AGENTS.md:** Section headers → categories, tables → structured patterns, code blocks → examples

**From Source Files:** File naming conventions, type definitions, function signatures, comment conventions

**From git history:** Fix commits → gotchas; PR review comments → recurring corrections; patches → real-world failure modes

## Skill Structure

```
<skills-dir>/{skill-name}/
├── SKILL.md             # Main skill file
└── references/          # Progressive disclosure
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

- {3-5 key points as bullets — add only what the agent lacks}

## Gotchas

- {Non-obvious env-specific facts that defy reasonable assumptions}

## Example

{short code example showing the default approach}

## Progressive disclosure

- **references/{file}.md** - Load when {specific trigger condition}
```

**Reference File Format:**

```markdown
# {pattern-name}: {Title}

**Statement:** {one sentence}
**Rationale:** {why this pattern}
**Example:** {code block}
```

## Implementation Details

**Discovery**: `Glob` for `**/AGENTS.md`; detect language from file extensions (`.c/.h` → C99, `.ts` → TypeScript, `.py` → Python, `.lua` → Lua)

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
- No AGENTS.md found → warn, continue with source analysis only
- Skill exists without `--update` → error, suggest `--update`
- No patterns extracted → error, path may not contain relevant code
- Language not detected → ask user or default to generic

## Examples

```bash
# Create new skill from source code
/xonovex-utility:skill-guide-extract example-guide packages/example --interactive

# Update existing skill with new patterns
/xonovex-utility:skill-guide-extract typescript-guide packages/api --update

# Preview without writing
/xonovex-utility:skill-guide-extract python-guide services/ml --dry-run
```
