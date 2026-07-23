---
description: Preview and remove only explicitly bound workspace resources with recovery reporting
allowed-tools:
  - Read
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target...] [--request <file>] [--workspace-provider <provider>]
  [--remove-reference] [--force] [--effect <preview|apply>]
---

# /xonovex-workflow:workspace-cleanup — Clean Up Workspaces

## Arguments

- `target` (repeatable, required unless `--request` supplies it): Exact workspace
  paths or opaque native references. Use a request file when targets have different
  providers or revisions.
- `--request` (optional): Structured request file for independently bound cleanup
  targets. Do not combine it with the shorthands.
- `--workspace-provider` (required for provider-native shorthand): Provider that owns
  every shorthand target.
- `--remove-reference` (optional): Include each target's exact associated native
  reference in the preview.
- `--force` (optional): Include an exact dirty or unmerged target in the preview.
- `--effect` (optional): `preview` or `apply`; defaults to `preview`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace cleanup** operation with these arguments. The skill is the source of
truth. This is the only workspace operation that removes workspace resources; apply
only the exact previewed set and report recovery for every effect.
