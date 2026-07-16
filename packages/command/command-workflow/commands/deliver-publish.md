---
description: Publish exact Development revisions as a provider-native reviewable candidate with an immutable revision
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<development-reference...> --target <native-target> [--revision <native-revision>] [--provider <selection>] [--idempotency-key <key>]"
---

# /xonovex-workflow:deliver-publish — Publish a Deliverable

## Arguments

- `development-reference...` (required): Opaque source Development references.
- `--revision` (repeatable): Exact native source revisions.
- `--target` (required): Provider-native intended target.
- `--provider` (optional): Explicit local, hosted, repository, database, or other provider selection.
- `--idempotency-key` (optional): Stable publication key when the provider supports safe retries.

## Delegation

Load `workflow-guide` and perform **deliver-publish**. Load the selected provider skill
for native publication, authentication, identifiers, revisions, and side effects. Return
a Deliverable Publication result; never fall back from an unavailable explicit provider.
