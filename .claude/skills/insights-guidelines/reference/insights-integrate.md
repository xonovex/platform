# insights-integrate: Convert Insights to Skills

**Guideline:** Integrate extracted insights into Claude skills with progressive disclosure structure, creating or merging into category-specific guidelines.

**Rationale:** Transforms session insights into permanent, discoverable guidelines organized by topic in detail files, enabling knowledge reuse without cluttering main skill documentation.

**Example:**

```markdown
---
name: testing-guidelines
description: Use when writing tests. Apply for type safety, mocking, async patterns.
---

## Essentials

- Validate all external inputs
- Type-check response data with Zod schemas
- Mock external dependencies deterministically
- Test error paths explicitly
- Use test factories for complex data

## Progressive Disclosure

- Read [details/json-response-validation.md](details/json-response-validation.md) - When validating API responses
```

**Techniques:**
- Search insights/ directory for category-matching insight files
- Parse content and extract topics with MISTAKE/DISCOVERY/FIX sections
- Generate or merge SKILL.md: Essentials (3-7 items), Progressive Disclosure links
- Create detail files from each insight using standard Guideline/Rationale/Example/Techniques format
- Deduplicate essentials keeping most important items
- Preserve existing detail files when merging
- Mark integrated insights as `applied: true`
