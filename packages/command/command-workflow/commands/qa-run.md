---
description: Validate an exact Deliverable Publication revision across recorded environments and publish independent QA evidence
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
  - AskUserQuestion
  - Skill
argument-hint: "<deliverable-reference> [--revision <native-revision>] [--scope <selection>] [--environment <reference>] [--provider <selection>] [--max-concurrency <count>]"
---

# /xonovex-workflow:qa-run — Run Quality Assurance

## Arguments

- `deliverable-reference` (required): Opaque Deliverable Publication reference.
- `--revision` (optional): Exact immutable candidate revision.
- `--scope` (repeatable): Functional, integration, security, accessibility, performance, AI, supply-chain, or domain test scope.
- `--environment` (repeatable): Required environment reference/version.
- `--provider`, `--max-concurrency` (optional): Result provider and suite concurrency bound.

## Delegation

Load `workflow-guide` and perform **qa-run**. Load `testing-guide` plus selected domain,
scanner, CI, and provider skills. Preserve each suite's environment and native evidence;
publish QA separately and never coerce skipped, stale, partial, or timed-out work to pass.
