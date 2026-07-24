---
description: Create one isolated workspace without merging or cleanup
allowed-tools:
  - Read
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target] [--request <file>] [--source <reference>]
  [--source-revision <revision>] [--context <context>...]
  [--idempotency-key <key>]
  [--effect <preview|apply>]
---

# /xonovex-workflow:workspace-create — Create Workspace

## Arguments

- `target` (required unless `--request` supplies it): Exact workspace path or opaque
  native destination reference.
- `--request` (optional): Markdown workflow handoff containing the workspace, source,
  retry identity, and equivalent inputs. Do not combine it with shorthand arguments.
- `--source` (required unless `--request` supplies it): Exact source reference.
- `--source-revision` (optional): Exact source revision; required when the provider
  exposes one.
- `--context` (repeatable, optional): Canonical explanatory context or an opaque
  context reference to preserve with the workspace handoff.
- `--idempotency-key` (optional): Stable retry key. Required for provider-native
  `apply` when the selected provider supports idempotency.
- `--effect` (optional): `preview` or `apply`; defaults to `preview`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace create** operation with these arguments. Create only the named resources;
never merge, remove, or publish.
