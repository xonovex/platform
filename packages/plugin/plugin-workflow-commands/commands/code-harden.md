---
description: "Harden code by improving type safety, validation, logging, and error handling"
model: sonnet
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
argument-hint: "[path] [--aspects <type-safety,logging,validation>] [--auto-fix] [--dry-run]"
---

# /code-harden – Research Code Hardening Opportunities

Analyzes code for hardening opportunities (type safety, validation, logging, error handling). Generates a detailed research report. Does NOT create plans or make changes - run `/plan-create` afterward to create an implementation plan.

## Goal

- Identify code quality issues (type safety, logging, validation, best practices, code smells)
- Apply fixes aligned with project standards
- Validate changes with typecheck, lint, and tests

## Usage

```bash
# Analyze all aspects
/code-harden packages/myapp/

# Focus on specific aspects
/code-harden src/ --aspects type-safety,logging

# Preview analysis
/code-harden . --dry-run
```

## Arguments

- `path` (required): Directory to analyze
- `--aspects` (optional): Comma-separated aspects (type-safety, logging, validation, error-handling, testing, or custom)
- `--auto-fix` (optional): Automatically apply safe fixes
- `--dry-run` (optional): Report issues without making changes

## Core Workflow

**Use Task agents with subagent_type=Explore and model=haiku for codebase analysis. Do NOT use EnterPlanMode.**

1. **Read Guidelines**: Check CLAUDE.md, POLICY.md, AGENTS.md and referenced guidelines for project standards
2. **Analyze**: Use Task agent with subagent_type=Explore and model=haiku to find anti-patterns and violations; categorize by priority
3. **Report**: Generate detailed report of issues found, grouped by package and priority

## Implementation Details

**Find guidelines**: Look for CLAUDE.md, POLICY.md, AGENTS.md in project root and subdirectories; check @-referenced documents

**Apply standards**: Follow project-specific patterns from guidelines for type safety, logging, validation, error handling

**Validation**: Fix one package at a time; validate immediately after each

## Error Handling

- **Lint failures**: Review fix against project linting rules; adjust to match standards
- **Test failures**: Review logic errors, validation strictness, mock compatibility
- **Type errors**: Check imports, type definitions, schema alignment
- **Guidelines not found**: Search for coding standards documentation; use language/framework best practices

## Next Steps

After running this research command:

1. Review the hardening report for accuracy
2. Run `/plan-create` to create an implementation plan from this research
