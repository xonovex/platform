---
description: >-
  Research code-hardening opportunities (type safety, validation, logging,
  error handling) — read-only report that feeds a follow-up plan; makes no
  changes
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
argument-hint: "[path] [--aspects <type-safety,logging,validation>]"
---

# /xonovex-workflow:plan-research-code-harden – Research Code Hardening Opportunities

Analyzes code for hardening opportunities (type safety, validation, logging, error handling). Generates a detailed research report. Does NOT create plans or make changes - run `/xonovex-workflow:plan-create` afterward to create an implementation plan.

## Goal

- Identify code quality issues (type safety, logging, validation, best practices, code smells)
- Assess each finding against project standards and conventions
- Produce a prioritized research report to feed into `/xonovex-workflow:plan-create`

## Usage

```bash
# Analyze all aspects
/xonovex-workflow:plan-research-code-harden packages/myapp/

# Focus on specific aspects
/xonovex-workflow:plan-research-code-harden src/ --aspects type-safety,logging
```

## Arguments

- `path` (required): Directory to analyze
- `--aspects` (optional): Comma-separated aspects (type-safety, logging, validation, error-handling, testing, or custom)

## Core Workflow

**IMPORTANT: This command is read-only research, do NOT edit code and do NOT author a plan. The output is a report; planning happens afterward via the plan commands. Delegate codebase analysis to read-only search agents where available; otherwise use grep/find/file-read directly.**

1. **Read Guidelines**: Check AGENTS.md, POLICY.md and referenced guidelines for project standards
2. **Analyze**: Run focused, read-only searches to find anti-patterns and violations; categorize by priority
3. **Report**: Generate detailed report of issues found, grouped by package and priority

## Implementation Details

**Find guidelines**: Look for AGENTS.md and POLICY.md in project root and subdirectories; check @-referenced documents

**Assess against standards**: Compare findings to project-specific patterns from guidelines for type safety, logging, validation, error handling

**Scope**: Report findings one package at a time so the downstream plan can be sequenced cleanly

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

- **No matching files**: Report that the path / aspect produced no candidates rather than failing silently
- **Conflicting standards**: When guidelines disagree, surface both in the report instead of picking one
- **Guidelines not found**: Fall back to language / framework best practices and note the assumption in the report

## Gotchas

- Hardening without reading project guidelines first applies generic best-practices that may conflict with project conventions — read AGENTS.md / POLICY.md first
- A pile of `any` types is often a symptom of a missing schema, not a typing problem — fix the boundary, not every site
- Adding logging everywhere creates noise — log at boundaries and on error paths, not every function entry
- Validation at every layer is wasteful — validate at trust boundaries (user input, external APIs); trust internal callers

## Next Steps

After running this research command:

1. Review the hardening report for accuracy
2. Optionally run `/xonovex-workflow:plan-clarify` to settle open decisions one by one
3. Run `/xonovex-workflow:plan-create` to create an implementation plan from this research
