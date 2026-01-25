# extract: Extract Skill from Codebase Patterns

**Guideline:** Create or update skill by analyzing codebase patterns, project instructions, and source files for reuse.

**Rationale:** Codebases contain established patterns that ensure consistency across development and preserve institutional knowledge.

**Example:**

```bash
# Analyze services/api project
# Find AGENTS.md: "Express 5+ API servers, route organization, error handling"
# Glob *.ts files in api/ → Find express imports, error handlers, middleware

# Sample patterns from codebase:
# 1. Routes: All use /api/v1/<resource> pattern
# 2. Errors: Custom AppError class with status codes
# 3. Validation: All routes use Zod schemas
# 4. Testing: Jest with supertest for HTTP assertions

# Create SKILL.md with 4 essentials + Express example
# Details:
# - details/routing.md (routes, versioning, organization)
# - details/error-handling.md (AppError, error middleware)
# - details/validation.md (Zod schemas, SafeParseResult patterns)
# - details/testing.md (Jest, supertest, mocking)

# Result: express-api-patterns skill extracted from codebase
```

**Techniques:**
- Find AGENTS.md and CLAUDE.md in source path for project guidelines
- Glob source files by extension to identify language and file patterns
- Extract patterns from project docs: section headers, tables, code blocks, architecture
- Sample source files for common patterns: naming, types, function signatures
- Detect language: .ts/.tsx → TypeScript, .py → Python, .c/.h → C99, .lua → Lua
- Categorize patterns: architecture, types, testing, safety, naming conventions
- Create SKILL.md with 3-5 key essentials, one code example, detail references
- Create details/{pattern}.md for each category with guideline, rationale, examples
- Skip patterns already in base skill to avoid duplication
- Use kebab-case for skill and detail file names
- Remove project-specific paths and names for reusability

## When to Apply

- Extracting patterns from project codebase
- Creating skill from established practices
- Documenting institutional knowledge
- Formalizing coding conventions
