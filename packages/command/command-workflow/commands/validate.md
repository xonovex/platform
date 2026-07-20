---
description: Validate an exact subject against explicit criteria and return evidence without changing the subject
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  <subject> --criteria <criteria>... [--reference <reference>...]
  [--revision <revision>] [--kind <selection>] [--perspective <selection>]
  [--method <selection>] [--capability <selection>...]
  [--provider <selection>] [--result <destination-reference>]
---

# /xonovex-workflow:validate — Validate

## Goal

- Evaluate one exact subject against explicit criteria.
- Produce reproducible evidence and a result for every criterion.
- Remain read-only with respect to the subject.

## Arguments

- `subject` (required): Inline content or one opaque provider-native reference.
- `--criteria` (repeatable, required): Inline criterion or opaque criteria reference.
- `--reference` (repeatable): Supporting evidence or opaque references.
- `--revision` (optional): Exact native revision; required when provider context does
  not otherwise pin a referenced subject.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--result` (optional): Explicit provider-native evidence destination.

## Core Workflow

1. Resolve explicit selections. Infer kind or provider only from unambiguous subject
   evidence, report the inference and basis, and stop on ambiguity.
2. Load only explicitly selected or unambiguous validation, domain, method, and
   provider capabilities. Name and stop on an unavailable explicit capability.
3. When a provider capability is resolved, let it interpret opaque references and
   revisions.
4. Evaluate each criterion against the exact subject and report pass, fail, or blocked
   with reproducible evidence and freshness.
5. Return evidence inline. Persist only to an explicit `--result`, returning its native
   locator and revision when supported.

## Implementation

Validation does not revise, accept, publish, or otherwise mutate the subject. Kind,
perspective, method, capability, and provider are independent; invocation trigger,
executor, identity, and agent maturity do not change the validation contract.

## Error Handling

- Missing criteria or unpinned mutable subject: stop before evaluating.
- Criterion cannot be evaluated: mark it blocked and state the missing evidence.
- Ambiguous kind or provider: report the selections instead of guessing.
- Unavailable explicit capability: identify it without falling back.
