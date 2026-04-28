# simplify: Condense Verbose Skill to Bullet Format

**Guideline:** Reduce SKILL.md by 60-80% to bullet format; move code to reference files; remove project-specific references.

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

- **Error handling** - Define custom AppError class, see [reference/error-handling.md]
- **Validation** - Use Zod schemas for request payload validation, see [reference/validation.md]
- **Testing** - Use Jest + supertest for HTTP assertions, see [reference/testing.md]

# Code examples moved to reference/ files

# reference/error-handling.md - Full AppError implementation

# reference/validation.md - Zod schema patterns

# reference/testing.md - Jest + supertest examples

# Reduction: 80% of original size, all content preserved in reference files
```

**Techniques:**

- Read SKILL.md and existing reference files completely
- Extract code examples to reference files, grouped by topic
- Remove project-specific paths, names, domains, and URLs
- Condense prose explanations to bullet format: `- **Rule** - How-to, see [reference/file.md]`
- Remove long paragraphs, code blocks, example sections from main file
- Keep frontmatter, description, section headings, and short inline snippets
- Create or update reference files for code examples and detailed explanations
- Validate all reference file links are correct
- Target 60-70% reduction for standard simplification
- Preserve all content by moving to reference files, don't delete
- Check for reference file overwrites before writing
