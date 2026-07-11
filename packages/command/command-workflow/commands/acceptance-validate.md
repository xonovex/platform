---
description: "Execution: validate a PR against its acceptance criteria, reporting pass/fail per scenario and flagging gaps"
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
  - Skill
argument-hint: "[feature-or-feature-dir] [--pr <ref>]"
---

# /xonovex-workflow:acceptance-validate — Validate a PR against acceptance criteria

## Arguments

- `[feature-or-feature-dir]` (optional): Feature name or path to `.feature` files (default `acceptance/<feature>/`)
- `[--pr <ref>]` (optional): PR / branch / diff to validate (default current branch)

## Delegation

Load the `bdd-guide` skill (plugin `xonovex-skill-bdd`), `code-review-guide` (structured,
blocking/non-blocking findings), and `testing-guide` (coverage assessment), and follow
them to check the implementation against each acceptance scenario: for every
Given-When-Then, find the covering test and verify the code path, boundary, and error
handling; mark each PASS / PARTIAL / FAIL with evidence; flag uncovered scenarios,
criteria changed without sign-off, and regressions; then give a verdict (accept /
accept-with-notes / reject). The skills are the source of truth for the method — do not
restate them.

Write the report to `acceptance/<feature>/validation-report.md`. If a work-item-tracker
host skill is installed, additionally post the results and transition the item; otherwise
write to disk only and note which host skill enables sync.
