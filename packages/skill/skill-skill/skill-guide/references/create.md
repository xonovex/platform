# create: Create Skill from Source

## Goal

- Build any kind of skill: guideline, workflow, capability, processor, or task automation
- Generate spec-compliant SKILL.md with essentials and progressive disclosure
- Create reference files for detail-heavy topics; scripts for bundled executables; assets for templates
- Make content project-independent and reusable

## Spec Constraints

Spec limits (name / description / body / optional-field rules) live in the parent SKILL.md — re-check after edits.

## Skill Types

- **Guideline** — coding conventions, style rules, framework patterns
- **Workflow** — multi-step procedures with validation gates
- **Capability** — task-specific automation with bundled scripts
- **Processor** — transforms input to output

## Core Workflow

1. **Identify Source & Type**: URL, file, conversation, or task description; classify the skill type (guideline skills: see [guideline-skills.md](guideline-skills.md))
2. **Fetch Source**: Fetch URLs from the web; read local files; mine conversation/task context
3. **Extract Content**: Requirements, procedures, rules, gotchas, examples, edge cases — whatever the type calls for
4. **Draft evals first**: Write 2-3 trigger/output eval cases for the gap _before_ authoring extensive body content — see [evaluating-triggers.md](evaluating-triggers.md), [evaluating-outputs.md](evaluating-outputs.md)
5. **Categorize**: Group by topic (or by step, for workflows); decide what lives in SKILL.md vs `references/`
6. **Generate SKILL.md**: Condensed quick reference with essentials (3-7 items) and a routing-first description (see [writing-descriptions.md](writing-descriptions.md)); include a Gotchas section for non-obvious env-specific facts
7. **Create Supporting Files**: `references/*.md` for detail (pair each with a load-when trigger), `scripts/` for executables (see [using-scripts.md](using-scripts.md)), `assets/` for templates
8. **Validate Structure**: Frontmatter limits met, all reference paths resolve, body under spec ceiling
9. **Write Files**: Save to `<skills-dir>/{name}/` (path depends on the agent harness) or preview without writing

## Output Structure

```markdown
---
name: {name}
description: "Use when {task}. Triggers on {patterns}."
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
- Reference files: full explanations, multiple examples, rationale; one topic each — don't restate when/why to read the file itself (its load-when trigger lives in the SKILL.md progressive-disclosure list); pointing to other references with a "read when" trigger is fine
- Apply the parent SKILL.md Core Principles (add what the agent lacks, defaults over menus, procedures over declarations)
- Match specificity to fragility — prescriptive only when consistency is required
- Mine non-obvious facts/corrections into a **Gotchas** section

## Error Handling

- URL unreachable: `Error: Could not fetch [url]`
- File not found: `Error: File not found at [path]`
- No usable content: `Warning: Could not extract content from source`
- Skill exists: Ask to merge (see [merge.md](merge.md)) or overwrite
- Name not kebab-case: `Error: Skill name must be kebab-case (matches ^[a-z0-9]+(-[a-z0-9]+)*$)`

## Safety

- Preview before writing
- Check for existing skill directory
- Preserve existing reference files when merging
- Remove source-specific paths and project names
