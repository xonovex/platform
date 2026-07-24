---
description: Validate and integrate one exact workspace while leaving cleanup separate
allowed-tools:
  - Read
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target] [--request <file>] [--destination <reference>]
  [--expected-revision <revision>] [--criterion <criterion>...]
  [--method <method>] [--effect <preview|apply>]
---

# /xonovex-workflow:workspace-merge — Merge Workspace

## Arguments

- `target` (required unless `--request` supplies it): Exact workspace path or opaque
  native source reference.
- `--request` (optional): Markdown file containing the workspace, destination, and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--destination` (required unless `--request` supplies it): Exact native destination
  reference.
- `--expected-revision` (optional): Expected destination revision for concurrency.
- `--criterion` (repeatable, optional): Binding pre-integration criterion.
- `--method` (optional): Requested validation and integration procedure.
- `--effect` (optional): `preview` or `apply`; defaults to `preview`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace merge** operation with these arguments. Integrate only after validation
and explicit authority; preserve workspace resources for separate cleanup.
