---
description: Create a new result from an inline or provider-referenced subject without changing the source
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
  [--result <destination-reference>] [--dry-run]
---

# /xonovex-workflow:create — Create

## Goal

- Create one new result from the supplied subject and references.
- Keep every selection independent and load only the capabilities the request needs.
- Return inline content unless persistence is explicitly requested.

## Arguments

- `subject` (required): Inline source content or one opaque provider-native reference.
- `--reference` (repeatable): Supporting inline content or opaque references.
- `--revision` (optional): Exact native revision of a referenced subject.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Explicit properties the new result must satisfy.
- `--result` (optional): Explicit provider-native destination reference.
- `--dry-run` (optional): Preview any persistence or other side effect.

## Core Workflow

1. Resolve explicit selections. Infer kind or provider only when the subject makes one
   choice unambiguous; report the inference and its basis. Stop on ambiguity.
2. Load only explicitly selected or unambiguous domain, method, or provider
   capabilities. If an explicit capability is unavailable, name it and stop.
3. When a provider capability is resolved, let it interpret opaque references and
   revisions.
4. Create a new result that satisfies the criteria without mutating the subject.
5. Return the result inline. Persist only when `--result` or the original request
   explicitly names a destination, then return its native locator and revision when
   supported.

## Implementation

Kind, perspective, method, capability, and provider are separate axes. Invocation
trigger, executor, identity, and agent maturity do not change this operation. Do not
construct a preset, central resolver, registry, or default provider.

## Error Handling

- Missing subject or criteria required by the selected method: stop and name the
  missing input.
- Ambiguous kind or provider: show the viable choices and request one selection.
- Unavailable explicit capability: identify it without substituting another.
- Persistence without a resolvable destination provider: keep the result inline and
  report why it was not persisted.
