# plan-research-code-harden: Research Code Hardening Opportunities

Analyze code for hardening opportunities (type safety, validation, logging, error handling). Generates a research report and stops there. Does **not** edit code and does **not** create plans — continue with the plan operations afterward (optionally `plan-clarify` to settle open decisions, then `plan-create`).

## Goal

- Identify code quality issues (type safety, logging, validation, best practices, code smells)
- Assess each finding against project standards and conventions
- Produce a prioritized research report to feed into `plan-create`

## Core Workflow

**Delegate codebase analysis to read-only search agents where available; otherwise use grep/find/file-read directly. Stay in research mode.**

1. **Read guidelines** — check AGENTS.md, POLICY.md and referenced guidelines for project standards
2. **Analyze** — run focused, read-only searches to find anti-patterns and violations; categorize by priority
3. **Report** — generate a detailed report grouped by package and priority

## Implementation Details

**Find guidelines:** AGENTS.md and POLICY.md in project root and subdirectories; check @-referenced documents

**Assess against standards:** compare findings to project-specific patterns for type safety, logging, validation, error handling

**Scope:** report findings one package at a time so the downstream plan can be sequenced cleanly

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

- **No matching files:** report that the path / aspect produced no candidates rather than failing silently
- **Conflicting standards:** when guidelines disagree, surface both in the report instead of picking one
- **Guidelines not found:** fall back to language / framework best practices and note the assumption in the report

## Gotchas

- Hardening without reading project guidelines first applies generic best-practices that may conflict with project conventions — read AGENTS.md / POLICY.md first
- A pile of `any` types is often a symptom of a missing schema, not a typing problem — fix the boundary, not every site
- Adding logging everywhere creates noise — log at boundaries and on error paths, not every function entry
- Validation at every layer is wasteful — validate at trust boundaries (user input, external APIs); trust internal callers
