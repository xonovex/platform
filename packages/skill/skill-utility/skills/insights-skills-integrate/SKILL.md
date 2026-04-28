---
description: "Convert extracted insights into a new or updated guideline skill with progressive disclosure. Use when the user asks to turn insights into a skill, generate a SKILL.md from lessons, or build a reusable guideline from session findings. Keywords: insights, skill creation, SKILL.md, progressive disclosure, guideline, lessons to skill."
---

# /xonovex-utility:insights-skills-integrate – Convert Insights to Skill

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
   - SKILL.md: Requirements (optional), Essentials (3-7 items), Examples (code), Progressive disclosure (reference links)
   - Reference files: Guideline, Rationale, How to Apply, Example (bad vs good), Related
3. **Merge:** If skill exists and not `--force`: combine metadata, deduplicate Essentials (keep 3-7), append examples, add reference file links
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

- **reference/{topic}.md** - When {scenario}

````

**Reference file (reference/{topic}.md):**
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

**Related:** reference/{other-topic}.md

```

## Example Output

```

[OK] Skill created: .claude/skills/api/SKILL.md

- 12 insights converted, 4 topics
- Created reference/status-codes.md, reference/validation-safety.md

```

## Error Handling

- Missing category: ask user
- No insights found: suggest `/xonovex-utility:insights-extract [category]`
- Output not writable: report error
```
