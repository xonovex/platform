---
description: Independently critique an exact Decision revision without acting as its authority
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--profile <reference>] [--provider <selection>] [--mode <mode>]"
---

# /xonovex-workflow:decision-critique — Critique Decision

## Arguments

- `native-reference` (required): Opaque Decision reference.
- `--revision` (required when not provider-implied): Exact revision to critique.
- `--profile`, `--provider`, `--mode` (optional): Applicable requirements, publication provider, and critique lens.

## Delegation

Load `plan-guide` and perform **decision-critique** in fresh independent context.
Challenge evidence, options, consequences, and authority; do not make or revise the
decision.
