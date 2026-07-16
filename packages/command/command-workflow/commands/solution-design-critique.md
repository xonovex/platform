---
description: Independently critique an exact Solution Design revision and publish separate severity-ranked findings
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

# /xonovex-workflow:solution-design-critique — Critique Solution Design

## Arguments

- `native-reference` (required): Opaque Solution Design reference.
- `--revision` (required when not provider-implied): Exact revision to critique.
- `--profile`, `--provider`, `--method` (optional): Review requirements, result provider, and critique capabilities.

## Delegation

Load `plan-guide` and perform **solution-design-critique** in fresh independent context;
load `workflow-guide` for exact-revision contracts. Publish separate findings and do not
revise the subject or claim architecture/security/privacy authority.
