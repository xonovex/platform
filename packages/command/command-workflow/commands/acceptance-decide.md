---
description: Record an accountable human Acceptance decision bound to an exact deliverable, target, evidence set, policy version, and expiry
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "<deliverable-reference> [--revision <native-revision>] --target <native-target> --evidence <reference> [--policy <reference>] [--expires <timestamp>] [--provider <selection>]"
---

# /xonovex-workflow:acceptance-decide — Record Human Acceptance

## Arguments

- `deliverable-reference` (required): Opaque Deliverable Publication reference.
- `--revision` (optional): Exact immutable deliverable revision.
- `--target` (required): Exact intended Integration target.
- `--evidence` (repeatable): Fresh provider-native acceptance evidence references.
- `--policy`, `--expires`, `--provider` (optional): Policy binding, decision expiry, and provider selection.

## Delegation

Perform **acceptance-decide** with the selected identity and decision-provider skills.
Never accept an agent, model, script, copied label, or ordinary tool call as accountable
human sign-off.
