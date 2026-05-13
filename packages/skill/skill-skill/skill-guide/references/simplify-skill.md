# simplify-skill: Condense a Verbose SKILL.md

## Goal

Reduce SKILL.md by 60-80% (default 70%) while preserving quick reference. Extract examples to reference files. Make content project-independent. Land under spec ceiling (<500 lines / ~5000 tokens body).

## Spec Constraints

- `name`: ≤64 chars, lowercase kebab-case, must match parent dir
- `description`: ≤1024 chars, imperative "Use when..." with explicit trigger contexts; preserve triggering quality when simplifying
- Body: target <500 lines / ~5000 tokens

## Core Workflow

1. Track steps in a task list
2. Read SKILL.md and measure baseline
3. Extract code examples to reference files
4. Simplify SKILL.md to bullet list
5. Remove project-specific references
6. Preview or write changes
7. Report metrics

## Skill Structure

**SKILL.md (condensed quick reference):**

```markdown
---
name: category-guide
description: Use this skill when...
---

# Category Guidelines

## Quick Reference

### Topic Group

- **Rule** - Brief how-to (5-10 words) (references/{topic}.md)
```

**Reference files (examples + explanations):**

```markdown
# topic-name: Topic Title

**Statement:** Clear statement
**Rationale:** Why this exists
**How to Apply:** Steps
**Example:** `code`
**Counter-Example:** `code`
```

## Simplification Rules

**Remove from SKILL.md:** Code blocks, prose paragraphs, long explanations, project paths/names, example sections, duplicate content, anything the agent already knows by general training

**Keep in SKILL.md:** Frontmatter, one-line description, section headings, bullet points with rule + brief how-to + link, small inline code, a **Gotchas** section for non-obvious env-specific facts

**Bullet format:** `- **Rule** - Brief 5-10 word how-to (references/{topic}.md)`

**Reference triggers:** Each reference link must state a load-when condition (e.g., "Load when API returns non-200"), not a generic "see X"

**Create reference files when:** Code examples exist, detailed explanation needed, multiple examples, counter-examples to show

**Defaults over menus:** Pick one default approach, mention alternatives briefly; never list 3+ equal options

**Project-independence:** Remove specific project names/paths/domains. Replace with generic equivalents ("your app", "project root").

## Implementation

**Discovery:** Accept `<skills-dir>/{category}/SKILL.md` or directory path

**Processing:**

1. Read SKILL.md and existing reference files
2. Extract code examples to reference files (group by topic)
3. Remove project-specific references
4. Condense to bullet list format
5. Update reference links

**Validation:** Check references point to existing files, verify each has a load-when trigger, verify reference file structure, confirm body under spec ceiling, skip if <30 lines

## Error Handling

- **File not found:** `Error: SKILL.md not found at [path]`
- **Already minimal:** `Skipping [file]: Already minimal at [N] lines`
- **Invalid target:** `Target reduction must be between 50-90%`
- **Broken references:** reference points to non-existent file

## Safety

- Recommend git commit before running
- Never modify skill `name:` in frontmatter
- Preserve `description:` triggering quality (imperative, ≤1024 chars, keeps trigger contexts)
- Skip skills <30 lines
- Warn if reference files would be overwritten

## Success Metrics

Report: Lines removed ([X]% reduction), reference files created/updated, project references removed, final size vs target
