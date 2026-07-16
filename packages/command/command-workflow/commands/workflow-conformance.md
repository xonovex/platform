---
description: Validate workflow and agent-governance contracts without assuming one provider, harness, policy engine, or configuration format
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Skill
argument-hint: "[subject] [--plane <workflow|governance|both>] [--profile <reference>]"
---

# /xonovex-workflow:workflow-conformance — Validate Semantic Conformance

## Arguments

- `subject` (optional): Result, handle, profile, module, policy, capability matrix, exception, onboarding result, or native reference; defaults to the current context.
- `--plane <workflow|governance|both>` (optional): Contract owner to validate; defaults to auto-detection, then both when the subject crosses planes.
- `--profile <reference>` (optional): Profile whose requirements and guarantees apply.

## Delegation

Load the owning skills and perform their **conformance** operations:

- `workflow-guide` (plugin `xonovex-skill-workflow`) for lifecycle results,
  ephemeral handles, provider handoffs, and workflow profiles.
- `agent-governance-guide` (plugin `xonovex-skill-agent-governance`) for
  executors, events, capability matrices, policies, enforcement, modules, authority,
  exceptions, onboarding, and governance profiles.

The skills own the semantic checks, output, and gotchas. Validate each plane
independently before validating their composition; do not invent a universal persisted
schema or treat skill installation as enforcement evidence.
