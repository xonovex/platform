---
description: Convert insights from a category into a progressive disclosure skill
model: sonnet
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - AskUserQuestion
argument-hint: "[category] [--dry-run] [--force] [--output <path>]"
---

# /insights-integrate – Convert Insights to Skill

Convert insights from a category into a Claude skill with progressive disclosure structure.

## Arguments

- `category` (required) - Category to convert (e.g., `hono`, `typescript`, `workflow`)
- `--dry-run` - Preview without writing
- `--force` - Overwrite existing skill instead of merging
- `--output <path>` - Custom output path (default: `.claude/skills/{category}/SKILL.md`)

## Workflow

1. **Discover:** Search `insights/` for category files, extract Problem/Solution/Example, group by topic
2. **Generate Structure:**
   - Metadata: `{category}-best-practices`, description under 150 chars starting with "Use when..."
   - SKILL.md: Requirements (optional), Essentials (3-7 items), Examples (code), Progressive disclosure (detail refs)
   - Detail files: Guideline, Rationale, How to Apply, Example (bad vs good), Related
3. **Merge:** If skill exists and not `--force`: combine metadata, deduplicate Essentials (keep 3-7), append examples, add detail file refs
4. **Output:** `--dry-run` shows structure without writing, otherwise creates directory and files

## Structure Template

**SKILL.md:**

````markdown
---
name: {category}-best-practices
description: Use when working with {category} to {purpose}. Apply for {scenarios}.
---

## Essentials

- {Core guideline 1}
- {Core guideline 2-6}

## Examples

```typescript
{Code showing best practice}
```
````

## Progressive disclosure

- **details/{topic}.md** - When {scenario}

````

**Detail file (details/{topic}.md):**
```markdown
# {topic}: {Title}

**Guideline:** {Rule statement}
**Rationale:** {Why this matters}

**How to Apply:**
1. {Step-by-step}

**Example:**
```typescript
// Bad
{Anti-pattern}
// Good
{Correct usage}
````

**Related:** details/{other-topic}.md

```

## Example Output

```

[OK] Skill created: .claude/skills/api/SKILL.md

- 12 insights converted, 4 topics
- Created details/status-codes.md, details/validation-safety.md

```

## Error Handling

- Missing category: ask user
- No insights found: suggest `/insights-extract [category]`
- Output not writable: report error
```
