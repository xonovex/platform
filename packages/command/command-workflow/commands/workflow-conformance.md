---
description: Validate an executable workflow composition
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Skill
argument-hint: "<invocation.json> --registry <registry.json>"
---

# /xonovex-workflow:workflow-conformance — Validate a Workflow Contract

## Arguments

- `invocation.json` (required): Runtime invocation to validate.
- `--registry` (required): Trusted plugin registry used to resolve the invocation.

## Delegation

Validate the invocation with the shared workflow runtime against the supplied registry.
Report schema failures, unresolved plugins, missing capabilities, selected control modes,
evidence failure behavior, and exact enforcement points without executing plugins.
