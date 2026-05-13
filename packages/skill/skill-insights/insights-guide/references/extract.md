# extract: Capture Development Lessons from a Session

Analyze the current session to identify development mistakes, how they were discovered, and lessons learned. Always save insights as individual files with frontmatter.

## Goal

1. **Analyze** the conversation for development errors and corrections
2. **Identify** general patterns that apply beyond the current task
3. **Extract** category, topic, and applicability from each insight
4. **Save** individual insights to the output directory with frontmatter

## Output Format

```markdown
---
category: testing
topic: json-response-type-safety
applies_to:
  - API testing
  - Hono testing
created: "2026-01-07"
applied: false
---

# [Category]: [Mistake Description]

- **MISTAKE**: [What went wrong]
- **DISCOVERY**: [How the mistake was discovered]
- **FIX**: [How to avoid this mistake]
- **APPLIES TO**: [Types of tasks this affects]
```

## Frontmatter Fields

- `category` — primary technology/domain (e.g. `testing`, `typescript`, `hono`)
- `topic` — specific topic slug (e.g. `json-response-type-safety`)
- `applies_to` — list of contexts where this applies
- `created` — ISO date when insight was created
- `applied` — boolean tracking whether the insight has been integrated into guidelines (always `false` on creation)

The `applies_to` list is used downstream by `integrate-instructions` and `integrate-skills` to determine target files.

## Output

```
Extracted insights from current session

Insights saved:
- insights/testing-json-response-type-safety.md
  Category: testing | Topic: json-response-type-safety
  Applies to: API testing, Hono testing

- insights/typescript-strict-null-checks.md
  Category: typescript | Topic: strict-null-checks
  Applies to: Type safety, Error handling

Total: 2 insights extracted

Next Steps:
1. Review extracted insight files in insights/
2. Verify content: mistakes, discoveries, fixes accurately captured
3. Integrate into skills: invoke integrate-skills for the category
4. Integrate into AGENTS.md: invoke integrate-instructions for the category
```

## Error Handling

- No insights found → warning if no development mistakes / corrections detected in session
- Invalid category → error if specified category doesn't exist or is malformed
- File write error → error if insight file can't be saved to output directory

## Gotchas

- Catching a mistake that was actually correct in context produces noise — only capture corrections you'd want a future agent to remember
- `applies_to` is the routing key downstream; a vague entry like `general` makes integration impossible — be specific
- The `applied: false` flag is the signal that integration is still pending — don't mark `true` until the insight is actually in a skill or AGENTS.md
