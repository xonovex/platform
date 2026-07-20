---
description: Publish exact-window operational observations from provider-native monitoring, user, security, AI, cost, accessibility, and delivery evidence
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - Task
  - Skill
argument-hint: "<subject-reference> [--revision <native-revision>] --window <range> [--baseline <reference>] [--profile <reference>] [--provider <selection>]"
---

# /xonovex-workflow:observe-run — Observe an Operational Subject

## Arguments

- `subject-reference` (required): Opaque released/deployed subject or observable resource.
- `--revision` (optional): Exact subject revision or resource version.
- `--window` (required): Observation time window.
- `--baseline`, `--profile`, `--provider` (optional): Baseline, requirements, and result provider.

## Delegation

Load `workflow-guide` and perform **observe-run**. Soft-select monitoring, user-feedback,
security, AI, cost, accessibility, delivery-outcome, and provider skills. Keep bounded
model/agent summaries advisory and linked to provider-native source evidence. For an
operator-hosted `AgentRun`, use its phase, conditions, Job/Pod references, and any
explicitly selected external evidence sink. Do not infer workflow controls or maturity
from operator status, and do not copy prompt or Secret content into the observation.
