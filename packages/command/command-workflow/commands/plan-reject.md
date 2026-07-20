---
description: Record an authorized rejection against an exact Planning revision without deleting it
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<native-reference> [--revision <native-revision>] [--authority-reference <reference>] [--reason <text>] [--provider <selection>]"
---

# /xonovex-workflow:plan-reject — Reject Planning Result

## Arguments

- `native-reference` (required): Opaque Planning reference.
- `--revision` (required when not provider-implied): Exact revision under rejection.
- `--authority-reference` (optional only when provider context proves authority): Actor/role reference.
- `--reason` (required): Rejection rationale or required changes.
- `--provider` (optional): Status/decision provider.

## Delegation

Load `plan-guide` and perform **plan-reject**. Record the rejection against the requested
subject, preserve history when supported, and stop.
