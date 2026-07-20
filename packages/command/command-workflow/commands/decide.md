---
description: Record one descriptive outcome and rationale without granting authority or changing a gate
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  <subject> [--outcome <text>] [--reference <reference>...]
  [--revision <revision>] [--kind <selection>] [--perspective <selection>]
  [--criteria <criteria>...] [--method <selection>]
  [--capability <selection>...] [--provider <selection>]
  [--result <destination-reference>]
---

# /xonovex-workflow:decide — Decide

## Goal

- Resolve one question into a stated outcome and rationale.
- Keep evidence, recommendation, and the recorded outcome distinguishable.
- Describe the decision without granting authority or changing operational state.

## Arguments

- `subject` (required): Inline decision question or one opaque provider-native
  reference.
- `--outcome` (optional): Explicit outcome to record; otherwise derive a recommendation
  from the supplied evidence and criteria.
- `--reference` (repeatable): Evidence, options, constraints, or opaque references.
- `--revision` (optional): Exact native revision of a referenced subject.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Explicit decision criteria.
- `--result` (optional): Explicit provider-native destination reference.

## Core Workflow

1. Resolve explicit selections. Infer kind or provider only when unambiguous, report
   the inference and basis, and stop when multiple choices remain.
2. Load only selected or unambiguous decision, domain, method, and provider
   capabilities. Name and stop on an unavailable explicit capability.
3. When a provider capability is resolved, let it interpret opaque references and
   revisions.
4. Compare the evidence against the criteria, then record the outcome, rationale,
   alternatives, assumptions, and uncertainty.
5. Return the result inline or persist it only to an explicit `--result`, returning a
   native locator and revision when supported.

## Implementation

The result is descriptive: it never approves, rejects, authorizes, promotes, or changes
a gate. Kind, perspective, method, capability, and provider stay independent, and
invocation trigger, executor, identity, and agent maturity do not affect the semantics.

## Error Handling

- Unclear decision question or materially missing evidence: stop and identify the gap.
- Ambiguous kind or provider: present the viable choices instead of guessing.
- Unavailable explicit capability: name it without substituting another.
- Requested authority effect: return the descriptive decision and state that the
  separate authority action was not performed.
