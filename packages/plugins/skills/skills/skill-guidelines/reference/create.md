# create: Create Guideline Skill from Document

**Guideline:** Generate guideline skill with progressive disclosure structure from document file or URL.

**Rationale:** External sources (documentation, guides, articles) contain valuable patterns. Converting to skill format makes guidelines accessible during development.

**Example:**

```bash
# Source: https://example.com/flask-patterns-guide
# Extract sections: Error Handling, Type Hints, Testing
# Code examples: JSON validation, decorators, pytest patterns

# SKILL.md created with essentials (3-5 bullets + 1 example)
- **Validation** - Use Marshmallow for JSON schema validation
- **Error handling** - Define custom exceptions for API errors
- **Testing** - Use pytest fixtures for database mocking
- **Type hints** - Add type hints to all functions

# Code example showing validation + error handling pattern

# Details extracted:
# details/validation.md    - Marshmallow schemas, custom fields
# details/error-handling.md - Custom exceptions, error responses
# details/testing.md       - Pytest fixtures, mocking patterns

# Result: Skill accessible in .claude/skills/flask-patterns/
```

**Techniques:**

- Fetch source: WebFetch for URLs, Read for local files
- Extract guidelines, requirements, rules, patterns, and code examples
- Parse headings as topic groups, code blocks for language context
- Identify patterns: do/don't, good/bad, prefer/avoid statements
- Categorize by topic: performance, architecture, testing, security, etc.
- Create SKILL.md with 3-7 essential bullets and one representative example
- Extract detailed explanations and code to details/{topic}.md files
- Use bullet format: `- **Rule** - Brief how-to, see [details/file.md](details/file.md)`
- Validate all references point to existing detail files
- Remove source-specific paths, project names, and domain references
- Save to .claude/skills/{name}/ with kebab-case naming
