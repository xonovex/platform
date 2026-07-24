---
description: Preview and remove only explicitly named workspace resources
allowed-tools:
  - Read
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target...] [--request <file>] [--target-revision <revision>...]
  [--remove-reference] [--force] [--idempotency-key <key>]
  [--effect <preview|apply>]
---

# /xonovex-workflow:workspace-cleanup — Clean Up Workspaces

## Arguments

- `target` (repeatable, required unless `--request` supplies it): Exact workspace path
  or opaque native reference.
- `--request` (optional): Markdown workflow handoff containing exact cleanup targets,
  revisions, recovery information, and equivalent inputs. Do not combine it with
  shorthand arguments.
- `--target-revision` (repeatable, optional): Exact native revision corresponding to
  each target when its provider exposes one.
- `--remove-reference` (optional): Include each target's exact associated reference.
- `--force` (optional): Include an exact dirty or unmerged target in the preview.
- `--idempotency-key` (optional): Stable retry key. Required for provider-native
  `apply` when the selected provider supports idempotency.
- `--effect` (optional): `preview` or `apply`; defaults to `preview`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace cleanup** operation with these arguments. Apply only the exact previewed
set and report recovery information for every effect.
