---
description: Record an authorized accept, reject, or conditional decision for an exact Experience Design revision
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--authority-reference <reference>] [--provider <selection>]"
---

# /xonovex-workflow:experience-design-accept — Accept Experience Design

## Arguments

- `native-reference` (required): Opaque Experience Design reference.
- `--revision` (required when not provider-implied): Exact revision under decision.
- `--authority-reference` (optional only when provider context proves authority): Actor/role authority reference.
- `--provider` (optional): Decision/result provider.

## Delegation

Load `plan-guide` and perform **experience-design-accept**. Bind the action to a revision
when the selected persistence and decision policy require it. A model may prepare the
brief but must not fabricate the actor, qualification, or decision.
