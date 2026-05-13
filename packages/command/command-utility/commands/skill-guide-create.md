---
description: Create a guideline skill from a provided document or URL
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - WebFetch
  - AskUserQuestion
  - TodoWrite
argument-hint: "[source] [--name <name>] [--dry-run]"
---

# /xonovex-utility:skill-guide-create – Create Guideline Skill from Document

Generate a guideline skill with progressive disclosure structure from a document file or URL.

## Goal

- Build any kind of skill: guideline, workflow, capability, processor, or task automation
- Generate spec-compliant SKILL.md with essentials and progressive disclosure
- Create reference files for detail-heavy topics; scripts for bundled executables; assets for templates
- Make content project-independent and reusable

## Spec Constraints

- `name`: 1-64 chars, lowercase kebab-case, no consecutive/leading/trailing hyphens; must match parent dir
- `description`: 1-1024 chars, imperative "Use when..." phrasing, covers what + when (incl. non-obvious triggers)
- Body: target <500 lines / ~5000 tokens; push detail to `references/`
- Optional: `license`, `compatibility` (≤500 chars), `metadata`, `allowed-tools`

## Skill Types

- **Guideline** — coding conventions, style rules, framework patterns
- **Workflow** — multi-step procedures with validation gates
- **Capability** — task-specific automation with bundled scripts
- **Processor** — transforms input to output

## Usage

```bash
# From URL
/xonovex-utility:skill-guide-create "https://example.com/react-best-practices" --name react-guide

# From local file
/xonovex-utility:skill-guide-create "./docs/coding-standards.md" --name typescript-guide

# Preview without writing
/xonovex-utility:skill-guide-create "https://example.com/go-style-guide" --name go-guide --dry-run
```

## Arguments

- `source` (required): URL or file path to the source document
- `--name` (required): Skill name in kebab-case (e.g., `react-guide`, `go-guide`)
- `--dry-run` (optional): Preview generated structure without writing files

## Core Workflow

1. **Identify Source & Type**: URL, file, conversation, or task description; classify the skill type
2. **Fetch Source**: Use WebFetch for URLs, Read for local files; mine conversation/task context
3. **Extract Content**: Requirements, procedures, rules, gotchas, examples, edge cases — whatever the type calls for
4. **Categorize**: Group by topic (or by step, for workflows); decide what lives in SKILL.md vs `references/`
5. **Generate SKILL.md**: Condensed quick reference with essentials (3-7 items); include a Gotchas section for non-obvious env-specific facts
6. **Create Supporting Files**: `references/*.md` for detail (pair each with a load-when trigger), `scripts/` for executables, `assets/` for templates
7. **Validate Structure**: Frontmatter limits met, all reference paths resolve, body under spec ceiling
8. **Write Files**: Save to `<skills-dir>/{name}/` (path depends on the agent harness) or preview with `--dry-run`

## Output Structure

```markdown
---
name: {name}
description: "Use when {task}. Triggers on {patterns}. Skip {adjacent-skill}."
---

# {Title}

## Requirements

- {Optional: tooling/version requirements}

## Essentials

- {Core point 1}
- {Core point 2-6}

## Gotchas

- {Non-obvious env-specific facts}

## Example

{Short representative example}

## Progressive disclosure

- [references/{topic}.md](references/{topic}.md) - Load when {specific trigger condition}
```

## Content Rules

- SKILL.md: 3-7 essentials, one representative example, a Gotchas section
- Reference files: full explanations, multiple examples, rationale; one topic each
- Add only what the agent lacks; omit general knowledge it already has
- Provide a default; mention alternatives briefly (no menus)
- Procedures over declarations (teach the approach, not the one-off answer)
- Match specificity to fragility — prescriptive only when consistency is required
- Mine non-obvious facts/corrections into a **Gotchas** section

## Error Handling

- URL unreachable: `Error: Could not fetch [url]`
- File not found: `Error: File not found at [path]`
- No guidelines found: `Warning: Could not extract guidelines from source`
- Skill exists: Ask to merge or overwrite
- Name not kebab-case: `Error: Skill name must be kebab-case (matches ^[a-z0-9]+(-[a-z0-9]+)*$)`

## Safety

- Preview with `--dry-run` before writing
- Check for existing skill directory
- Preserve existing reference files when merging
- Remove source-specific paths and project names
