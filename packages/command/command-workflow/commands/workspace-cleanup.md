---
description: Preview and remove only explicitly named workspace resources
allowed-tools:
  - Read
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target...] [--request <file>] [--remove-reference] [--force]
  [--effect <preview|apply>]
---

# /xonovex-workflow:workspace-cleanup — Clean Up Workspaces

## Arguments

- `target` (repeatable, required unless `--request` supplies it): Exact workspace path
  or opaque native reference.
- `--request` (optional): Markdown file containing exact cleanup targets and equivalent
  inputs. Do not combine it with shorthand arguments.
- `--remove-reference` (optional): Include each target's exact associated reference.
- `--force` (optional): Include an exact dirty or unmerged target in the preview.
- `--effect` (optional): `preview` or `apply`; defaults to `preview`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace cleanup** operation with these arguments. Apply only the exact previewed
set and report recovery information for every effect.
