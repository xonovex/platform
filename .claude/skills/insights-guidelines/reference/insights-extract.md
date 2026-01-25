# insights-extract: Extract Development Lessons

**Guideline:** Extract development mistakes and lessons learned from sessions into structured insight files with frontmatter metadata.

**Rationale:** Captures learning from errors during development. Structured insights with frontmatter can be tracked and later integrated into permanent guidelines to improve team practices.

**Example:**

```markdown
---
category: testing
topic: json-response-type-safety
applies_to: [API testing, Hono testing]
created: "2025-01-07"
applied: false
---

# Testing: JSON Response Type Safety

- **MISTAKE**: Expected JSON response parsing to provide type safety automatically
- **DISCOVERY**: Tests passed but runtime errors occurred with incorrect property access
- **FIX**: Use Zod schema validation on response.json() results
- **APPLIES TO**: API testing, Response validation
```

**Techniques:**
- Review session for development errors and corrections made
- Extract patterns applicable beyond current task
- Create insight file with frontmatter: category, topic, applies_to, created, applied
- Document: MISTAKE, DISCOVERY, FIX, APPLIES TO sections
- Categorize by technology and domain for later discovery
