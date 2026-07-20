---
description: Produce a traceable revision from explicit feedback without overwriting the source
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
  <subject> --feedback <feedback>... [--reference <reference>...]
  [--revision <revision>] [--kind <selection>] [--perspective <selection>]
  [--criteria <criteria>...] [--method <selection>]
  [--capability <selection>...] [--provider <selection>]
  [--result <destination-reference>] [--dry-run]
---

# /xonovex-workflow:revise — Revise

## Goal

- Produce one new revision from explicit feedback.
- Preserve the source, feedback provenance, and unresolved items.
- Return the revision inline unless persistence is explicitly requested.

## Arguments

- `subject` (required): Inline content or one opaque provider-native reference.
- `--feedback` (repeatable, required): Inline feedback or opaque feedback reference.
- `--reference` (repeatable): Supporting inline content or opaque references.
- `--revision` (optional): Exact native source revision; required when provider
  context does not otherwise pin it.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Constraints the revision must retain.
- `--result` (optional): Explicit provider-native destination reference.
- `--dry-run` (optional): Preview persistence or source-adjacent edits.

## Core Workflow

1. Resolve explicit selections. Report any unambiguous kind or provider inference and
   stop rather than guess when inference is ambiguous.
2. Load only explicitly selected or unambiguous revision, domain, method, or provider
   capabilities. Name and stop on an unavailable explicit capability.
3. When a provider capability is resolved, let it interpret opaque subjects, feedback,
   references, and revisions.
4. Apply the feedback to a new result. Record applied, deferred, and conflicting
   feedback; never overwrite or rewrite the source's history.
5. Return the revised result inline. Persist only to an explicit `--result`, returning
   its native locator and revision when supported.

## Implementation

Kind, perspective, method, capability, and provider are orthogonal selections.
Invocation trigger, executor, identity, and agent maturity do not alter revision
semantics. Do not use a preset, central resolver, registry, or implicit provider.

## Error Handling

- Missing feedback or unpinned mutable source: stop and request the exact input.
- Conflicting feedback: report the conflict and request a decision; do not guess.
- Ambiguous kind or provider: report the viable selections.
- Unavailable explicit capability: identify it without substitution.
