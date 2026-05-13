# code-harden: Research Code Hardening Opportunities

Analyze code for hardening opportunities (type safety, validation, logging, error handling). Generates a research report. Does **not** create plans or make changes — run `plan-create` afterward.

## Goal

- Identify code quality issues (type safety, logging, validation, best practices, code smells)
- Apply fixes aligned with project standards
- Validate changes with typecheck / lint / tests

## Core Workflow

**Delegate codebase analysis to read-only search agents where available; otherwise use grep/find/file-read directly. Stay in research mode.**

1. **Read guidelines** — check AGENTS.md, POLICY.md and referenced guidelines for project standards
2. **Analyze** — run focused, read-only searches to find anti-patterns and violations; categorize by priority
3. **Report** — generate a detailed report grouped by package and priority

## Implementation Details

**Find guidelines:** AGENTS.md and POLICY.md in project root and subdirectories; check @-referenced documents

**Apply standards:** follow project-specific patterns for type safety, logging, validation, error handling

**Validation:** fix one package at a time; validate immediately after each

## Example Transformation

```typescript
// Before: weak types, no validation
function processUser(data: any) {
  return {id: data.id, email: data.email};
}

// After: strong types, validation, error handling
function processUser(data: unknown): Result<User, ValidationError> {
  const parsed = userSchema.safeParse(data);
  if (!parsed.success) return Err(parsed.error);
  logger.info("Processing user", {id: parsed.data.id});
  return Ok(parsed.data);
}
```

## What to Look For

- **Type safety:** `any` types, implicit types, unchecked assertions
- **Missing validation:** unvalidated inputs, absent schema checks, missing guards
- **Error handling:** unhandled errors, swallowed exceptions, silent failures
- **Logging:** missing context, inconsistent levels, inadequate debugging info
- **Code smells:** long functions (>30 lines), deep nesting, complex branches

Group by severity + category; prioritize by impact and fix effort.

## Error Handling

- **Lint failures:** review against project rules
- **Test failures:** review logic, validation strictness, mock compatibility
- **Type errors:** check imports, type definitions, schema alignment
- **Guidelines not found:** fall back to language / framework best practices

## Gotchas

- Hardening without reading project guidelines first applies generic best-practices that may conflict with project conventions — read AGENTS.md / POLICY.md first
- A pile of `any` types is often a symptom of a missing schema, not a typing problem — fix the boundary, not every site
- Adding logging everywhere creates noise — log at boundaries and on error paths, not every function entry
- Validation at every layer is wasteful — validate at trust boundaries (user input, external APIs); trust internal callers
