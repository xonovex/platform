---
description: Validate and integrate one exact workspace while leaving all cleanup separate
allowed-tools:
  - Read
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target] [--request <file>] [--workspace-provider <provider>]
  [--workspace-revision <revision>] [--destination-provider <provider>]
  [--destination-reference <reference>] [--destination-revision <revision>]
  [--criterion <criterion>...] [--method <method>]
  [--effect <preview|apply>]
---

# /xonovex-workflow:workspace-merge — Merge Workspace

## Arguments

- `target` (required unless `--request` supplies it): Exact workspace path or opaque
  provider-native source reference.
- `--request` (optional): Structured request file for independently bound workspace,
  destination, evidence, and criteria. Do not combine it with the shorthands.
- `--workspace-provider`, `--workspace-revision` (required for provider-native
  shorthand): Provider and exact revision of the source workspace.
- `--destination-provider`, `--destination-reference` (required for shorthand):
  Provider and opaque native destination locator.
- `--destination-revision` (optional): Expected destination revision for concurrency.
- `--criterion` (repeatable, optional): Binding pre-integration condition.
- `--method` (optional): Requested validation and integration procedure.
- `--effect` (optional): `preview` or `apply`; defaults to `preview`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace merge** operation with these arguments. The skill is the source of truth.
Integrate only after validation and exact authorization. Leave the workspace, branch,
reference, and metadata intact; cleanup is a separate operation.
