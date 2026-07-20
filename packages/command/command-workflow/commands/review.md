---
description: Review an exact subject against explicit criteria without changing it
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  <subject> [--reference <reference>...] [--revision <revision>]
  [--kind <selection>] [--perspective <selection>] [--criteria <criteria>...]
  [--method <selection>] [--capability <selection>...] [--provider <selection>]
  [--result <destination-reference>]
---

# /xonovex-workflow:review — Review

## Goal

- Evaluate one exact subject without modifying it.
- Produce evidence-linked findings from an explicit perspective and criteria.
- Return inline findings unless persistence is explicitly requested.

## Arguments

- `subject` (required): Inline content or one opaque provider-native reference.
- `--reference` (repeatable): Supporting evidence or opaque references.
- `--revision` (optional): Exact native revision; required when provider context does
  not otherwise pin a referenced subject.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, conditionally required): Explicit review criteria. A
  selected method may supply them only when that contract is unambiguous.
- `--result` (optional): Explicit provider-native destination reference.

## Core Workflow

1. Resolve explicit selections. Infer kind or provider only from unambiguous subject
   evidence, report the inference and basis, and stop when more than one choice fits.
2. Load only explicitly selected or unambiguous review, domain, method, and provider
   capabilities. Stop and name any unavailable explicit capability.
3. When a provider capability is resolved, let it read opaque references and revisions.
4. Review the exact subject against the criteria and report findings, evidence,
   severity when relevant, and uncertainty. Do not revise the subject.
5. Return findings inline. Persist only to an explicitly requested `--result`, then
   return its native locator and revision when supported.

## Implementation

Kind, perspective, method, capability, and provider remain independent. Invocation
trigger, executor, identity, and agent maturity never select review behavior. Do not
invent a preset, central resolver, registry, or default provider.

## Error Handling

- Missing or mutable referenced revision: stop before reviewing the wrong subject.
- Missing review standard: request criteria or an unambiguous method.
- Ambiguous kind or provider: report the choices and request one.
- Unavailable explicit capability: identify it without falling back.
