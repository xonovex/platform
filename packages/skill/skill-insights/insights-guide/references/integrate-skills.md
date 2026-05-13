# integrate-skills: Convert Insights into a Guideline Skill

Convert insights from a category into a new or updated guideline skill with progressive-disclosure structure.

## Workflow

1. **Discover** — search `insights/` for category files, extract Problem/Solution/Example, group by topic
2. **Generate Structure**
   - Metadata: `{category}-best-practices`, description starting with "Use when..."
   - SKILL.md: Requirements (optional), Essentials (3-7 items), Examples (code), Progressive disclosure (reference links with load-when triggers)
   - Reference files: Guideline, Rationale, How to Apply, Example (bad vs good), Related
3. **Merge** — if skill exists (unless user asked to overwrite): combine metadata, deduplicate Essentials (keep 3-7), append examples, add reference-file links
4. **Output** — preview shows structure without writing; otherwise creates directory and files

## Structure Template

**SKILL.md:**

````markdown
---
name: {category}-best-practices
description: "Use when working with {category} to {purpose}. Apply for {scenarios}."
---

## Essentials

- {Core guideline 1}
- {Core guideline 2-6}

## Examples

```typescript
{Code showing best practice}
```

## Progressive disclosure

- **references/{topic}.md** - Load when {scenario}
````

**Reference file (`references/{topic}.md`):**

````markdown
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
```

**Related:** references/{other-topic}.md
````

## Example Output

```
[OK] Skill created: <skills-dir>/api/SKILL.md

- 12 insights converted, 4 topics
- Created references/status-codes.md, references/validation-safety.md
```

## Error Handling

- Missing category → ask user
- No insights found → suggest running `extract` for the category
- Output not writable → report error

## Gotchas

- Insights that are one-off corrections don't deserve a whole skill — fold them into AGENTS.md via `integrate-instructions` instead
- Generating a skill with only 1-2 essentials produces noise — wait until the category has enough insights to fill 3-7 essential bullets
- Each reference link must have an explicit load-when trigger; a bare `see references/x.md` defeats progressive disclosure
