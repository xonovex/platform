# extract: Capture Development Lessons from a Session

Analyze the current session for development mistakes, how they were discovered, and the lesson; extract each insight's `category`, `topic`, and `applies_to`; save as individual files with frontmatter.

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

- `category` — primary technology/domain (`testing`, `typescript`, `hono`)
- `topic` — specific topic slug (`json-response-type-safety`)
- `applies_to` — contexts where this applies; the routing key for `integrate-instructions` / `integrate-skills`
- `created` — ISO date
- `applied` — boolean; always `false` on creation, flipped after integration

## Gotchas

- Catching a mistake that was actually correct in context produces noise — only capture corrections you'd want a future agent to remember
- A vague `applies_to` like `general` makes integration impossible — be specific
- Don't mark `applied: true` until the insight is actually in a skill or AGENTS.md
