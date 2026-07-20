---
description: Publish an exact subject to an explicit provider destination and return its native locator
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
  <subject> --result <destination-reference> [--reference <reference>...]
  [--revision <revision>] [--kind <selection>] [--perspective <selection>]
  [--criteria <criteria>...] [--method <selection>]
  [--capability <selection>...] [--provider <selection>]
  [--confirm] [--dry-run]
---

# /xonovex-workflow:publish — Publish

## Goal

- Publish one exact subject to one explicit destination.
- Keep provider behavior behind the selected provider capability.
- Return the provider-native locator and revision without changing the source.

## Arguments

- `subject` (required): Inline content or one opaque provider-native reference.
- `--result` (required): Explicit provider-native destination reference.
- `--reference` (repeatable): Supporting inputs or opaque references.
- `--revision` (optional): Exact native source revision; required when provider
  context does not otherwise pin it.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Preconditions the publication must satisfy.
- `--confirm` (optional): Explicitly authorize the described publication effect.
- `--dry-run` (optional): Resolve and preview the publication without applying it.

## Core Workflow

1. Pin the subject revision and destination. Infer kind or provider only when
   unambiguous, report the inference and basis, and stop on ambiguity.
2. Load only explicitly selected or unambiguous publication, domain, method, and
   provider capabilities. Never substitute for an unavailable explicit capability.
3. Let the resolved provider capability interpret all opaque references and own native
   authentication, identifiers, revisions, idempotency, and publication effects.
4. Preview the exact effect. Apply it only when `--confirm` or the original request
   already explicitly authorizes that subject, destination, and effect.
5. Return the provider-native result locator and revision when supported.

## Implementation

Publication never revises the source or implies a gate decision. Kind, perspective,
method, capability, and provider stay orthogonal; invocation trigger, executor,
identity, and agent maturity do not select provider behavior.

## Error Handling

- Missing destination or mutable source revision: stop before publishing.
- Ambiguous provider or destination: report the viable choices without guessing.
- Unavailable explicit capability: name it and stop without fallback.
- Missing authorization: return the dry-run preview and request confirmation.
