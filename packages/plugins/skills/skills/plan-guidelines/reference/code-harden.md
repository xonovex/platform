# code-harden: Identify Code Hardening Opportunities

**Guideline:** Analyze code for hardening focusing on type safety, validation, logging, and error handling.

**Rationale:** Hardening improves reliability, debugging, and maintainability. Type safety catches bugs compile-time, validation prevents invalid states, logging aids debugging, and explicit error handling prevents silent failures.

**Example:**

```typescript
// Before: Weak type safety, no validation
function processUser(data: any) {
  return {id: data.id, email: data.email};
}

// After: Strong types, validation, error handling
function processUser(data: unknown): Result<User, ValidationError> {
  const parsed = userSchema.safeParse(data);
  if (!parsed.success) return Err(parsed.error);
  logger.info("Processing user", {id: parsed.data.id});
  return Ok(parsed.data);
}
```

**Techniques:**

- Review project standards by reading CLAUDE.md, AGENTS.md, and style guides
- Scan for type safety issues: any types, implicit types, unchecked assertions
- Identify missing validation: unvalidated inputs, absent schema checks, missing guards
- Analyze error handling patterns: unhandled errors, swallowed exceptions, silent failures
- Detect logging gaps: missing context, inconsistent levels, inadequate debugging info
- Find code smells: long functions (>30 lines), deep nesting, complex branches
- Group issues by severity and category, prioritize by impact and fix effort
