---
description: Create one isolated workspace from one exact source without merging or cleanup
allowed-tools:
  - Read
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target] [--request <file>] [--workspace-provider <provider>]
  [--source-reference <reference>] [--source-revision <revision>]
  [--branch-reference <reference>] [--effect <preview|apply>]
---

# /xonovex-workflow:workspace-create — Create Workspace

## Arguments

- `target` (required unless `--request` supplies it): Exact workspace path or opaque
  provider-native destination reference.
- `--request` (optional): Structured request file for independently bound source and
  workspace resources. Do not combine it with the shorthands.
- `--workspace-provider` (required for provider-native shorthand): Provider that owns
  workspace creation and target interpretation.
- `--source-reference`, `--source-revision` (required for shorthand): Opaque source
  locator and exact native revision.
- `--branch-reference` (optional): Explicit native branch when the provider uses one.
- `--effect` (optional): `preview` or `apply`; defaults to `preview`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace create** operation with these arguments. The skill is the source of truth.
Create only the named workspace resources; never merge, remove, or publish.
