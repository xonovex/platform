---
description: "Condense verbose SKILL.md files into bullet-list quick references and move examples to reference files. Use when the user asks to simplify, slim down, or de-project-ify a skill. Keywords: SKILL.md, skill simplify, condense, project-independent, progressive disclosure, reference files."
---

# /xonovex-utility:skill-guidelines-simplify – Condense verbose skill files

## Goal

Reduce SKILL.md by 60-80% (default 70%) while preserving quick reference. Extract examples to reference files. Make content project-independent.

## Core Workflow

1. Use TodoWrite to track steps
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
name: category-guidelines
description: Use this skill when...
claude:allowed-tools: ["Read", "Edit"]
---

# Category Guidelines

## Quick Reference

### Topic Group

- **Rule** - Brief how-to (5-10 words) (@reference/{topic}.md)
```

**Reference files (examples + explanations):**

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

**Bullet format:** `- **Rule** - Brief 5-10 word how-to (@reference/{topic}.md)`

**Create reference files when:** Code examples exist, detailed explanation needed, multiple examples, counter-examples to show

**Project-independence:** Remove specific project names/paths/domains. Replace with generic equivalents ("your app", "project root").

## Implementation

**Discovery:** Accept `.claude/skills/{category}/SKILL.md` or directory path

**Processing:**

1. Read SKILL.md and existing reference files
2. Extract code examples to reference files (group by topic)
3. Remove project-specific references
4. Condense to bullet list format
5. Update @references

**Validation:** Check @references point to existing files, verify reference file structure, skip if <30 lines

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
- **Memoize expensive calculations** - Use `useMemo` for costly operations (@reference/performance.md)
````

**Reference file created (reference/performance.md):**

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
- Warn if reference files would be overwritten

## Success Metrics

Report: Lines removed ([X]% reduction), reference files created/updated, project references removed, final size vs target
```
