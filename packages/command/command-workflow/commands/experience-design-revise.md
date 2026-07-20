---
description: Revise an exact Experience Design revision from explicit feedback and critique references
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--feedback <reference>...] [--provider <selection>]"
---

# /xonovex-workflow:experience-design-revise — Revise Experience Design

## Arguments

- `native-reference` (required): Opaque Experience Design reference.
- `--revision` (required when the provider does not imply it): Exact subject revision.
- `--feedback` (repeatable): Opaque feedback, critique, validation, or decision reference.
- `--provider` (optional): Explicit result-provider selection.

## Delegation

Load `plan-guide` and perform **experience-design-revise**. Resolve every feedback item,
preserve supersession when supported, and never silently accept the subject.
