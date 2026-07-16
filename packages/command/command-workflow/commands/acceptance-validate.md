---
description: Assemble fresh provider-native evidence for human Acceptance of an exact deliverable revision
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
  - Skill
argument-hint: "<criteria-reference> [--subject <native-reference>] [--revision <native-revision>] [--method <selection>] [--provider <selection>]"
---

# /xonovex-workflow:acceptance-validate — Assemble Acceptance Evidence

## Arguments

- `criteria-reference` (required): Opaque native criteria/Formulation reference.
- `--subject` (required): Exact deliverable, branch, pull-request, or artifact reference.
- `--revision` (required when not provider-implied): Exact subject revision.
- `--method` (optional): Criteria method; neutral by default, with BDD/Gherkin selectable when installed.
- `--provider` (optional): Provider for the validation result.

## Delegation

Load `workflow-guide` and perform **acceptance-validate**. Soft-select installed review,
testing, assessment, criterion-method, and provider skills required by the profile. This
operation assembles advisory evidence and cannot create accountable human Acceptance.
