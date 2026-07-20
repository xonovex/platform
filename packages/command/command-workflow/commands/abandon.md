---
description: Stop work on one subject while preserving its reason, partial result, and retry boundary
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
  <subject> --reason <text> [--reference <reference>...]
  [--revision <revision>] [--kind <selection>] [--perspective <selection>]
  [--criteria <criteria>...] [--method <selection>]
  [--capability <selection>...] [--provider <selection>]
  [--result <destination-reference>] [--cleanup <selection>] [--dry-run]
---

# /xonovex-workflow:abandon — Abandon

## Goal

- Stop work on one subject and record why.
- Preserve partial results, evidence, cleanup state, and a clear retry boundary.
- Avoid deleting or rewriting the subject.

## Arguments

- `subject` (required): Inline work description or one opaque provider-native
  reference.
- `--reason` (required): Present-tense reason for stopping.
- `--reference` (repeatable): Partial results, evidence, or opaque references.
- `--revision` (optional): Exact native revision being abandoned.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Retention or cleanup constraints.
- `--result` (optional): Explicit destination for the abandonment record.
- `--cleanup` (optional): Explicit cleanup capability or scope; preserve by default.
- `--dry-run` (optional): Preview provider updates or cleanup.

## Core Workflow

1. Resolve the exact subject and explicit selections. Report any unambiguous kind or
   provider inference and stop rather than guess when inference is ambiguous.
2. Load only explicitly selected or subject-unambiguous domain, abandonment, cleanup,
   and provider capabilities. Name and stop on an unavailable explicit capability.
3. When a provider capability is resolved, let it interpret opaque references and
   revisions.
4. Record the reason, partial result, evidence, cleanup state, and retry boundary.
   Preserve the source and apply cleanup only when explicitly selected and authorized.
5. Return the record inline or persist it only to an explicit `--result`, returning a
   native locator and revision when supported.

## Implementation

Abandonment is descriptive unless a selected provider capability performs an explicitly
authorized state update. Kind, perspective, method, capability, and provider remain
independent; trigger, executor, identity, and agent maturity never select behavior.

## Error Handling

- Missing subject or reason: stop and request the missing input.
- Ambiguous kind or provider: report the viable choices without guessing.
- Unavailable explicit capability: identify it without substitution.
- Destructive cleanup without exact scope and authorization: preserve state and return
  a dry-run preview.
