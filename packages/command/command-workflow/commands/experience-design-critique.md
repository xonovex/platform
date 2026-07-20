---
description: Independently critique an exact Experience Design revision and publish separate findings
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - Task
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--profile <reference>] [--provider <selection>] [--method <selection>]"
---

# /xonovex-workflow:experience-design-critique — Critique Experience Design

## Arguments

- `native-reference` (required): Opaque Experience Design reference.
- `--revision` (required when not provider-implied): Exact revision to critique.
- `--profile`, `--provider`, `--method` (optional): Review requirements, publication provider, and critique method.

## Delegation

Load `plan-guide` and perform **experience-design-critique** in a fresh independent
context. Publish separate findings; do not revise or accept the design.
