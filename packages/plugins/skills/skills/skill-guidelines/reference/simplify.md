# simplify: Condense Verbose Skill to Bullet Format

**Guideline:** Reduce SKILL.md by 60-80% to bullet format; move code to details; remove project-specific references.

**Rationale:** Verbose skills become hard to scan. Progressive disclosure loads only what's needed during development.

**Example:**

```markdown
# BEFORE (SKILL.md, verbose 45 lines)

## Error Handling

When working with Express applications, error handling is crucial for providing meaningful responses to clients. You should create custom error classes that extend the built-in Error class to encapsulate domain-specific errors...

[large code example here]

## Validation

Use Zod for runtime validation of request payloads. Zod provides type-safe schema validation...

[large code example here]

# AFTER (SKILL.md simplified to 8 lines)

- **Error handling** - Define custom AppError class, see [details/error-handling.md]
- **Validation** - Use Zod schemas for request payload validation, see [details/validation.md]
- **Testing** - Use Jest + supertest for HTTP assertions, see [details/testing.md]

# Code examples moved to details/ files

# details/error-handling.md - Full AppError implementation

# details/validation.md - Zod schema patterns

# details/testing.md - Jest + supertest examples

# Reduction: 80% of original size, all content preserved in details
```

**Techniques:**

- Read SKILL.md and existing detail files completely
- Extract code examples to detail files, grouped by topic
- Remove project-specific paths, names, domains, and URLs
- Condense prose explanations to bullet format: `- **Rule** - How-to, see [details/file.md]`
- Remove long paragraphs, code blocks, example sections from main file
- Keep frontmatter, description, section headings, and short inline snippets
- Create or update detail files for code examples and detailed explanations
- Validate all detail file references are correct
- Target 60-70% reduction for standard simplification
- Preserve all content by moving to details, don't delete
- Check for detail file overwrites before writing

## When to Apply

- SKILL.md over 30 lines with verbose explanations
- Many code examples clutter the main file
- Project-specific details need removal for reusability
- Preparing skill for general distribution
