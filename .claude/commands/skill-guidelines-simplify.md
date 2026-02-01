---
description: >-
  Make skills project-independent, remove redundancy, condense SKILL.md to
  bullet list with examples in detail files
model: sonnet
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - TodoWrite
argument-hint: "[skill-file] [--dry-run] [--target-reduction <percent>]"
---

# /skill-guidelines-simplify – Condense verbose skill files

## Goal

Reduce SKILL.md by 60-80% (default 70%) while preserving quick reference. Extract examples to detail files. Make content project-independent.

## Arguments

- `[skill-file]` (required) - Path to SKILL.md file or skill directory
- `[--dry-run]` (optional) - Preview without modifying
- `[--target-reduction <percent>]` (optional) - Override default 70% (range: 50-90)

## Core Workflow

1. Use TodoWrite to track steps
2. Read SKILL.md and measure baseline
3. Extract code examples to detail files
4. Simplify SKILL.md to bullet list
5. Remove project-specific references
6. Preview or write changes
7. Report metrics

## Skill Structure

**SKILL.md (condensed quick reference):**

```markdown
---
name: category-guidelines
description: Use this skill when...
claude:allowed-tools: ["Read", "Edit"]
---

# Category Guidelines

## Quick Reference

### Topic Group

- **Rule** - Brief how-to (5-10 words) (@details/detail-file.md)
```

**Detail files (examples + explanations):**

```markdown
# topic-name: Topic Title

**Guideline:** Clear statement
**Rationale:** Why this exists
**How to Apply:** Steps
**Example:** `code`
**Counter-Example:** `code`
```

## Simplification Rules

**Remove from SKILL.md:** Code blocks, prose paragraphs, long explanations, project paths/names, example sections, duplicate content

**Keep in SKILL.md:** Frontmatter, one-line description, section headings, bullet points with rule + brief how-to + @link, small inline code

**Bullet format:** `- **Rule** - Brief 5-10 word how-to (@details/detail-file.md)`

**Create detail files when:** Code examples exist, detailed explanation needed, multiple examples, counter-examples to show

**Project-independence:** Remove specific project names/paths/domains. Replace with generic equivalents ("your app", "project root").

## Implementation

**Discovery:** Accept `.claude/skills/{category}/SKILL.md` or directory path

**Processing:**

1. Read SKILL.md and existing detail files
2. Extract code examples to detail files (group by topic)
3. Remove project-specific references
4. Condense to bullet list format
5. Update @references

**Validation:** Check @references point to existing files, verify detail file structure, skip if <30 lines

## Examples

**Before (verbose):**

````markdown
## Performance

Use memoization for expensive calculations:

```typescript
const processed = useMemo(() => processData(data), [data]);
```
````

This prevents recalculations on every render.

````

**After (condensed):**
```markdown
### Performance
- **Memoize expensive calculations** - Use `useMemo` for costly operations (@details/performance.md)
````

**Detail file created (details/performance.md):**

````markdown
# performance: Performance Optimization

**Guideline:** Use memoization hooks for expensive calculations
**Example:** ```typescript
const processed = useMemo(() => processData(data), [data]);
````

```

## Error Handling

- **File not found:** `Error: SKILL.md not found at [path]`
- **Already minimal:** `Skipping [file]: Already minimal at [N] lines`
- **Invalid target:** `Target reduction must be between 50-90%`
- **Broken references:** `@reference points to non-existent file`

## Safety

- Recommend git commit before running
- Never modify skill `name:` in frontmatter
- Skip skills <30 lines
- Warn if detail files would be overwritten

## Success Metrics

Report: Lines removed ([X]% reduction), detail files created/updated, project references removed, final size vs target
```
