---
description: "Analyze the current session for development mistakes, surprises, and lessons learned. Use when the user asks to extract insights, capture lessons, review what went wrong, or run a post-session reflection. Keywords: insights, retrospective, lessons learned, mistakes, post-mortem, reflection, session review."
---

# /xonovex-utility:insights-extract – Extract Development Lessons

Analyzes the current session to identify general development mistakes, how they were discovered, and lessons learned. Always saves insights as individual files with frontmatter.

## Goal

1.  **Analyzes** the conversation for development errors and corrections.
2.  **Identifies** general patterns that apply beyond the current task.
3.  **Extracts** category, topic, and applicability from the insight.
4.  **Saves** individual insights to the output directory with frontmatter.

## Output Format

Creates files with this structure:

```markdown
---
category: testing
topic: json-response-type-safety
applies_to:
  - API testing
  - Hono testing
created: "2025-01-07"
applied: false
---

# [Category]: [Mistake Description]

- **MISTAKE**: [What went wrong]
- **DISCOVERY**: [How the mistake was discovered]
- **FIX**: [How to avoid this mistake]
- **APPLIES TO**: [Types of tasks this affects]
```

## Frontmatter Fields

- `category`: Primary technology/domain (e.g., `testing`, `typescript`, `hono`)
- `topic`: Specific topic slug (e.g., `json-response-type-safety`)
- `applies_to`: List of contexts where this applies (e.g., `["API testing", "Hono testing"]`)
- `created`: ISO date when insight was created
- `applied`: Boolean tracking if insight has been integrated into guidelines (always `false` on creation)

The `applies_to` list is used by `/xonovex-utility:insights-skills-integrate` and `/xonovex-utility:insights-instructions-integrate` to determine target files.

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
1. Review extracted insight files in insights/ directory
2. Verify content: Check that mistakes, discoveries, and fixes are accurately captured
3. Integrate into skills: invoke insights-skills-integrate for the category
4. Integrate into AGENTS.md: invoke insights-instructions-integrate for the category
```

## Error Handling

- No insights found: Warning if no development mistakes or corrections are detected in session
- Invalid category: Error if specified category does not exist or is malformed
- File write error: Error if insight file cannot be saved to output directory
