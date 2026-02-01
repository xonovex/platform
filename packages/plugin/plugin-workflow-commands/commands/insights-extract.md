---
description: Analyze the session for development mistakes and lessons learned
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
argument-hint: "[category] [--out-dir <dir>]"
---

# /insights-extract – Extract Development Lessons

Analyzes the current session to identify general development mistakes, how they were discovered, and lessons learned. Always saves insights as individual files with frontmatter.

## Usage

```
/insights-extract
/insights-extract tool-usage
/insights-extract validation --out-dir docs/insights
```

## Arguments

- `category` (optional): Focus on a specific mistake category (e.g., `tool-usage`, `dependencies`, `validation`).
- `--out-dir` (optional): The directory to save insight files in. Defaults to `insights/`.

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

The `applies_to` list is used by `/insights-integrate` to determine which guideline files this insight should be added to.

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
3. Integrate: /insights-integrate --all - Transform insights into guideline structure
4. Alternative: /insights-integrate insights/<file>.md - Integrate specific insight
5. Share: Discuss insights with team if applicable before integration
```

## Examples

```bash
# Extract all insights from current session
/insights-extract

# Extract insights for specific category
/insights-extract tool-usage

# Save to custom directory
/insights-extract validation --out-dir docs/insights
```

## Error Handling

- No insights found: Warning if no development mistakes or corrections are detected in session
- Invalid category: Error if specified category does not exist or is malformed
- File write error: Error if insight file cannot be saved to output directory
