---
description: Validate an exact deliverable revision against provider-native acceptance criteria using the selected method
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
  - Skill
argument-hint: "<criteria-reference> [--subject <native-reference>] [--revision <native-revision>] [--method <selection>] [--provider <selection>]"
---

# /xonovex-workflow:acceptance-validate — Validate Acceptance Criteria

## Arguments

- `criteria-reference` (required): Opaque native criteria/Formulation reference.
- `--subject` (required): Exact deliverable, branch, pull-request, or artifact reference.
- `--revision` (required when not provider-implied): Exact subject revision.
- `--method` (optional): Criteria method; neutral by default, with BDD/Gherkin selectable when installed.
- `--provider` (optional): Provider for the validation result.

## Delegation

Load `code-review-guide` for structured findings, `testing-guide` for coverage, and
`workflow-guide` for exact-revision/provider-native result contracts. Load an installed
method skill only when the criteria select it (for example, a BDD capability for
Given-When-Then). Report PASS/PARTIAL/FAIL per criterion with evidence and publish through
the selected provider. This advisory validation cannot fabricate accountable human
Acceptance, and an unavailable explicit provider must not fall back to a local file.
