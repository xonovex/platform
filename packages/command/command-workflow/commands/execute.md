---
description: Execute one bounded subject and report its observable result without implicit publication or workspace changes
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  <subject> [--reference <reference>...] [--revision <revision>]
  [--kind <selection>] [--perspective <selection>] [--criteria <criteria>...]
  [--method <selection>] [--capability <selection>...] [--provider <selection>]
  [--result <destination-reference>] [--dry-run]
---

# /xonovex-workflow:execute — Execute

## Goal

- Carry out one bounded subject using only selected or unambiguous capabilities.
- Preserve explicit scope and surface partial progress or failure.
- Report the observable result without implicitly publishing or managing a workspace.

## Arguments

- `subject` (required): Inline instructions or one opaque provider-native reference.
- `--reference` (repeatable): Supporting inputs or opaque references.
- `--revision` (optional): Exact native revision of a referenced subject.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Explicit completion and verification criteria.
- `--result` (optional): Explicit provider-native result destination.
- `--dry-run` (optional): Preview actions and side effects without applying them.

## Core Workflow

1. Bound the requested work and resolve explicit selections. Infer kind or provider
   only when unambiguous, report the inference, and stop on ambiguity.
2. Load only explicitly selected or subject-unambiguous domain, method, execution, and
   provider capabilities. Name and stop on an unavailable explicit capability.
3. When a provider capability is resolved, let it interpret opaque references and
   revisions.
4. Execute the bounded work, verifying criteria in proportion to risk and preserving
   usable partial results when a later action fails.
5. Return changes, evidence, failures, and remaining work inline. Persist only to an
   explicit `--result`, returning its native locator and revision when supported.

## Implementation

Execution never implicitly creates, merges, abandons, or cleans a workspace and never
implicitly publishes. Kind, perspective, method, capability, and provider remain
orthogonal; trigger, executor, identity, and agent maturity do not select behavior.

## Error Handling

- Unbounded subject or criteria required by the selected method: stop and name what
  must be supplied.
- Ambiguous kind or provider: report the viable choices without guessing.
- Unavailable explicit capability: identify it and stop without fallback.
- Partial failure: preserve completed work and report the exact retry boundary.
